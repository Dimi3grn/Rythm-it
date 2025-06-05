package services

import (
	"rythmitbackend/internal/repositories"
)

// Service interface de base pour tous les services
type Service interface {
	// GetRepository retourne le repository utilisé par ce service
	GetRepository() repositories.Repository
}

// BaseService structure de base pour tous les services
type BaseService struct {
	repo repositories.Repository
}

// NewBaseService crée un nouveau service de base
func NewBaseService(repo repositories.Repository) *BaseService {
	return &BaseService{
		repo: repo,
	}
}

// GetRepository retourne le repository
func (s *BaseService) GetRepository() repositories.Repository {
	return s.repo
}
