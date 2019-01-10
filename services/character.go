package services

import (
	"cbs/dtos"
	"cbs/models"
)

// Character represents the Character service layer
type Character interface {
	New(data dtos.ReqCreateCharacter) *models.Character
	FindByID(id string) (*models.Character, error)
	FindByUniverse(universe *models.Universe, ctx dtos.CharacterQuery) (*[]models.CharacterReference, int, error)
	Validate(character *models.Character, universe *models.Universe) error
	Create(universe *models.Universe, character *models.Character, owner *models.User) (*models.Character, error)
	Update(character *models.Character) (*models.Character, error)
	Delete(character *models.Character) error
}
