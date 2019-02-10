package services

import (
	"cbs/dtos"
	"cbs/models"
)

// Universe represents the Universe service layer
type Universe interface {
	New(data dtos.ReqCreateUniverse) *models.Universe
	Find() (*[]models.Universe, error)
	FindByID(id string) (*models.Universe, error)
	FindFromUser(user *models.User) (*[]models.UniverseReference, error)
	FindCollaborators(universe *models.Universe) (*[]models.Collaborator, error)
	FindCollaboratorByID(universeid string, userid string) (*models.Collaborator, error)
	CreateCollaborator(
		universe *models.Universe,
		user *models.User,
		role models.CollaboratorRole,
	) (*models.Collaborator, error)
	UpdateCollaborator(universe *models.Universe, collaborator *models.Collaborator) (*models.Collaborator, error)
	Create(universe *models.Universe, owner *models.User) error
	Update(universe *models.Universe, owner *models.User) error
	Delete(universe *models.Universe) error
	RemoveCollaborator(universe *models.Universe, collaborator *models.Collaborator) error
}
