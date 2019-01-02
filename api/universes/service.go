package universes

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jinzhu/gorm/dialects/postgres"
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
		Name:        data.Name,
		Description: data.Description,
		Guide: postgres.Jsonb{
			RawMessage: json.RawMessage(DefaultUniverseGuideJSON),
		},
		Settings: postgres.Jsonb{
			RawMessage: json.RawMessage(DefaultUniverseSettingsJSON),
		},
	}
	return universe
}

// Find returns all Universes
func (s *Service) Find() (*[]models.Universe, error) {
	var universes []models.Universe
	if err := s.Providers.DB.Find(&universes).Error; err != nil {
		return nil, err
	}
	return &universes, nil
}

// FindByID returns a Universe by their ID
func (s *Service) FindByID(id string) (*models.Universe, error) {
	var universe models.Universe
	if err := s.Providers.DB.Where("id = ?", id).First(&universe).Error; err != nil {
		return nil, err
	}
	return &universe, nil
}

// FindByOwner returns a selection of universes the user owns
func (s *Service) FindByOwner(owner *models.User) (*[]models.Universe, error) {
	var universes []models.Universe
	if err := s.Providers.DB.Where("owner = ?", owner.ID).Find(&universes).Error; err != nil {
		return nil, err
	}
	return &universes, nil
}

// FindFromUser returns a selection of universe references the user is collaborating in
func (s *Service) FindFromUser(user *models.User) (*[]models.UniverseReference, error) {
	// Calling make initializes the slice to prevent JSON from marshalling empty results into "null"
	universes := make([]models.UniverseReference, 0)
	rows, err := s.Providers.DB.Table("collaborators").Select("universes.id, universes.name").Joins(
		"join universes on universes.id = collaborators.universe_id",
	).Where("user_id = ?", user.ID).Rows()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var universe models.UniverseReference
		if err := s.Providers.DB.ScanRows(rows, &universe); err != nil {
			return nil, err
		}
		universes = append(universes, universe)
	}
	return &universes, nil
}

// FindCollaboratorFromUser returns a collaborator relation from a given universe ID and user
func (s *Service) FindCollaboratorFromUser(uid string, user *models.User) (*models.Collaborator, error) {
	var collaborator models.Collaborator
	if err := s.Providers.DB.Where(
		"user_id = ? AND universe_id = ?", user.ID, uid,
	).First(&collaborator).Error; err != nil {
		return nil, err
	}
	return &collaborator, nil
}

// Save saves a universe to the database
func (s *Service) Save(universe *models.Universe, owner *models.User) error {
	tx := s.Providers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if s.Providers.DB.NewRecord(&universe) {
		if err := tx.Create(&universe).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Save(&universe).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// NOTE: GORM likes to auto-create entries, which can destroy pre-existing data
	// Collaborator creation MUST come after universe creation so this doesn't happen
	if owner != nil {
		if !s.Providers.DB.NewRecord(&universe) {
			if err := tx.Where(
				"role = ? AND universe_id = ?",
				models.CollaboratorOwner,
				universe.ID,
			).Delete(
				models.Collaborator{},
			).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		collaborator := &models.Collaborator{
			UniverseID: universe.ID,
			UserID:     owner.ID,
			Role:       models.CollaboratorOwner,
		}
		if err := tx.Create(&collaborator).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
