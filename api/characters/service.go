package characters

import (
	"bytes"
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// AvatarSize represents the width and height dimensions for character avatars
const AvatarSize = 512

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
		ID:     s.Providers.ShortID.MustGenerate(),
		Name:   data.Name,
		Tag:    data.Tag,
		Fields: s.convertDTOFields(*data.Fields),
		Meta: &models.CharacterMeta{
			Hidden:     data.Meta.Hidden,
			NameHidden: data.Meta.NameHidden,
			Name:       &data.Meta.Name,
		},
	}
}

// FindCharacterImages returns images associated with a character
func (s *Service) FindCharacterImages(id string) (models.CharacterImages, error) {
	images := make(models.CharacterImages)

	rows, err := s.Providers.DB.Queryx(`SELECT key, url FROM character_images WHERE character_id=$1`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key, url string
		if err := rows.Scan(&key, &url); err != nil {
			return nil, err
		}
		images[key] = url
	}
	return images, nil
}

// FindByID returns a character by their ID
func (s *Service) FindByID(id string) (*models.Character, error) {
	var character models.Character
	if err := s.Providers.DB.Get(&character, QueryFindByID, id); err != nil {
		return nil, err
	}
	images, err := s.FindCharacterImages(id)
	if err != nil {
		return nil, err
	}
	character.Images = images

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
		query      = ""
		/* query      = `SELECT id, name, tag, owner_id, created_at, updated_at, character_images.url AS avatar_url,
		meta->'hidden' AS hidden FROM characters`
		cquery = "SELECT count(*) FROM characters"*/
	)

	// Normalize the search query
	if ctx.Query == "" {
		query = "%"
	} else {
		query = fmt.Sprintf("%%%v%%", strings.Replace(ctx.Query, " ", "%", -1))
	}

	// Create the search query
	gensql := s.Providers.SQLBuilder.Select(`id, name, tag, owner_id, created_at, updated_at, character_images.url
	AS avatar_url, (meta->>'hidden')::boolean AS hidden, CASE WHEN meta->>'nameHidden' IS NULL THEN false ELSE
	(meta->>'nameHidden')::boolean END AS name_hidden, meta->'name' AS parsed_name`).From(`characters`).LeftJoin(`
	character_images ON character_images.character_id = characters.id`).Where(`universe_id = ? AND name ILIKE ?`,
		universe.ID, query)

	// Factor whether all characters should be included into the query
	if ctx.Collaborator.Role != models.CollaboratorMember {
		// Factor whether hidden characters should be included or not
		if !ctx.IncludeHidden {
			gensql = gensql.Where(`(meta->>'hidden')::boolean IS FALSE`)
		}
	} else {
		// Factor whether hidden characters should be included or not
		if !ctx.IncludeHidden {
			gensql = gensql.Where(`((meta->>'hidden')::boolean IS FALSE OR (owner_id=? AND (meta->>'hidden')::boolean
			IS FALSE))`, ctx.Collaborator.UserID)
		} else {
			gensql = gensql.Where(`((meta->>'hidden')::boolean IS FALSE OR owner_id=?)`, ctx.Collaborator.UserID)
		}
	}

	// Factor whether characters should be sorted nominally or lexicographically
	if ctx.Sort == dtos.CharacterQuerySortLexicographical {
		gensql = gensql.OrderBy(`meta->'name'->>'lastName' = '' OR meta->'name'->>'firstName' = '' OR (meta->>
		'nameHidden')::boolean IS TRUE, CASE WHEN meta->'name'->>'preferredName' != '' THEN meta->'name'->>
		'preferredName' ELSE meta->'name'->>'lastName' END, meta->'name'->>'lastName', meta->'name'->>'firstName'`)
	} else {
		gensql = gensql.OrderBy(`(meta->>'nameHidden')::boolean IS TRUE, name`)
	}

	// Apply the rest of the statements
	gensql = gensql.Limit(uint64(s.Config.CharacterPageLimit)).Offset(uint64(ctx.Page * s.Config.CharacterPageLimit))

	// Convert to SQL statement
	querysql, queryargs, err := gensql.ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Create the count query
	gensql = s.Providers.SQLBuilder.Select(`COUNT(*)`).From(`characters`).Where(`universe_id = ? AND name ILIKE ?`,
		universe.ID, query)

	// Factor whether all characters should be included in the query
	if ctx.Collaborator.Role != models.CollaboratorMember {
		// Factor whether hidden characters should be included or not
		if !ctx.IncludeHidden {
			gensql = gensql.Where(`(meta->>'hidden')::boolean IS FALSE`)
		}
	} else {
		// Factor whether hidden charactesr should be included or not
		if !ctx.IncludeHidden {
			gensql = gensql.Where(`((meta->>'hidden')::boolean IS FALSE OR (owner_id=? AND (meta->>'hidden')::boolean
			IS FALSE))`, ctx.Collaborator.UserID)
		} else {
			gensql = gensql.Where(`((meta->>'hidden')::boolean IS FALSE OR owner_id=?)`, ctx.Collaborator.UserID)
		}
	}

	// Convert to SQL statement
	countsql, countargs, err := gensql.ToSql()
	if err != nil {
		return nil, 0, err
	}

	// Run the queries
	if err := s.Providers.DB.Select(&characters, querysql, queryargs...); err != nil {
		return nil, 0, err
	}
	if err := s.Providers.DB.Get(&count, countsql, countargs...); err != nil {
		return nil, 0, err
	}

	for i, c := range characters {
		if ctx.Collaborator.Role == models.CollaboratorMember && c.OwnerID != ctx.Collaborator.UserID {
			characters[i].HideHiddenFields()
		}
	}

	return &characters, count, nil
	/*if ctx.Collaborator.Role != models.CollaboratorMember {
		var err error
		// Use admin database query
		if ctx.Sort == dtos.CharacterQuerySortLexicographical {
			// Use admin lexicographical database query

		} else {
			// Use admin nominal database query
			q, args, _ := s.Providers.SQLBuilder.Select(`id, name, tag, owner_id, created_at, updated_at,
			character_images.url AS avatar_url, meta->'hidden' AS hidden, CASE WHEN meta->'nameHidden' IS NULL THEN false ELSE (meta->>'nameHidden')::boolean END AS name_hidden, meta->'name' AS parsed_name`).
				From(`characters`).LeftJoin(`character_images ON character_images.character_id = characters.id`).
				Where(`universe_id = ? AND name ILIKE ?`, universe.ID, query).
				OrderBy(`CASE WHEN (meta->>'nameHidden')::boolean=true THEN NULL ELSE name END, name`).Limit(
				uint64(s.Config.CharacterPageLimit)).Offset(uint64(ctx.Page)).ToSql()
			err = s.Providers.DB.Select(
				&characters,
				q,
				args...,
			)
		}

		if err != nil {
			return nil, 0, err
		}

		// Use admin database count query
		if err := s.Providers.DB.Get(
			&count,
			QueryFindByUniverseAllCount,
			universe.ID,
			query,
		); err != nil {
			return nil, 0, err
		}
	} else {
		var err error
		// Use public database query
		if ctx.Sort == dtos.CharacterQuerySortLexicographical {
			// Use public lexicographical database query
			err = s.Providers.DB.Select(
				&characters,
				QueryFindByUniversePublicLex,
				universe.ID,
				ctx.Collaborator.UserID,
				query,
				s.Config.CharacterPageLimit,
				ctx.Page*s.Config.CharacterPageLimit,
			)
		} else {
			// Use public nominal database query
			err = s.Providers.DB.Select(
				&characters,
				QueryFindByUniversePublicNom,
				universe.ID,
				ctx.Collaborator.UserID,
				query,
				s.Config.CharacterPageLimit,
				ctx.Page*s.Config.CharacterPageLimit,
			)
		}

		if err != nil {
			return nil, 0, err
		}

		// Use public database count query
		if err := s.Providers.DB.Get(
			&count,
			QueryFindByUniversePublicCount,
			universe.ID,
			query,
			ctx.Collaborator.UserID,
		); err != nil {
			return nil, 0, err
		}
	}

	return &characters, count, nil*/
	/*if ctx.Collaborator.Role != models.CollaboratorMember {
		qs := "LEFT JOIN character_images ON character_images.character_id = characters.id WHERE universe_id = $1"
		query = fmt.Sprintf("%v %v AND name ILIKE $2 ORDER BY name LIMIT $3 OFFSET $4", query, qs)
		cquery = fmt.Sprintf("%v %v AND name ILIKE $2", cquery, qs)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&characters,
			query,
			universe.ID,
			ctx.Query,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}

		// Retrieve total count
		if err := s.Providers.DB.Get(&count, cquery, universe.ID, ctx.Query); err != nil {
			return nil, 0, err
		}
	} else {
		qs := `LEFT JOIN character_images ON character_images.character_id = characters.id WHERE universe_id = $1 AND
		(meta->'hidden'='false' OR owner_id=$2)`
		query = fmt.Sprintf("%v %v AND name ILIKE $3 ORDER BY name LIMIT $4 OFFSET $5", query, qs)
		cquery = fmt.Sprintf("%v %v AND name ILIKE $3", cquery, qs)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&characters,
			query,
			universe.ID,
			ctx.Collaborator.UserID,
			ctx.Query,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}

		// Retrieve total count
		if err := s.Providers.DB.Get(&count, cquery, universe.ID, ctx.Collaborator.UserID, ctx.Query); err != nil {
			return nil, 0, err
		}
	}
	return &characters, count, nil*/
}

// Search returns a selection of characters according to a specified search query
func (s *Service) Search(
	universe *models.Universe,
	squery string,
	ctx dtos.CharacterQuery,
) (*[]models.CharacterReference, int, error) {
	squery = strings.Replace(squery, " ", "%", -1)
	var (
		references []models.CharacterReference
		count      = 0
		query      = `SELECT id, name, tag, owner_id, created_at, updated_at, character_images.url AS avatar_url
		FROM characters WHERE name ILIKE $1 AND universe_id = $2`
		cquery = `SELECT count(*) FROM characters WHERE name ILIKE $1 AND universe_id = $2`
	)
	if ctx.Collaborator.Role != models.CollaboratorMember {
		query = fmt.Sprintf("%v ORDER BY name LIMIT $3 OFFSET $4 ", query)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&references,
			query,
			squery,
			universe.ID,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}
	} else {
		query = fmt.Sprintf("%v AND (meta->'hidden'='false' OR owner_id=$3) ORDER BY name LIMIT $4 OFFSET $5", query)

		// Retrieve characters
		if err := s.Providers.DB.Select(
			&references,
			query,
			squery,
			universe.ID,
			ctx.Collaborator.UserID,
			s.Config.CharacterPageLimit,
			ctx.Page*s.Config.CharacterPageLimit,
		); err != nil {
			return nil, 0, err
		}
	}
	if err := s.Providers.DB.Get(&count, cquery, squery, universe.ID); err != nil {
		return nil, 0, err
	}

	return &references, count, nil
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
		($1, $2, $3, $4, $5, $6, $7) RETURNING id, universe_id, name, tag, fields, meta`,
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
	c.Owner = owner
	return &c, nil
}

// Update updates an existing character in the database
func (s *Service) Update(character *models.Character) (*models.Character, error) {
	var c models.Character
	character.UpdatedAt = time.Now()
	rows, err := s.Providers.DB.NamedQuery(
		`UPDATE characters SET name = :name, tag = :tag, fields = :fields, meta = :meta,
		updated_at = :updated_at WHERE id = :id RETURNING id, name, tag, fields, meta, updated_at, created_at`,
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
	images, err := s.FindCharacterImages(character.ID)
	if err != nil {
		return nil, err
	}
	c.Images = images
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
						fmt.Printf("%v - %T", cField.Value, cField.Value)
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
				break
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

// SetImage assigns an image to a character
func (s *Service) SetImage(character *models.Character, key string, image io.Reader) error {
	path := fmt.Sprintf("%s_%s", character.ID, key)

	optimized, err := s.optimizeImage(image)
	if err != nil {
		return err
	}

	url, err := s.Providers.Storage.Upload(optimized, path)
	if err != nil {
		return err
	}
	if _, err := s.Providers.DB.Exec(
		`INSERT INTO character_images (character_id, key, url) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		character.ID,
		key,
		url,
	); err != nil {
		return err
	}
	return nil
}

// DeleteImage removes an image associated with a character
func (s *Service) DeleteImage(character *models.Character, key string) error {
	path := fmt.Sprintf("%v_%v", character.ID, key)
	if err := s.Providers.Storage.Delete(path); err != nil {
		return err
	}
	if _, err := s.Providers.DB.Exec(
		`DELETE FROM character_images WHERE character_id=$1 AND key=$2`,
		character.ID,
		key,
	); err != nil {
		return err
	}
	return nil
}

// DeleteAll deletes all characters from a specified universe
func (s *Service) DeleteAll(universe *models.Universe) error {
	var ids []models.Character
	if err := s.Providers.DB.Select(
		&ids,
		`SELECT id FROM characters WHERE universe_id = $1`,
		universe.ID,
	); err != nil {
		return err
	}
	for _, id := range ids {
		s.DeleteImage(&id, "avatar")
	}
	if _, err := s.Providers.DB.Exec(`DELETE FROM characters WHERE universe_id = $1`, universe.ID); err != nil {
		return err
	}
	return nil
}

func (s *Service) optimizeImage(file io.Reader) (io.Reader, error) {
	buff := new(bytes.Buffer)
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	dstImg := imaging.Thumbnail(img, AvatarSize, AvatarSize, imaging.CatmullRom)
	if err := jpeg.Encode(buff, dstImg, nil); err != nil {
		return nil, err
	}

	return bytes.NewReader(buff.Bytes()), nil
}
