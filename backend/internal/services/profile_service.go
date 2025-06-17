package services

import (
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"
)

// ProfileService interface pour les services de profil
type ProfileService interface {
	GetProfile(userID uint) (*ProfileResponseDTO, error)
	UpdateProfile(userID uint, dto ProfileUpdateDTO) (*ProfileResponseDTO, error)
	CreateProfile(userID uint, dto ProfileCreateDTO) (*ProfileResponseDTO, error)
	GetOrCreateProfile(userID uint) (*ProfileResponseDTO, error)
}

// ProfileCreateDTO données pour créer un profil
type ProfileCreateDTO struct {
	DisplayName *string `json:"display_name" validate:"omitempty,max=100"`
	AvatarImage *string `json:"avatar_image" validate:"omitempty"`
	BannerImage *string `json:"banner_image" validate:"omitempty"`
}

// ProfileUpdateDTO données pour mettre à jour un profil
type ProfileUpdateDTO struct {
	DisplayName *string `json:"display_name" validate:"omitempty,max=100"`
	AvatarImage *string `json:"avatar_image" validate:"omitempty"`
	BannerImage *string `json:"banner_image" validate:"omitempty"`
}

// ProfileResponseDTO profil utilisateur en réponse
type ProfileResponseDTO struct {
	ID          uint    `json:"id"`
	UserID      uint    `json:"user_id"`
	DisplayName *string `json:"display_name"`
	AvatarImage *string `json:"avatar_image"`
	BannerImage *string `json:"banner_image"`
}

// profileService implémentation du service de profil
type profileService struct {
	profileRepo repositories.ProfileRepository
	userRepo    repositories.UserRepository
}

// NewProfileService crée une nouvelle instance du service de profil
func NewProfileService(profileRepo repositories.ProfileRepository, userRepo repositories.UserRepository) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
		userRepo:    userRepo,
	}
}

// GetProfile récupère le profil d'un utilisateur
func (s *profileService) GetProfile(userID uint) (*ProfileResponseDTO, error) {
	// Vérifier que l'utilisateur existe
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Récupérer le profil
	profile, err := s.profileRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération profil: %w", err)
	}

	return ToProfileResponseDTO(profile), nil
}

// UpdateProfile met à jour le profil d'un utilisateur
func (s *profileService) UpdateProfile(userID uint, dto ProfileUpdateDTO) (*ProfileResponseDTO, error) {
	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return nil, fmt.Errorf("données invalides: %v", validationErrors)
	}

	// Vérifier que l'utilisateur existe
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Récupérer le profil existant
	profile, err := s.profileRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("profil non trouvé: %w", err)
	}

	// Mettre à jour les champs modifiés
	if dto.DisplayName != nil {
		profile.DisplayName = dto.DisplayName
	}
	if dto.AvatarImage != nil {
		profile.AvatarImage = dto.AvatarImage
	}
	if dto.BannerImage != nil {
		profile.BannerImage = dto.BannerImage
	}

	// Sauvegarder en base
	if err := s.profileRepo.Update(profile); err != nil {
		return nil, fmt.Errorf("erreur mise à jour profil: %w", err)
	}

	return ToProfileResponseDTO(profile), nil
}

// CreateProfile crée un nouveau profil pour un utilisateur
func (s *profileService) CreateProfile(userID uint, dto ProfileCreateDTO) (*ProfileResponseDTO, error) {
	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return nil, fmt.Errorf("données invalides: %v", validationErrors)
	}

	// Vérifier que l'utilisateur existe
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Vérifier qu'un profil n'existe pas déjà
	exists, err := s.profileRepo.ExistsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification existence profil: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("un profil existe déjà pour cet utilisateur")
	}

	// Créer le nouveau profil
	profile := &models.UserProfile{
		UserID:      userID,
		DisplayName: dto.DisplayName,
		AvatarImage: dto.AvatarImage,
		BannerImage: dto.BannerImage,
	}

	// Sauvegarder en base
	if err := s.profileRepo.Create(profile); err != nil {
		return nil, fmt.Errorf("erreur création profil: %w", err)
	}

	return ToProfileResponseDTO(profile), nil
}

// GetOrCreateProfile récupère le profil ou le crée s'il n'existe pas
func (s *profileService) GetOrCreateProfile(userID uint) (*ProfileResponseDTO, error) {
	// Vérifier que l'utilisateur existe
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Essayer de récupérer le profil existant
	profile, err := s.profileRepo.FindByUserID(userID)
	if err != nil {
		// Si le profil n'existe pas, le créer avec des valeurs par défaut
		profile = &models.UserProfile{
			UserID:      userID,
			DisplayName: nil, // Sera nil par défaut
			AvatarImage: nil,
			BannerImage: nil,
		}

		if err := s.profileRepo.Create(profile); err != nil {
			return nil, fmt.Errorf("erreur création profil par défaut: %w", err)
		}
	}

	return ToProfileResponseDTO(profile), nil
}

// ToProfileResponseDTO convertit un UserProfile en ProfileResponseDTO
func ToProfileResponseDTO(profile *models.UserProfile) *ProfileResponseDTO {
	if profile == nil {
		return nil
	}

	return &ProfileResponseDTO{
		ID:          profile.ID,
		UserID:      profile.UserID,
		DisplayName: profile.DisplayName,
		AvatarImage: profile.AvatarImage,
		BannerImage: profile.BannerImage,
	}
}
