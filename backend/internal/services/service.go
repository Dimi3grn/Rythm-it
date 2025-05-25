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

// PaginationParams paramètres de pagination
type PaginationParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}

// DefaultPagination retourne les paramètres de pagination par défaut
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 10,
		Sort:    "id",
		Order:   "DESC",
	}
}

// ValidatePagination valide et normalise les paramètres de pagination
func ValidatePagination(params *PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}

	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 10
	}

	if params.Order != "ASC" && params.Order != "DESC" {
		params.Order = "DESC"
	}
}
