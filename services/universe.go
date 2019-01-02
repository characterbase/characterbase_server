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
	FindByOwner(owner *models.User) (*[]models.Universe, error)
	FindFromUser(user *models.User) (*[]models.UniverseReference, error)
	FindCollaboratorFromUser(universeid string, user *models.User) (*models.Collaborator, error)
	Save(universe *models.Universe, owner *models.User) error
}
