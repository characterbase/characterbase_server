package dtos

import (
	"cbs/models"
)

// ReqCreateUniverse represents a request DTO for creating a new universe
type ReqCreateUniverse struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:""`
}

// ReqEditUniverse represents a request DTO for editing an existing universe
type ReqEditUniverse struct {
	*models.Universe
}

// ResGetUniverse represents a response DTO containing universe data
type ResGetUniverse struct {
	*models.Universe
}

// ResGetUniverses represents a response DTO containing a collection of universe data
type ResGetUniverses struct {
	References *[]models.UniverseReference `json:"universes"`
}
