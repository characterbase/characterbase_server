package models

import (
	"encoding/json"
	"errors"
	"fmt"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/jinzhu/gorm/dialects/postgres"
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
	GuideFieldProgress    GuideFieldType = "progress"
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
	Generic
	Name        string         `json:"name" gorm:"not null" validate:"required,omitempty"`
	Description string         `json:"description"`
	Guide       postgres.Jsonb `json:"guide"`
	Settings    postgres.Jsonb `json:"settings"`
}

// UniverseReference represents a stripped version of Universe,
// used when sending multiple Universe instances over the API
// for network optimization purposes
type UniverseReference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UniverseSettings represents settings for a universe
type UniverseSettings struct {
	TitleField   string `json:"title_field"`
	AllowAvatars bool   `json:"allow_avatars"`
}

// UniverseGuide represents a universe guide
type UniverseGuide struct {
	Groups *[]UniverseGuideGroup `json:"groups"`
}

// UniverseGuideGroup represents a field group inside of a universe guide
type UniverseGuideGroup struct {
	Name   string                `json:"name"`
	Fields *[]UniverseGuideField `json:"fields"`
}

// UniverseGuideField represents a field inside of a universe guide
type UniverseGuideField struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default"`
	Meta        interface{} `json:"meta"`
}

// UniverseGuideMetaText represents universe guide settings regarding a Text field
type UniverseGuideMetaText struct {
	Pattern   string `json:"pattern"`
	MinLength int    `json:"min_length"`
	MaxLength int    `json:"max_length"`
}

// UniverseGuideMetaDescription represents universe guide settings regarding a Description field
type UniverseGuideMetaDescription struct {
	Markdown  bool `json:"markdown"`
	MinLength int  `json:"min_length"`
	MaxLength int  `json:"max_length"`
}

// UniverseGuideMetaNumber represents universe guide settings regarding a Number field
type UniverseGuideMetaNumber struct {
	Float bool    `json:"float"`
	Min   float32 `json:"min"`
	Max   float32 `json:"max"`
	Tick  float32 `json:"tick"`
}

// UniverseGuideMetaProgress represents universe guide settings regarding a Progress field
type UniverseGuideMetaProgress struct {
	Bar   bool             `json:"bar"`
	Color ProgressBarColor `json:"color"`
	Min   int              `json:"min"`
	Max   int              `json:"max"`
	Tick  float32          `json:"tick"`
}

// UniverseGuideMetaOptions represents unvierse guide settings regarding an Options field
type UniverseGuideMetaOptions struct {
	Multiple bool     `json:"multiple"`
	Options  []string `json:"options"`
}

// UniverseGuideMetaList represents universe guide settings regarding a List field
type UniverseGuideMetaList struct {
	MinElements int `json:"min_elements"`
	MaxElements int `json:"max_elements"`
}

// Collaborator represents a universe collaborator
// NOTE: At the current time, GORM doesn't automatically create
// foreign keys with struct tags, so an explicit SQL tag is necessary
type Collaborator struct {
	UniverseID string           `json:"universe_id" gorm:"primary_key" sql:"type:text REFERENCES universes(id)"`
	UserID     string           `json:"user_id" gorm:"primary_key" sql:"type:text REFERENCES users(id)"`
	Role       CollaboratorRole `json:"role"`
}

// GetGuide returns the deserialized value of the universe's guide
func (u *Universe) GetGuide() (*UniverseGuide, error) {
	var guide UniverseGuide
	data, err := u.Guide.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &guide); err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", guide)
	return &guide, nil
}

// GetSettings returns the deserialized value of the universe's settings
func (u *Universe) GetSettings() (*UniverseSettings, error) {
	var settings UniverseSettings
	data, err := u.Settings.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// Validate validates the Universe, including its guide and settings
func (u *Universe) Validate() (error, error) {
	v := validator.New()
	v.RegisterTagNameFunc(ValidatorJSONPlugin)
	guide, err := u.GetGuide()
	if err != nil {
		return errors.New("failed to get guide")
	}
	settings, err := u.GetSettings()
	if err != nil {
		return errors.New("failed to get settings")
	}
	uerr := v.Struct(u)
	if uerr != nil {
		verr, err := GetValidationError(uerr)
		if err != nil {
		}
	}
	gerr := v.Struct(guide)
	serr := v.Struct(settings)
}
