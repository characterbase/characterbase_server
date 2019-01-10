package dtos

import (
	"cbs/models"
)

// CharacterQuery represents injectable context information for querying datasets from this service
type CharacterQuery struct {
	Collaborator *models.Collaborator
	Page         int
	Query        string
}

// ReqCreateCharacter represents a request DTO for creating a new character
type ReqCreateCharacter struct {
	Name   string              `json:"name" validate:"required"`
	Tag    string              `json:"tag"`
	Fields *ReqCharacterFields `json:"fields" validate:"required"`
	Meta   *ReqCharacterMeta   `json:"meta" validate:"required"`
}

// ReqCharacterMeta represents a sub-request DTO for a character's meta settings
type ReqCharacterMeta struct {
	Archived bool `json:"archived"`
	Hidden   bool `json:"hidden"`
}

// ReqCharacterFields represents a sub-request DTO for a character's fields
type ReqCharacterFields struct {
	Groups map[string]ReqCharacterGroup `json:"groups" validate:"dive"`
}

// ReqCharacterGroup represents a sub-request DTO for a character field group
type ReqCharacterGroup struct {
	Fields map[string]ReqCharacterField `json:"fields" validate:"dive"`
	Hidden bool                         `json:"hidden"`
}

// ReqCharacterField represents a sub-request DTO for a character field
type ReqCharacterField struct {
	Value  interface{}           `json:"value"`
	Type   models.GuideFieldType `json:"type" validate:"oneof=text description number toggle progress options list picture"`
	Hidden bool                  `json:"hidden"`
}

// ResGetCharacter represents a response DTO containing a character's information
type ResGetCharacter struct {
	*models.Character
}

// ResGetCharacters represents a response DTO containing a collection of character information
type ResGetCharacters struct {
	Characters *[]models.CharacterReference `json:"characters"`
	Page       int                          `json:"page"`
	Total      int                          `json:"total"`
}
