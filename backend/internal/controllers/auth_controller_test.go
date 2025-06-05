package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"
)

func setupAuthController(t *testing.T) (*AuthController, services.AuthService, repositories.UserRepository) {
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
	authController := NewAuthController(authService)

	return authController, authService, userRepo
}

func teardownAuthController() {
	database.Close()
}

func createValidRegisterRequest() services.RegisterDTO {
	return services.RegisterDTO{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPassword123!",
	}
}

func createValidLoginRequest() services.LoginDTO {
	return services.LoginDTO{
		Identifier: "test@example.com",
		Password:   "ValidPassword123!",
	}
}

func TestNewAuthController(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	if authController == nil {
		t.Error("NewAuthController ne devrait pas retourner nil")
	}

	// Vérifier que les routes sont définies
	routes := authController.Routes()
	if len(routes) == 0 {
		t.Error("AuthController devrait avoir des routes définies")
	}

	// Vérifier les routes principales
	routeNames := make(map[string]bool)
	for _, route := range routes {
		routeNames[route.Name] = true
	}

	expectedRoutes := []string{"Register", "Login", "Profile", "UpdateProfile", "RefreshToken"}
	for _, expected := range expectedRoutes {
		if !routeNames[expected] {
			t.Errorf("Route manquante: %s", expected)
		}
	}
}

func TestAuthController_Register(t *testing.T) {
	authController, _, userRepo := setupAuthController(t)
	defer teardownAuthController()

	registerDTO := createValidRegisterRequest()

	// Convertir en JSON
	jsonData, err := json.Marshal(registerDTO)
	if err != nil {
		t.Fatalf("Erreur marshalling JSON: %v", err)
	}

	// Créer la requête
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Erreur création requête: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Créer le ResponseRecorder
	rr := httptest.NewRecorder()

	// Exécuter la requête
	authController.Register(rr, req)

	// Vérifier le status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusCreated)
	}

	// Vérifier le contenu de la réponse
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Erreur parsing réponse JSON: %v", err)
	}

	// Vérifier que la réponse contient les champs attendus
	if !response["success"].(bool) {
		t.Error("La réponse devrait indiquer un succès")
	}

	if response["message"] != "Inscription réussie" {
		t.Errorf("Message incorrect: got %v", response["message"])
	}

	// Vérifier que l'utilisateur a été créé en base
	user, err := userRepo.FindByEmail(registerDTO.Email)
	if err != nil {
		t.Fatalf("Utilisateur non trouvé en base: %v", err)
	}

	if user.Username != registerDTO.Username {
		t.Errorf("Username incorrect en base: got %v want %v", user.Username, registerDTO.Username)
	}

	// Nettoyer
	userRepo.Delete(user.ID)
}

func TestAuthController_Register_ValidationErrors(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	tests := []struct {
		name string
		dto  services.RegisterDTO
		want int
	}{
		{
			name: "Username trop court",
			dto: services.RegisterDTO{
				Username: "ab",
				Email:    "test@example.com",
				Password: "ValidPassword123!",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Email invalide",
			dto: services.RegisterDTO{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "ValidPassword123!",
			},
			want: http.StatusBadRequest,
		},
		{
			name: "Mot de passe trop faible",
			dto: services.RegisterDTO{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "weak",
			},
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.dto)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			authController.Register(rr, req)

			if status := rr.Code; status != tt.want {
				t.Errorf("Status code incorrect: got %v want %v", status, tt.want)
			}
		})
	}
}

func TestAuthController_Register_InvalidJSON(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	// Envoyer du JSON invalide
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authController.Register(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code incorrect pour JSON invalide: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestAuthController_Login(t *testing.T) {
	authController, authService, userRepo := setupAuthController(t)
	defer teardownAuthController()

	// Créer d'abord un utilisateur
	registerDTO := createValidRegisterRequest()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	// Tenter de se connecter
	loginDTO := createValidLoginRequest()
	jsonData, _ := json.Marshal(loginDTO)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authController.Login(rr, req)

	// Vérifier le status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusOK)
	}

	// Vérifier le contenu de la réponse
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Erreur parsing réponse JSON: %v", err)
	}

	// Vérifier que la réponse contient un token
	data := response["data"].(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Error("La réponse devrait contenir un token")
	}

	// Vérifier que la réponse contient les infos utilisateur
	userInfo := data["user"].(map[string]interface{})
	if userInfo["username"] != registerDTO.Username {
		t.Errorf("Username incorrect dans réponse: got %v want %v", userInfo["username"], registerDTO.Username)
	}
}

func TestAuthController_Login_InvalidCredentials(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	// Tenter de se connecter avec des credentials invalides
	loginDTO := services.LoginDTO{
		Identifier: "nonexistent@example.com",
		Password:   "WrongPassword123!",
	}

	jsonData, _ := json.Marshal(loginDTO)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authController.Login(rr, req)

	// Devrait retourner une erreur d'authentification
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthController_RefreshToken(t *testing.T) {
	authController, authService, userRepo := setupAuthController(t)
	defer teardownAuthController()

	// Créer un utilisateur et obtenir un token
	registerDTO := createValidRegisterRequest()
	user, err := authService.Register(registerDTO)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer userRepo.Delete(user.ID)

	loginDTO := createValidLoginRequest()
	token, _, err := authService.Login(loginDTO)
	if err != nil {
		t.Fatalf("Erreur login: %v", err)
	}

	// Attendre un peu pour que le nouveau token soit différent
	time.Sleep(time.Second * 1)

	// Rafraîchir le token
	requestData := map[string]string{"token": token}
	jsonData, _ := json.Marshal(requestData)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authController.RefreshToken(rr, req)

	// Vérifier le status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusOK)
	}

	// Vérifier que la réponse contient un nouveau token
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Erreur parsing réponse JSON: %v", err)
	}

	data := response["data"].(map[string]interface{})
	newToken := data["token"].(string)

	if newToken == "" {
		t.Error("Le nouveau token ne devrait pas être vide")
	}

	if newToken == token {
		t.Error("Le nouveau token devrait être différent de l'ancien")
	}
}

func TestAuthController_GetProfile(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	// Cette fonction ne peut pas être testée complètement car GetUserFromContext retourne nil
	// Une fois le middleware auth implémenté, on pourra tester correctement

	req, _ := http.NewRequest("GET", "/profile", nil)
	rr := httptest.NewRecorder()

	authController.GetProfile(rr, req)

	// Pour l'instant, devrait retourner Unauthorized car pas d'utilisateur dans le contexte
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthController_UpdateProfile(t *testing.T) {
	authController, _, _ := setupAuthController(t)
	defer teardownAuthController()

	// Comme pour GetProfile, ne peut pas être testé complètement sans middleware auth

	updateData := map[string]string{
		"username":  "newusername",
		"biography": "New bio",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	authController.UpdateProfile(rr, req)

	// Devrait retourner Unauthorized car pas d'utilisateur dans le contexte
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Status code incorrect: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestParseIDFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		paramName string
		want      uint
		wantErr   bool
	}{
		{
			name:      "ID valide",
			url:       "/users/123",
			paramName: "id",
			want:      123,
			wantErr:   false,
		},
		{
			name:      "ID invalide",
			url:       "/users/abc",
			paramName: "id",
			want:      0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Pour tester correctement, il faudrait utiliser mux.SetURLVars
			// ou créer un vrai router. Pour l'instant, ce test est incomplet.

			// Cette fonction sera testée plus complètement quand on aura le router intégré
		})
	}
}

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		queryStr string
		want     models.PaginationParams
	}{
		{
			name:     "Paramètres par défaut",
			queryStr: "",
			want: models.PaginationParams{
				Page:    1,
				PerPage: 10,
				Sort:    "id",
				Order:   "DESC",
			},
		},
		{
			name:     "Paramètres personnalisés",
			queryStr: "?page=2&per_page=20&sort=name&order=ASC",
			want: models.PaginationParams{
				Page:    2,
				PerPage: 20,
				Sort:    "name",
				Order:   "ASC",
			},
		},
		{
			name:     "Paramètres invalides ignorés",
			queryStr: "?page=-1&per_page=1000&order=INVALID",
			want: models.PaginationParams{
				Page:    1,
				PerPage: 10,
				Sort:    "id",
				Order:   "DESC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test"+tt.queryStr, nil)

			params := GetPaginationParams(req)

			if params.Page != tt.want.Page {
				t.Errorf("Page: got %v want %v", params.Page, tt.want.Page)
			}
			if params.PerPage != tt.want.PerPage {
				t.Errorf("PerPage: got %v want %v", params.PerPage, tt.want.PerPage)
			}
			if params.Sort != tt.want.Sort {
				t.Errorf("Sort: got %v want %v", params.Sort, tt.want.Sort)
			}
			if params.Order != tt.want.Order {
				t.Errorf("Order: got %v want %v", params.Order, tt.want.Order)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test GetUserFromContext (retourne nil pour l'instant)
	req, _ := http.NewRequest("GET", "/test", nil)
	user := GetUserFromContext(req)
	if user != nil {
		t.Error("GetUserFromContext devrait retourner nil avant implémentation du middleware")
	}

	// Test GetUserIDFromContext
	userID, exists := GetUserIDFromContext(req)
	if exists || userID != 0 {
		t.Error("GetUserIDFromContext devrait retourner (0, false) avant implémentation du middleware")
	}

	// Test IsAdminFromContext
	isAdmin := IsAdminFromContext(req)
	if isAdmin {
		t.Error("IsAdminFromContext devrait retourner false avant implémentation du middleware")
	}
}
