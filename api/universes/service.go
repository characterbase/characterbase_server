package universes

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/ksuid"
)

// DefaultUniverseGuide represents the default guide given to all new universes
var DefaultUniverseGuide = &models.UniverseGuide{
	Groups: &[]models.UniverseGuideGroup{
		{
			Name: "General",
			Fields: &[]models.UniverseGuideField{
				{
					Name:        "Biography",
					Description: "The history of this character",
					Required:    true,
					Default:     nil,
					Meta: &models.UniverseGuideMetaText{
						MinLength: 1,
						MaxLength: 4000,
						Pattern:   "",
					},
				},
			},
		},
	},
}

// DefaultUniverseSettings represents the default settings given to all new universes
var DefaultUniverseSettings = &models.UniverseSettings{
	TitleField:   "Name",
	AllowAvatars: true,
}

// DefaultUniverseGuideJSON represents the JSON marshalled version of DefaultUniverseGuide
var DefaultUniverseGuideJSON []byte

// DefaultUniverseSettingsJSON represents the JSON marshalled version of DefaultUniverseSettings
var DefaultUniverseSettingsJSON []byte

// Service represents a service implementation for the "universes" resource
type Service api.Service

func init() {
	defaultGuide, err := json.Marshal(DefaultUniverseGuide)
	if err != nil {
		log.Fatal("failed to marshal default guide")
	}
	defaultSettings, err := json.Marshal(DefaultUniverseSettings)
	if err != nil {
		log.Fatal("failed to marshal default settings")
	}
	DefaultUniverseGuideJSON = defaultGuide
	DefaultUniverseSettingsJSON = defaultSettings
}

// New creates a new universe
func (s *Service) New(data dtos.ReqCreateUniverse) *models.Universe {
	universe := &models.Universe{
		ID:          ksuid.New().String(),
		Name:        data.Name,
		Description: data.Description,
		Guide:       DefaultUniverseGuide,
		Settings:    DefaultUniverseSettings,
	}
	return universe
}

// Find returns all Universes
func (s *Service) Find() (*[]models.Universe, error) {
	var universes []models.Universe
	if err := s.Providers.DB.Select(
		&universes,
		"SELECT id, name, description, guide, settings FROM universes",
	); err != nil {
		return nil, err
	}
	return &universes, nil
}

// FindByID returns a Universe by their ID
func (s *Service) FindByID(id string) (*models.Universe, error) {
	var universe models.Universe
	if err := s.Providers.DB.Get(
		&universe,
		"SELECT id, name, description, guide, settings FROM universes WHERE id = $1",
		id,
	); err != nil {
		return nil, err
	}
	return &universe, nil
}

// FindFromUser returns a selection of universe references the user is collaborating in
func (s *Service) FindFromUser(user *models.User) (*[]models.UniverseReference, error) {
	// Calling make initializes the slice to prevent JSON from marshalling empty results into "null"
	universes := make([]models.UniverseReference, 0)
	if err := s.Providers.DB.Select(
		&universes,
		`SELECT universes.id, universes.name, collaborators.role FROM collaborators JOIN universes
		ON universes.id = collaborators.universe_id WHERE user_id = $1`,
		user.ID,
	); err != nil {
		return nil, err
	}
	return &universes, nil
}

// FindCollaboratorByID returns a collaborator relation from a given universe ID and user
func (s *Service) FindCollaboratorByID(universeid string, userid string) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	if err := s.Providers.DB.Get(
		&collaborator,
		"SELECT universe_id, user_id, role FROM collaborators WHERE universe_id = $1 AND user_id = $2",
		universeid,
		userid); err != nil {
		return nil, err
	}
	return &collaborator, nil
}

// Create creates a new universe in the database
func (s *Service) Create(universe *models.Universe, owner *models.User) error {
	tx, err := s.Providers.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	rows, err := tx.NamedQuery(
		`INSERT INTO universes (id, name, description, guide, settings) VALUES (:id, 
	:name, :description, :guide, :settings) RETURNING id, name, description, guide, settings`,
		universe,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(universe); err != nil {
			return err
		}
	}
	_, err = tx.Exec(
		`INSERT INTO collaborators (universe_id, user_id, role) VALUES ($1,$2,$3)`,
		universe.ID,
		owner.ID,
		models.CollaboratorOwner,
	)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Update updates an existing universe in the database
func (s *Service) Update(universe *models.Universe, owner *models.User) error {
	tx, err := s.Providers.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	rows, err := tx.NamedQuery(
		`UPDATE universes SET name = :name, description = :description, guide = :guide,
	settings = :settings WHERE id = :id RETURNING id, name, description, guide, settings`,
		universe,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(universe); err != nil {
			return err
		}
	}
	if owner != nil {
		_, err = tx.Exec(`DELETE FROM collaborators WHERE universe_id = $1 AND role = 2`, universe.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`INSERT INTO collaborators (universe_id, user_id, role) VALUES ($1,$2,$3)`,
			universe.ID,
			owner.ID,
			models.CollaboratorOwner,
		)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// FindCollaborators returns a list of collaborators pertaining to a universe
func (s *Service) FindCollaborators(universe *models.Universe) (*[]models.Collaborator, error) {
	collaborators := make([]models.Collaborator, 0)
	if err := s.Providers.DB.Select(
		&collaborators,
		`SELECT users.id "user.id", users.display_name "user.display_name", users.email "user.email",
		collaborators.role FROM collaborators JOIN users ON users.id = collaborators.user_id WHERE universe_id = $1`,
		universe.ID,
	); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &collaborators, nil
}

// CreateCollaborator adds a collaborator to a universe
func (s *Service) CreateCollaborator(
	universe *models.Universe,
	user *models.User,
	role models.CollaboratorRole,
) (*models.Collaborator, error) {
	var c models.Collaborator
	if err := s.Providers.DB.Get(
		&c,
		`INSERT INTO collaborators (universe_id, user_id, role) VALUES ($1, $2, $3) RETURNING *`,
		universe.ID,
		user.ID,
		role,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCollaborator updates an existing collaborator
func (s *Service) UpdateCollaborator(
	universe *models.Universe,
	collaborator *models.Collaborator,
) (*models.Collaborator, error) {
	var c models.Collaborator
	if err := s.Providers.DB.Get(
		&c,
		`UPDATE collaborators SET role = $1 WHERE universe_id = $2 AND user_id = $3 RETURNING *`,
		collaborator.Role,
		universe.ID,
		collaborator.UserID,
	); err != nil {
		return nil, err
	}
	return &c, nil
}

// RemoveCollaborator deletes an existing collaborator
func (s *Service) RemoveCollaborator(universe *models.Universe, collaborator *models.Collaborator) error {
	if _, err := s.Providers.DB.Exec(
		"DELETE FROM collaborators WHERE universe_id = $1 AND user_id = $2",
		universe.ID,
		collaborator.UserID,
	); err != nil {
		return err
	}
	return nil
}
