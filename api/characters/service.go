package characters

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
)

// Service represents a service implementation for the "characters" resource
type Service api.Service

func strInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (s *Service) convertDTOFields(fields dtos.ReqCharacterFields) *models.CharacterFields {
	groups := make(map[string]*models.CharacterFieldGroup)
	for k, v := range fields.Groups {
		groups[k] = &models.CharacterFieldGroup{
			Fields: make(map[string]*models.CharacterField),
			Hidden: v.Hidden,
		}
		for k2, v2 := range v.Fields {
			groups[k].Fields[k2] = &models.CharacterField{
				Value:  v2.Value,
				Hidden: v2.Hidden,
				Type:   v2.Type,
			}
		}
	}
	return &models.CharacterFields{
		Groups: groups,
	}
}

// New creates a new character
func (s *Service) New(data dtos.ReqCreateCharacter) *models.Character {
	return &models.Character{
		ID:     ksuid.New().String(),
		Name:   data.Name,
		Tag:    data.Tag,
		Fields: s.convertDTOFields(*data.Fields),
		Meta: &models.CharacterMeta{
			Archived: data.Meta.Archived,
			Hidden:   data.Meta.Hidden,
		},
	}
}

// FindByID returns a character by their ID
func (s *Service) FindByID(id string) (*models.Character, error) {
	var character models.Character
	if err := s.Providers.DB.Get(&character, "SELECT * FROM characters WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &character, nil
}

// FindByUniverse returns a collection of character references associated with a universe
func (s *Service) FindByUniverse(
	universe *models.Universe,
	ctx dtos.CharacterQuery,
) (*[]models.CharacterReference, int, error) {
	var (
		count      = 0
		characters = make([]models.CharacterReference, 0)
		query      = "SELECT id, name, tag, owner_id, created_at, updated_at FROM characters"
		cquery     = "SELECT count(*) FROM characters"
	)
	if ctx.Collaborator.Role != models.CollaboratorMember {
		qs := "WHERE universe_id = $1"
		query = fmt.Sprintf("%v %v  LIMIT $2 OFFSET $3", query, qs)
		cquery = fmt.Sprintf("%v %v", cquery, qs)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&characters,
			query,
			universe.ID,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}

		// Retrieve total count
		if err := s.Providers.DB.Get(&count, cquery, universe.ID); err != nil {
			return nil, 0, err
		}
	} else {
		qs := "WHERE universe_id = $1 AND (meta->'hidden'='false' OR owner_id=$2)"
		query = fmt.Sprintf("%v %v LIMIT $3 OFFSET $4", query, qs)
		cquery = fmt.Sprintf("%v %v", cquery, qs)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&characters,
			query,
			universe.ID,
			ctx.Collaborator.UserID,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}

		// Retrieve total count
		if err := s.Providers.DB.Get(&count, cquery, universe.ID, ctx.Collaborator.UserID); err != nil {
			return nil, 0, err
		}
	}
	return &characters, count, nil
}

// Create saves a new character to the database
func (s *Service) Create(
	universe *models.Universe,
	character *models.Character,
	owner *models.User,
) (*models.Character, error) {
	var c models.Character
	if err := s.Providers.DB.Get(
		&c,
		`INSERT INTO characters (id, universe_id, owner_id, name, tag, fields, meta) VALUES
		($1, $2, $3, $4, $5, $6, $7) RETURNING *`,
		character.ID,
		universe.ID,
		owner.ID,
		character.Name,
		character.Tag,
		character.Fields,
		character.Meta,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

// Update updates an existing character in the database
func (s *Service) Update(character *models.Character) (*models.Character, error) {
	var c models.Character
	character.UpdatedAt = time.Now()
	rows, err := s.Providers.DB.NamedQuery(
		`UPDATE characters SET name = :name, tag = :tag, fields = :fields, meta = :meta,
		updated_at = :updated_at WHERE id = :id RETURNING *`,
		character,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Validate validates a character according to a universe's guide, and fixes auto-fixable
// errors if possible and specified
func (s *Service) Validate(character *models.Character, universe *models.Universe) error {
	for _, group := range *universe.Guide.Groups {
		if cGroup, ok := character.Fields.Groups[group.Name]; ok {
			for _, field := range *group.Fields {
				if cField, ok := cGroup.Fields[field.Name]; ok {
					if cField.Type != field.Type {
						return api.ErrBadBody(
							fmt.Sprintf(
								"Field '%s' in group '%s' must specify type '%s'",
								field.Name,
								group.Name,
								field.Type,
							),
						)
					}
					switch field.Type {
					case models.GuideFieldText:
						v, ok := cField.Value.(string)
						if !ok {
							return api.ErrBadBody(
								fmt.Sprintf("Field '%s' in group '%s' must be a string", field.Name, group.Name),
							)
						}
						meta, _ := field.Meta.(models.UniverseGuideMetaText)
						if ok, _ := regexp.Match(meta.Pattern, []byte(v)); !ok {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must match pattern %s",
									field.Name,
									group.Name,
									meta.Pattern,
								),
							)
						}
						if len(v) < meta.MinLength || len(v) > meta.MaxLength {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must be in range of %d and %d",
									field.Name,
									group.Name,
									meta.MinLength,
									meta.MaxLength,
								),
							)
						}
						cField.Value = strings.TrimSpace(v)
					case models.GuideFieldDescription:
						v, ok := cField.Value.(string)
						if !ok {
							return api.ErrBadBody(
								fmt.Sprintf("Field '%s' in group '%s' must be a string", field.Name, group.Name),
							)
						}
						meta, _ := field.Meta.(models.UniverseGuideMetaDescription)
						if len(v) < meta.MinLength || len(v) > meta.MaxLength {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must be in range of %d and %d",
									field.Name,
									group.Name,
									meta.MinLength,
									meta.MaxLength,
								),
							)
						}
						cField.Value = strings.TrimSpace(v)
					case models.GuideFieldNumber:
						meta, _ := field.Meta.(models.UniverseGuideMetaNumber)
						if meta.Float {
							v, ok := cField.Value.(float64)
							if !ok {
								return api.ErrBadBody(
									fmt.Sprintf("Field '%s' in group '%s' must be a float", field.Name, group.Name),
								)
							}
							if v < meta.Min || v > meta.Max {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be in range of %f and %f",
										field.Name,
										group.Name,
										meta.Min,
										meta.Max,
									),
								)
							}
							if meta.Tick != 0 && math.Mod(float64(v), float64(meta.Tick)) != 0 {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be divisible by %f",
										field.Name,
										group.Name,
										meta.Tick,
									),
								)
							}
						} else {
							v, ok := cField.Value.(float64)
							if !ok || (ok && v != float64(int64(v))) {
								return api.ErrBadBody(
									fmt.Sprintf("Field '%s' in group '%s' must be an integer", field.Name, group.Name),
								)
							}
							if int(v) < int(meta.Min) || int(v) > int(meta.Max) {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be in range of %d and %d",
										field.Name,
										group.Name,
										int(meta.Min),
										int(meta.Max),
									),
								)
							}
							if meta.Tick != 0 && math.Mod(v, meta.Tick) != 0 {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be divisible by %d",
										field.Name,
										group.Name,
										int(meta.Tick),
									),
								)
							}
						}
					case models.GuideFieldProgress:
						v, ok := cField.Value.(float64)
						if !ok {
							return api.ErrBadBody(
								fmt.Sprintf("Field '%s' in group '%s' must be a float", field.Name, group.Name),
							)
						}
						meta, _ := field.Meta.(models.UniverseGuideMetaProgress)
						if v < float64(meta.Min) || v > float64(meta.Max) {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must be in range of %f and %f",
									field.Name,
									group.Name,
									meta.Min,
									meta.Max,
								),
							)
						}
						if meta.Tick != 0 && math.Mod(v, meta.Tick) != 0 {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must be divisible by %f",
									field.Name,
									group.Name,
									meta.Tick,
								),
							)
						}
					case models.GuideFieldOptions:
						meta, _ := field.Meta.(models.UniverseGuideMetaOptions)
						if meta.Multiple {
							v, ok := cField.Value.([]interface{})
							if !ok {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be a string list",
										field.Name,
										group.Name,
									),
								)
							}
							l := make([]string, 0)
							for _, v2 := range v {
								v2, ok := v2.(string)
								if !ok {
									return api.ErrBadBody(
										fmt.Sprintf(
											"Field '%s' in group '%s' must be a string list",
											field.Name,
											group.Name,
										),
									)
								}
								if !strInSlice(v2, meta.Options) {
									return api.ErrBadBody(
										fmt.Sprintf(
											"Field '%s' in group '%s' must all be one of %v",
											field.Name,
											group.Name,
											meta.Options,
										),
									)
								}
								l = append(l, strings.TrimSpace(v2))
							}
							cField.Value = l
						} else {
							v, ok := cField.Value.(string)
							if !ok {
								return api.ErrBadBody(
									fmt.Sprintf("Field '%s' in group '%s' must be a string", field.Name, group.Name),
								)
							}
							cField.Value = strings.TrimSpace(v)
							if !strInSlice(v, meta.Options) {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be one of %v",
										field.Name,
										group.Name,
										meta.Options,
									),
								)
							}
						}
					case models.GuideFieldList:
						v, ok := cField.Value.([]interface{})
						if !ok {
							return api.ErrBadBody(
								fmt.Sprintf("Field '%s' in group '%s' must be a string list", field.Name, group.Name),
							)
						}
						l := make([]string, 0)
						for _, v2 := range v {
							vs, ok := v2.(string)
							if !ok {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' must be a string list",
										field.Name,
										group.Name,
									),
								)
							}
							if strings.TrimSpace(vs) == "" {
								return api.ErrBadBody(
									fmt.Sprintf(
										"Field '%s' in group '%s' may not contain empty strings",
										field.Name,
										group.Name,
									),
								)
							}
							l = append(l, strings.TrimSpace(vs))
						}
						cField.Value = l
						meta, _ := field.Meta.(models.UniverseGuideMetaList)
						if len(v) < meta.MinElements || len(v) > meta.MaxElements {
							return api.ErrBadBody(
								fmt.Sprintf(
									"Field '%s' in group '%s' must contain number of elements in range of %d and %d",
									field.Name,
									group.Name,
									meta.MinElements,
									meta.MaxElements,
								),
							)
						}
					}
				} else {
					if field.Required {
						return api.ErrBadBody(
							fmt.Sprintf("Field '%s' in group '%s' is required", field.Name, group.Name),
						)
					}
				}
			}
		} else {
			if group.Required {
				return api.ErrBadBody(
					fmt.Sprintf("Group '%s' is required", group.Name),
				)
			}
		}
	}

	// Ensure no forbidden groups or fields are sent over
	for i, group := range character.Fields.Groups {
		var hasGroup *models.UniverseGuideGroup
		for _, uGroup := range *universe.Guide.Groups {
			if uGroup.Name == i {
				hasGroup = &uGroup
			}
		}
		if hasGroup != nil {
			for k := range group.Fields {
				var hasField bool
				for _, uField := range *hasGroup.Fields {
					if uField.Name == k {
						hasField = true
					}
				}
				if !hasField {
					return api.ErrBadBody(
						fmt.Sprintf("Guide does not document provided field '%s' in group '%s'", k, hasGroup.Name),
					)
				}
			}
		} else {
			return api.ErrBadBody(
				fmt.Sprintf("Guide does not document provided group '%s'", i),
			)
		}
	}
	return nil
}

// Delete deletes a character
func (s *Service) Delete(character *models.Character) error {
	if _, err := s.Providers.DB.Exec(`DELETE FROM characters WHERE id=$1`, character.ID); err != nil {
		return err
	}
	return nil
}
