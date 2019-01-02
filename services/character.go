package services

import (
	"cbs/dtos"
	"cbs/models"
)

// Character represents the Character service layer
type Character interface {
	New(data dtos.CreateCharacter) *models.Character
	Find() (*[]models.Character, error)
	FindByID(id string) (*models.Character, error)
	FindByUniverse(universe *models.Universe) (*[]models.Character, error)
	FindByOwner(owner *models.User) (*[]models.Character, error)
	Save(universe *models.Universe, character *models.Character) error
}
