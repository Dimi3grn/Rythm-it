package services

import (
	"database/sql"
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

// NewFriendshipServiceWithDB crée un nouveau service d'amitiés avec une connexion DB
func NewFriendshipServiceWithDB(db *sql.DB) FriendshipService {
	friendshipRepo := repositories.NewFriendshipRepository(db)
	userRepo := repositories.NewUserRepository(db)
	return NewFriendshipService(friendshipRepo, userRepo)
}
