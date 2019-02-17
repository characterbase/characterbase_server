package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ProgressBarColor represents a bar color for a Progress field with show bar enabled
type ProgressBarColor string

// GuideFieldType represents a type for universe guide fields and character fields
type GuideFieldType string

// CollaboratorRole represents a basic role assigned to all universe collaborators
type CollaboratorRole int

// All the available progress bar colors
var (
	BarColorRed    ProgressBarColor = "red"
	BarColorYellow ProgressBarColor = "yellow"
	BarColorGreen  ProgressBarColor = "green"
	BarColorBlue   ProgressBarColor = "blue"
	BarColorTeal   ProgressBarColor = "teal"
	BarColorGray   ProgressBarColor = "gray"
	BarColorDark   ProgressBarColor = "dark"
)

// All the available guide field types
var (
	GuideFieldText        GuideFieldType = "text"
	GuideFieldDescription GuideFieldType = "description"
	GuideFieldNumber      GuideFieldType = "number"
	GuideFieldToggle      GuideFieldType = "toggle"
	GuideFieldProgress    GuideFieldType = "progress"
	GuideFieldReference   GuideFieldType = "reference"
	GuideFieldOptions     GuideFieldType = "options"
	GuideFieldList        GuideFieldType = "list"
	GuideFieldPicture     GuideFieldType = "picture"
)

var (
	// CollaboratorMember represents a collaborator with basic permissions
	CollaboratorMember CollaboratorRole
	// CollaboratorAdmin represents a collaborator with admin permissions
	CollaboratorAdmin CollaboratorRole = 1
	// CollaboratorOwner represents the singular owner of the universe
	CollaboratorOwner CollaboratorRole = 2
)

// Universe represents a CharacterBase universe
type Universe struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	Guide       *UniverseGuide    `json:"guide" db:"guide"`
	Settings    *UniverseSettings `json:"settings" db:"settings"`
}

// UniverseReference represents a stripped version of Universe,
// used when sending multiple Universe instances over the API
// for network optimization purposes
type UniverseReference struct {
	ID   string           `json:"id"`
	Name string           `json:"name"`
	Role CollaboratorRole `json:"role"`
}

// UniverseSettings represents settings for a universe
type UniverseSettings struct {
	TitleField                   string `json:"titleField" validate:"required"`
	AllowAvatars                 bool   `json:"allowAvatars"`
	AllowLexicographicalOrdering bool   `json:"allowLexicographicalOrdering"`
}

// UniverseGuide represents a universe guide
type UniverseGuide struct {
	Groups *[]UniverseGuideGroup `json:"groups" validate:"dive"` // "dive" is necessary so slices can get validated
}

// UniverseGuideGroup represents a field group inside of a universe guide
type UniverseGuideGroup struct {
	Name     string                `json:"name" validate:"required"`
	Fields   *[]UniverseGuideField `json:"fields" validate:"min=1,dive"`
	Required bool                  `json:"required"`
}

// UniverseGuideField represents a field inside of a universe guide
type UniverseGuideField struct {
	Name        string         `json:"name" validate:"required"`
	Type        GuideFieldType `json:"type" validate:"oneof=text description number toggle progress options list picture"`
	Description string         `json:"description"`
	Required    bool           `json:"required"`
	Meta        interface{}    `json:"meta" validate:"required"`
}

// UniverseGuideMetaText represents universe guide settings regarding a Text field
type UniverseGuideMetaText struct {
	Pattern   string `json:"pattern" mapstructure:"pattern"`
	MinLength int    `json:"minLength" mapstructure:"minLength" validate:"gte=0"`
	MaxLength int    `json:"maxLength" mapstructure:"maxLength" validate:"gtecsfield=MinLength"`
}

// UniverseGuideMetaDescription represents universe guide settings regarding a Description field
type UniverseGuideMetaDescription struct {
	Markdown  bool `json:"markdown" mapstructure:"markdown"`
	MinLength int  `json:"minLength" mapstructure:"minLength" validate:"gte=0"`
	MaxLength int  `json:"maxLength" mapstructure:"maxLength" validate:"gtecsfield=MinLength"`
}

// UniverseGuideMetaNumber represents universe guide settings regarding a Number field
type UniverseGuideMetaNumber struct {
	Float bool    `json:"float" mapstructure:"float"`
	Min   float64 `json:"min" mapstructure:"min"`
	Max   float64 `json:"max" mapstructure:"max" validate:"gtecsfield=Min"`
	Tick  float64 `json:"tick" mapstructure:"tick" validate:"gte=0"`
}

// UniverseGuideMetaToggle represents universe guide settings regarding a Toggle field
type UniverseGuideMetaToggle struct{}

// UniverseGuideMetaProgress represents universe guide settings regarding a Progress field
type UniverseGuideMetaProgress struct {
	Bar   bool             `json:"bar" mapstructure:"bar"`
	Color ProgressBarColor `json:"color" mapstructure:"color" validate:"oneof=red yellow green blue teal gray dark"`
	Min   float64          `json:"min" mapstructure:"min" validate:"gte=0"`
	Max   float64          `json:"max" mapstructure:"max" validate:"gtecsfield=Min"`
	Tick  float64          `json:"tick" mapstructure:"tick" validate:"gte=0,ltecsfield=Max"`
}

// UniverseGuideMetaOptions represents unvierse guide settings regarding an Options field
type UniverseGuideMetaOptions struct {
	Multiple bool     `json:"multiple" mapstructure:"multiple"`
	Options  []string `json:"options" mapstructure:"options" validate:"min=1"`
}

// UniverseGuideMetaList represents universe guide settings regarding a List field
type UniverseGuideMetaList struct {
	MinElements int `json:"minElements" mapstructure:"minElements" validate:"gte=0"`
	MaxElements int `json:"maxElements" mapstructure:"maxElements" validate:"gtecsfield=MinElements"`
}

// Collaborator represents a universe collaborator
// NOTE: At the current time, GORM doesn't automatically create
// foreign keys with struct tags, so an explicit SQL tag is necessary
type Collaborator struct {
	UniverseID string           `json:"-" db:"universe_id"`
	UserID     string           `json:"userId,omitempty" db:"user_id"`
	User       *User            `json:"user,omitempty" db:"user"`
	Role       CollaboratorRole `json:"role" db:"role"`
}

// Value returns a serialized representation of this guide
func (ug *UniverseGuide) Value() (driver.Value, error) {
	return json.Marshal(ug)
}

// Scan deserializes the serialized representation of this guide
func (ug *UniverseGuide) Scan(val interface{}) error { // "val" represents current type of data scanned from database
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &ug)
	case string:
		json.Unmarshal([]byte(v), &ug)
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
	if err := ug.SetFieldMeta(); err != nil {
		return err
	}
	return nil
}

// Value returns a serialized representation of this settings
func (us *UniverseSettings) Value() (driver.Value, error) {
	return json.Marshal(us)
}

// Scan deserializes the serialized representation of this settings
func (us *UniverseSettings) Scan(val interface{}) error { // "val" represents current type of data scanned from database
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &us)
		return nil
	case string:
		json.Unmarshal([]byte(v), &us)
		return nil
	default:
		return fmt.Errorf("Unsupported type: %T", v)
	}
}

// SetFieldMeta sets the appropriate meta structs for the universe guide's fields based on their types
func (ug *UniverseGuide) SetFieldMeta() error {
	var err error
	for i, v := range *ug.Groups {
		g := (*ug.Groups)[i]
		for j, f := range *v.Fields {
			switch f.Type {
			case GuideFieldText:
				m := UniverseGuideMetaText{}
				err = mapstructure.Decode(f.Meta, &m)
				fmt.Println(m)
				(*g.Fields)[j].Meta = m
			case GuideFieldDescription:
				m := UniverseGuideMetaDescription{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			case GuideFieldNumber:
				m := UniverseGuideMetaNumber{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			case GuideFieldToggle:
				m := UniverseGuideMetaToggle{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			case GuideFieldProgress:
				m := UniverseGuideMetaProgress{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			case GuideFieldOptions:
				m := UniverseGuideMetaOptions{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			case GuideFieldList:
				m := UniverseGuideMetaList{}
				err = mapstructure.Decode(f.Meta, &m)
				(*g.Fields)[j].Meta = m
			}
		}
	}
	if err != nil {
		return err
	}
	return nil
}
