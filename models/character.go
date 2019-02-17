package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// CharacterImages represents a map of field keys to image URLs associated with the character
type CharacterImages map[string]string

// Character represents a CharacterBase character
type Character struct {
	ID         string           `json:"id" db:"id"`
	Name       string           `json:"name" db:"name" validate:"required"`
	Tag        string           `json:"tag" db:"tag"`
	Owner      *User            `json:"owner,omitempty" db:"owner"`
	OwnerID    string           `json:"ownerId,omitempty" db:"owner_id"`
	Universe   *Universe        `json:"universe,omitempty" db:"universe"`
	UniverseID string           `json:"universeId,omitempty" db:"universe_id"`
	Fields     *CharacterFields `json:"fields" db:"fields" validate:"required"`
	Images     CharacterImages  `json:"images"`
	CreatedAt  time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time        `json:"updatedAt" db:"updated_at"`
	Meta       *CharacterMeta   `json:"meta" db:"meta" validate:"required"`
}

// CharacterReference represents a data-minimized representation of a Character
type CharacterReference struct {
	ID         string             `json:"id" db:"id"`
	Name       string             `json:"name" db:"name"`
	Tag        string             `json:"tag" db:"tag"`
	OwnerID    string             `json:"ownerId" db:"owner_id"`
	CreatedAt  time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time          `json:"updatedAt" db:"updated_at"`
	AvatarURL  *string            `json:"avatarUrl" db:"avatar_url"`
	Hidden     bool               `json:"hidden" db:"hidden"`
	NameHidden bool               `json:"nameHidden" db:"name_hidden"`
	ParsedName *CharacterMetaName `json:"parsedName" db:"parsed_name"`
}

// CharacterMeta represents underlying information associated with a character
type CharacterMeta struct {
	NameHidden bool               `json:"nameHidden"`
	Hidden     bool               `json:"hidden"`
	Name       *CharacterMetaName `json:"name"`
}

// CharacterMetaName represents a sub-group in meta that categorizes parts of the character's name
type CharacterMetaName struct {
	FirstName     string `json:"firstName"`
	MiddleName    string `json:"middleName"`
	LastName      string `json:"lastName"`
	Nickname      string `json:"nickname"`
	PreferredName string `json:"preferredName"`
}

// CharacterFields represents fields associated with a character
type CharacterFields struct {
	Groups map[string]*CharacterFieldGroup `json:"groups" validate:"dive"`
}

// CharacterFieldGroup represents a field group associated with a character
type CharacterFieldGroup struct {
	Fields map[string]*CharacterField `json:"fields" validate:"dive"`
	Hidden bool                       `json:"hidden"`
}

// CharacterField represents a field associated with a character
type CharacterField struct {
	Value  interface{}    `json:"value"`
	Type   GuideFieldType `json:"type" validate:"oneof=text description number toggle progress options list picture"`
	Hidden bool           `json:"hidden"`
}

// CharacterImage represents an image associated with a character
type CharacterImage struct {
	Character Character
	Field     string
	PublicURL string
}

// Value returns a serialized representation of this character meta
func (m *CharacterMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan deserializes the serialized representation of this character meta
func (m *CharacterMeta) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &m)
		return nil
	case string:
		json.Unmarshal([]byte(v), &m)
		return nil
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// Value returns a serialized representation of the character's parsed name
func (n *CharacterMetaName) Value() (driver.Value, error) {
	return json.Marshal(n)
}

// Scan deserializes the serialized representation of the character's aprsed name
func (n *CharacterMetaName) Scan(val interface{}) error { // "val" represents current type of data scanned from database
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &n)
		return nil
	case string:
		json.Unmarshal([]byte(v), &n)
		return nil
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// Value returns a serialized representation of this character fields
func (f *CharacterFields) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan deserializes the serialized representation of this character fields
func (f *CharacterFields) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &f)
		return nil
	case string:
		json.Unmarshal([]byte(v), &f)
		return nil
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// HideHiddenFields obscures values in the character's fields that are marked as hidden
func (c *Character) HideHiddenFields() {
	if c.Meta.NameHidden {
		c.Name = ""
		c.Tag = ""
		c.Meta.Name = nil
	}
	for _, g := range c.Fields.Groups {
		if g.Hidden {
			g.Fields = nil
		} else {
			for _, f := range g.Fields {
				if f.Hidden {
					f.Value = nil
				}
			}
		}
	}
}

// HideHiddenFields obscures values in the character's fields that are marked as hidden
func (c *CharacterReference) HideHiddenFields() {
	if c.NameHidden {
		c.Name = ""
		c.Tag = ""
		c.ParsedName = nil
	}
}
