package services_test

import (
	"testing"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/internal/utils"
	"rythmitbackend/pkg/database"
)

func setupAuthService(t *testing.T) (services.AuthService, repositories.UserRepository) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Impossible de se connecter à la base de données: %v", err)
	}

	userRepo := repositories.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo, cfg)

	return authService, userRepo
}

func teardownAuthService() {
	database.Close()
}

func createValidRegisterDTO() services.RegisterDTO {
	return services.RegisterDTO{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPassword123!",
	}
}

func createValidLoginDTO() services.LoginDTO {
	return services.LoginDTO{
		Identifier: "test@example.com",
		Password:   "ValidPassword123!",
	}
}

func TestAuthService_Register(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	dto := createValidRegisterDTO()

	// Test inscription réussie
	user, err := authService.Register(dto)
	if err != nil {
		t.Fatalf("Erreur inscription: %v", err)
	}
	defer userRepo.Delete(user.ID)

	// Vérifications
	if user.Username != dto.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", dto.Username, user.Username)
	}

	if user.Email != dto.Email {
		t.Errorf("Email attendu: %s, obtenu: %s", dto.Email, user.Email)
	}

	if user.Password != "" {
		t.Error("Le mot de passe ne devrait pas être retourné")
	}

	if user.IsAdmin {
		t.Error("L'utilisateur ne devrait pas être admin par défaut")
	}

	// Vérifier que l'utilisateur est bien en base
	savedUser, err := userRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Utilisateur non trouvé en base: %v", err)
	}

	if savedUser.Password == "" {
		t.Error("Le mot de passe devrait être haché en base")
	}
}

func TestAuthService_Register_ValidationErrors(t *testing.T) {
	authService, _ := setupAuthService(t)
	defer teardownAuthService()

	tests := []struct {
		name string
		dto  services.RegisterDTO
	}{
		{
			name: "Username trop court",
			dto: services.RegisterDTO{
				Username: "ab",
				Email:    "test@example.com",
				Password: "ValidPassword123!",
			},
		},
		{
			name: "Email invalide",
			dto: services.RegisterDTO{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "ValidPassword123!",
			},
		},
		{
			name: "Mot de passe trop faible",
			dto: services.RegisterDTO{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authService.Register(tt.dto)
			if err == nil {
				t.Error("Devrait retourner une erreur de validation")
			}
		})
	}
}

func TestAuthService_Register_DuplicateConstraints(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	// Créer un premier utilisateur
	dto1 := createValidRegisterDTO()
	user1, err := authService.Register(dto1)
	if err != nil {
		t.Fatalf("Erreur création premier utilisateur: %v", err)
	}
	defer userRepo.Delete(user1.ID)

	// Tenter de créer un utilisateur avec le même email
	dto2 := services.RegisterDTO{
		Username: "differentuser",
		Email:    dto1.Email, // Même email
		Password: "ValidPassword123!",
	}

	_, err = authService.Register(dto2)
	if err == nil {
		t.Error("Devrait échouer avec un email dupliqué")
	}
	if err != utils.ErrEmailAlreadyUsed {
		t.Errorf("Erreur attendue: %v, obtenue: %v", utils.ErrEmailAlreadyUsed, err)
	}

	// Tenter de créer un utilisateur avec le même username
	dto3 := services.RegisterDTO{
		Username: dto1.Username, // Même username
		Email:    "different@example.com",
		Password: "ValidPassword123!",
	}

	_, err = authService.Register(dto3)
	if err == nil {
		t.Error("Devrait échouer avec un username dupliqué")
	}
	if err != utils.ErrUsernameTaken {
		t.Errorf("Erreur attendue: %v, obtenue: %v", utils.ErrUsernameTaken, err)
	}
}

func TestAuthService_Login(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	// Créer un utilisateur
	registerDTO := createValidRegisterDTO()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	// Test connexion avec email
	loginDTO := services.LoginDTO{
		Identifier: registerDTO.Email,
		Password:   registerDTO.Password,
	}

	token, loggedUser, err := authService.Login(loginDTO)
	if err != nil {
		t.Fatalf("Erreur connexion: %v", err)
	}

	// Vérifications
	if token == "" {
		t.Error("Le token ne devrait pas être vide")
	}

	if loggedUser.ID != user.ID {
		t.Errorf("ID utilisateur attendu: %d, obtenu: %d", user.ID, loggedUser.ID)
	}

	if loggedUser.Password != "" {
		t.Error("Le mot de passe ne devrait pas être retourné")
	}

	// Test connexion avec username
	loginDTO.Identifier = registerDTO.Username
	token2, _, err := authService.Login(loginDTO)
	if err != nil {
		t.Fatalf("Erreur connexion avec username: %v", err)
	}

	if token2 == "" {
		t.Error("Le token ne devrait pas être vide")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	// Créer un utilisateur
	registerDTO := createValidRegisterDTO()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	tests := []struct {
		name string
		dto  services.LoginDTO
	}{
		{
			name: "Mauvais mot de passe",
			dto: services.LoginDTO{
				Identifier: registerDTO.Email,
				Password:   "WrongPassword123!",
			},
		},
		{
			name: "Email inexistant",
			dto: services.LoginDTO{
				Identifier: "inexistant@example.com",
				Password:   registerDTO.Password,
			},
		},
		{
			name: "Username inexistant",
			dto: services.LoginDTO{
				Identifier: "inexistant",
				Password:   registerDTO.Password,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := authService.Login(tt.dto)
			if err == nil {
				t.Error("Devrait échouer avec des identifiants invalides")
			}
			if err != utils.ErrInvalidCredentials {
				t.Errorf("Erreur attendue: %v, obtenue: %v", utils.ErrInvalidCredentials, err)
			}
		})
	}
}

func TestAuthService_ValidatePassword(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	// Créer un utilisateur
	registerDTO := createValidRegisterDTO()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	// Récupérer l'utilisateur avec le mot de passe haché
	savedUser, err := userRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Erreur récupération utilisateur: %v", err)
	}

	// Test mot de passe correct
	isValid := authService.ValidatePassword(savedUser, registerDTO.Password)
	if !isValid {
		t.Error("Le mot de passe devrait être valide")
	}

	// Test mot de passe incorrect
	isValid = authService.ValidatePassword(savedUser, "WrongPassword")
	if isValid {
		t.Error("Le mot de passe ne devrait pas être valide")
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	authService, _ := setupAuthService(t)
	defer teardownAuthService()

	user := &models.User{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	token, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Erreur génération token: %v", err)
	}

	if token == "" {
		t.Error("Le token ne devrait pas être vide")
	}

	// Vérifier que le token peut être parsé
	claims, err := authService.ParseToken(token)
	if err != nil {
		t.Fatalf("Erreur parsing token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID attendu: %d, obtenu: %d", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", user.Username, claims.Username)
	}

	if claims.Email != user.Email {
		t.Errorf("Email attendu: %s, obtenu: %s", user.Email, claims.Email)
	}

	if claims.IsAdmin != user.IsAdmin {
		t.Errorf("IsAdmin attendu: %t, obtenu: %t", user.IsAdmin, claims.IsAdmin)
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	authService, _ := setupAuthService(t)
	defer teardownAuthService()

	user := &models.User{
		BaseModel: models.BaseModel{
			ID: 1,
		},
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  true,
	}

	// Générer un token valide
	token, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Erreur génération token: %v", err)
	}

	// Test parsing token valide
	claims, err := authService.ParseToken(token)
	if err != nil {
		t.Fatalf("Erreur parsing token valide: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID attendu: %d, obtenu: %d", user.ID, claims.UserID)
	}

	// Test token invalide
	_, err = authService.ParseToken("invalid.token.here")
	if err == nil {
		t.Error("Devrait échouer avec un token invalide")
	}

	// Test token vide
	_, err = authService.ParseToken("")
	if err == nil {
		t.Error("Devrait échouer avec un token vide")
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	authService, userRepo := setupAuthService(t)
	defer teardownAuthService()

	// Créer un utilisateur
	registerDTO := createValidRegisterDTO()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	// Générer un token initial
	originalToken, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Erreur génération token initial: %v", err)
	}

	// Attendre un peu pour s'assurer que les timestamps diffèrent
	time.Sleep(time.Second * 1)

	// Rafraîchir le token
	newToken, err := authService.RefreshToken(originalToken)
	if err != nil {
		t.Fatalf("Erreur refresh token: %v", err)
	}

	if newToken == "" {
		t.Error("Le nouveau token ne devrait pas être vide")
	}

	if newToken == originalToken {
		t.Error("Le nouveau token devrait être différent de l'original")
	}

	// Vérifier que le nouveau token est valide
	claims, err := authService.ParseToken(newToken)
	if err != nil {
		t.Fatalf("Erreur parsing nouveau token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID attendu: %d, obtenu: %d", user.ID, claims.UserID)
	}
}

func TestToUserResponseDTO(t *testing.T) {
	now := time.Now()
	user := &models.User{
		BaseModel: models.BaseModel{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       "shouldnotbeincluded",
		IsAdmin:        true,
		MessageCount:   5,
		ThreadCount:    3,
		LastConnection: &now,
	}

	dto := services.ToUserResponseDTO(user)

	// Vérifications
	if dto.ID != user.ID {
		t.Errorf("ID attendu: %d, obtenu: %d", user.ID, dto.ID)
	}

	if dto.Username != user.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", user.Username, dto.Username)
	}

	if dto.Email != user.Email {
		t.Errorf("Email attendu: %s, obtenu: %s", user.Email, dto.Email)
	}

	if dto.IsAdmin != user.IsAdmin {
		t.Errorf("IsAdmin attendu: %t, obtenu: %t", user.IsAdmin, dto.IsAdmin)
	}

	if dto.MessageCount != user.MessageCount {
		t.Errorf("MessageCount attendu: %d, obtenu: %d", user.MessageCount, dto.MessageCount)
	}

	if dto.ThreadCount != user.ThreadCount {
		t.Errorf("ThreadCount attendu: %d, obtenu: %d", user.ThreadCount, dto.ThreadCount)
	}

	if dto.CreatedAt != user.CreatedAt {
		t.Errorf("CreatedAt ne correspond pas")
	}
}
