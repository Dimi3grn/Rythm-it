package services

import (
	"errors"
	"fmt"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService interface pour les services d'authentification
type AuthService interface {
	Register(dto RegisterDTO) (*models.User, error)
	Login(dto LoginDTO) (string, *models.User, error)
	ValidatePassword(user *models.User, password string) bool
	GenerateToken(user *models.User) (string, error)
	ParseToken(tokenString string) (*CustomClaims, error)
	RefreshToken(oldToken string) (string, error)
}

// RegisterDTO données d'inscription
type RegisterDTO struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

// LoginDTO données de connexion
type LoginDTO struct {
	Identifier string `json:"identifier" validate:"required"` // Email ou Username
	Password   string `json:"password" validate:"required"`
}

// UserResponseDTO utilisateur sans mot de passe
type UserResponseDTO struct {
	ID             uint       `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	IsAdmin        bool       `json:"is_admin"`
	ProfilePic     *string    `json:"profile_pic"`
	Biography      *string    `json:"biography"`
	LastConnection *time.Time `json:"last_connection"`
	MessageCount   int        `json:"message_count"`
	ThreadCount    int        `json:"thread_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AuthResponseDTO réponse d'authentification
type AuthResponseDTO struct {
	Token string          `json:"token"`
	User  UserResponseDTO `json:"user"`
}

// CustomClaims claims JWT personnalisées
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// authService implémentation du service d'authentification
type authService struct {
	userRepo repositories.UserRepository
	config   *configs.Config
}

// NewAuthService crée une nouvelle instance du service d'authentification
func NewAuthService(userRepo repositories.UserRepository, config *configs.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		config:   config,
	}
}

// Register inscrit un nouvel utilisateur
func (s *authService) Register(dto RegisterDTO) (*models.User, error) {
	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return nil, fmt.Errorf("données invalides: %v", validationErrors)
	}

	// Vérifier que l'email n'existe pas déjà
	emailExists, err := s.userRepo.ExistsByEmail(dto.Email)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification email: %w", err)
	}
	if emailExists {
		return nil, utils.ErrEmailAlreadyUsed
	}

	// Vérifier que le username n'existe pas déjà
	usernameExists, err := s.userRepo.ExistsByUsername(dto.Username)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification username: %w", err)
	}
	if usernameExists {
		return nil, utils.ErrUsernameTaken
	}

	// Hacher le mot de passe
	hashedPassword, err := s.hashPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("erreur hachage mot de passe: %w", err)
	}

	// Créer le nouvel utilisateur
	user := &models.User{
		Username: dto.Username,
		Email:    dto.Email,
		Password: hashedPassword,
		IsAdmin:  false, // Par défaut, pas admin
	}

	// Sauvegarder en base
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("erreur création utilisateur: %w", err)
	}

	// Retourner l'utilisateur sans le mot de passe
	user.Password = ""
	return user, nil
}

// Login authentifie un utilisateur
func (s *authService) Login(dto LoginDTO) (string, *models.User, error) {
	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return "", nil, fmt.Errorf("données invalides: %v", validationErrors)
	}

	// Trouver l'utilisateur par email ou username
	var user *models.User
	var err error

	// Tenter d'abord par email
	if utils.ValidateEmail(dto.Identifier) {
		user, err = s.userRepo.FindByEmail(dto.Identifier)
	} else {
		user, err = s.userRepo.FindByUsername(dto.Identifier)
	}

	if err != nil {
		return "", nil, utils.ErrInvalidCredentials
	}

	// Vérifier le mot de passe
	if !s.ValidatePassword(user, dto.Password) {
		return "", nil, utils.ErrInvalidCredentials
	}

	// Générer le token JWT
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("erreur génération token: %w", err)
	}

	// Mettre à jour la dernière connexion
	if err := s.userRepo.UpdateLastConnection(user.ID); err != nil {
		// Log l'erreur mais ne pas faire échouer la connexion
		fmt.Printf("Erreur mise à jour dernière connexion: %v\n", err)
	}

	// Retourner le token et l'utilisateur sans mot de passe
	user.Password = ""
	return token, user, nil
}

// ValidatePassword vérifie si un mot de passe correspond au hash
func (s *authService) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// GenerateToken génère un token JWT pour un utilisateur
func (s *authService) GenerateToken(user *models.User) (string, error) {
	// Définir l'expiration
	expirationTime := time.Now().Add(time.Duration(s.config.JWT.ExpirationHours) * time.Hour)

	// Créer les claims
	claims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rythmit-api",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Créer le token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signer le token
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", fmt.Errorf("erreur signature token: %w", err)
	}

	return tokenString, nil
}

// ParseToken analyse et valide un token JWT
func (s *authService) ParseToken(tokenString string) (*CustomClaims, error) {
	// Parser le token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, utils.ErrTokenExpired
		}
		return nil, utils.ErrTokenInvalid
	}

	// Vérifier que le token est valide
	if !token.Valid {
		return nil, utils.ErrTokenInvalid
	}

	// Extraire les claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, utils.ErrTokenInvalid
	}

	return claims, nil
}

// RefreshToken génère un nouveau token à partir d'un ancien (si pas expiré)
func (s *authService) RefreshToken(oldToken string) (string, error) {
	// Parser l'ancien token
	claims, err := s.ParseToken(oldToken)
	if err != nil {
		return "", err
	}

	// Récupérer l'utilisateur pour s'assurer qu'il existe encore
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", utils.ErrUserNotFound
	}

	// Générer un nouveau token
	return s.GenerateToken(user)
}

// hashPassword hache un mot de passe avec bcrypt
func (s *authService) hashPassword(password string) (string, error) {
	cost := s.config.Security.BcryptCost
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ToUserResponseDTO convertit un User en UserResponseDTO
func ToUserResponseDTO(user *models.User) UserResponseDTO {
	return UserResponseDTO{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		IsAdmin:        user.IsAdmin,
		ProfilePic:     user.ProfilePic,
		Biography:      user.Biography,
		LastConnection: user.LastConnection,
		MessageCount:   user.MessageCount,
		ThreadCount:    user.ThreadCount,
		CreatedAt:      user.CreatedAt,
	}
}
