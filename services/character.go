package services

import (
	"cbs/dtos"
	"cbs/models"
	"io"
)

// Character represents the Character service layer
type Character interface {
	New(data dtos.ReqCreateCharacter) *models.Character
	FindByID(id string) (*models.Character, error)
	FindByUniverse(universe *models.Universe, ctx dtos.CharacterQuery) (*[]models.CharacterReference, int, error)
	Validate(character *models.Character, universe *models.Universe) error
	SetImage(character *models.Character, key string, image io.Reader) error
	DeleteImage(character *models.Character, key string) error
	Create(universe *models.Universe, character *models.Character, owner *models.User) (*models.Character, error)
	Update(character *models.Character) (*models.Character, error)
	Delete(character *models.Character) error
	DeleteAll(universe *models.Universe) error
	FindCharacterImages(id string) (models.CharacterImages, error)
	Search(universe *models.Universe, query string, ctx dtos.CharacterQuery) (*[]models.CharacterReference, int, error)
}
