package models

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Character represents a CharacterBase character
type Character struct {
	Generic
	Name       string
	Owner      User
	OwnerID    string
	Universe   Universe
	UniverseID string
	Fields     postgres.Jsonb
	Images     []CharacterImage
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CharacterImage represents an image associated with a character
type CharacterImage struct {
	Character Character
	Field     string
	PublicURL string
}
