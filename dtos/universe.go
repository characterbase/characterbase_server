package dtos

import (
	"cbs/models"
)

// ReqCreateUniverse represents a request DTO for creating a new universe
type ReqCreateUniverse struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:""`
}

// ReqEditUniverse represents a request DTO for modifying an existing universe
type ReqEditUniverse struct {
	*models.Universe
}

// ReqAddCollaborator represents a request DTO for adding a new collaborator to a universe
type ReqAddCollaborator struct {
	ID    string                  `json:"id"`
	Email string                  `json:"email"`
	Role  models.CollaboratorRole `json:"role" validate:"oneof=0 1"`
}

// ReqEditCollaborator represents a request DTO for modifying an existing collaborator
type ReqEditCollaborator struct {
	ID   string                  `json:"id" validate:"required"`
	Role models.CollaboratorRole `json:"role" validate:"oneof=0 1"`
}

// ReqRemoveCollaborator represents a request DTO for deleting an existing collaborator
type ReqRemoveCollaborator struct {
	ID string `json:"id" validate:"required"`
}

// ResGetUniverse represents a response DTO containing universe data
type ResGetUniverse struct {
	*models.Universe
}

// ResGetUniverses represents a response DTO containing a collection of universe data
type ResGetUniverses struct {
	References *[]models.UniverseReference `json:"universes"`
}

// ResGetCollaborator represents a response DTO containing collaborator data
type ResGetCollaborator struct {
	*models.Collaborator
}

// ResGetCollaborators represents a response DTO containing a collection of universe collaborators
type ResGetCollaborators struct {
	Collaborators *[]models.Collaborator `json:"collaborators"`
}
