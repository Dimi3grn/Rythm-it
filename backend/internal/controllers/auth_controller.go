package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"rythmitbackend/internal/services"
	"rythmitbackend/internal/utils"

	"github.com/gorilla/mux"
)

// AuthController gestionnaire des routes d'authentification
type AuthController struct {
	Controller
	authService services.AuthService
}

// NewAuthController crée une nouvelle instance du contrôleur d'authentification
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Routes retourne les routes gérées par ce contrôleur
func (ac *AuthController) Routes() []Route {
	return []Route{
		{
			Name:        "Register",
			Method:      "POST",
			Pattern:     "/register",
			HandlerFunc: ac.Register,
			Protected:   false,
		},
		{
			Name:        "Login",
			Method:      "POST",
			Pattern:     "/login",
			HandlerFunc: ac.Login,
			Protected:   false,
		},
		{
			Name:        "Profile",
			Method:      "GET",
			Pattern:     "/profile",
			HandlerFunc: ac.GetProfile,
			Protected:   true,
		},
		{
			Name:        "UpdateProfile",
			Method:      "PUT",
			Pattern:     "/profile",
			HandlerFunc: ac.UpdateProfile,
			Protected:   true,
		},
		{
			Name:        "RefreshToken",
			Method:      "POST",
			Pattern:     "/refresh",
			HandlerFunc: ac.RefreshToken,
			Protected:   false,
		},
	}
}

// Register gère l'inscription des nouveaux utilisateurs
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var dto services.RegisterDTO

	// Parser le JSON de la requête
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.BadRequest(w, "Format JSON invalide")
		return
	}

	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		utils.ValidationErrors(w, validationErrors)
		return
	}

	// Inscription via le service
	user, err := ac.authService.Register(dto)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	// Convertir en DTO de réponse
	userResponse := services.ToUserResponseDTO(user)

	// Retourner l'utilisateur créé
	utils.Created(w, "Inscription réussie", userResponse)
}

// Login gère la connexion des utilisateurs
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var dto services.LoginDTO

	// Parser le JSON de la requête
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.BadRequest(w, "Format JSON invalide")
		return
	}

	// Validation des données
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		utils.ValidationErrors(w, validationErrors)
		return
	}

	// Connexion via le service
	token, user, err := ac.authService.Login(dto)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	// Préparer la réponse avec token et utilisateur
	authResponse := services.AuthResponseDTO{
		Token: token,
		User:  services.ToUserResponseDTO(user),
	}

	// Retourner le token et les infos utilisateur
	utils.Success(w, "Connexion réussie", authResponse)
}

// GetProfile retourne le profil de l'utilisateur connecté
func (ac *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'utilisateur depuis le contexte (injecté par le middleware auth)
	user := GetUserFromContext(r)
	if user == nil {
		utils.Unauthorized(w, "Utilisateur non trouvé")
		return
	}

	utils.Success(w, "Profil récupéré", user)
}

// UpdateProfile met à jour le profil de l'utilisateur connecté
func (ac *AuthController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'utilisateur depuis le contexte
	currentUser := GetUserFromContext(r)
	if currentUser == nil {
		utils.Unauthorized(w, "Utilisateur non trouvé")
		return
	}

	// Structure pour les données de mise à jour
	var updateData struct {
		Username   *string `json:"username,omitempty" validate:"omitempty,min=3,max=30,username"`
		Email      *string `json:"email,omitempty" validate:"omitempty,email"`
		Biography  *string `json:"biography,omitempty" validate:"omitempty,max=500"`
		ProfilePic *string `json:"profile_pic,omitempty" validate:"omitempty,url"`
	}

	// Parser le JSON de la requête
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		utils.BadRequest(w, "Format JSON invalide")
		return
	}

	// Validation des données
	if validationErrors := utils.ValidateStruct(updateData); len(validationErrors) > 0 {
		utils.ValidationErrors(w, validationErrors)
		return
	}

	// Mettre à jour les champs modifiés
	if updateData.Username != nil {
		currentUser.Username = *updateData.Username
	}
	if updateData.Email != nil {
		currentUser.Email = *updateData.Email
	}
	if updateData.Biography != nil {
		currentUser.Biography = updateData.Biography
	}
	if updateData.ProfilePic != nil {
		currentUser.ProfilePic = updateData.ProfilePic
	}

	// TODO: Mettre à jour via le service/repository
	// Pour l'instant, on retourne les données sans sauvegarder

	utils.Success(w, "Profil mis à jour", currentUser)
}

// RefreshToken génère un nouveau token à partir d'un token existant
func (ac *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Token string `json:"token" validate:"required"`
	}

	// Parser le JSON de la requête
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		utils.BadRequest(w, "Format JSON invalide")
		return
	}

	// Validation des données
	if validationErrors := utils.ValidateStruct(requestData); len(validationErrors) > 0 {
		utils.ValidationErrors(w, validationErrors)
		return
	}

	// Rafraîchir le token via le service
	newToken, err := ac.authService.RefreshToken(requestData.Token)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	// Retourner le nouveau token
	response := map[string]string{
		"token": newToken,
	}

	utils.Success(w, "Token rafraîchi", response)
}

// Helper functions pour l'authentification

func GetUserFromContext(r *http.Request) *services.UserResponseDTO {
	if user, ok := r.Context().Value("user").(*services.UserResponseDTO); ok {
		return user
	}
	return nil
}

// GetUserIDFromContext récupère l'ID utilisateur depuis le contexte
func GetUserIDFromContext(r *http.Request) (uint, bool) {
	if userID, ok := r.Context().Value("user_id").(uint); ok {
		return userID, true
	}
	return 0, false
}

// IsAdminFromContext vérifie si l'utilisateur est admin
func IsAdminFromContext(r *http.Request) bool {
	if isAdmin, ok := r.Context().Value("is_admin").(bool); ok {
		return isAdmin
	}
	return false
}

// ParseIDFromURL extrait un ID depuis l'URL (helper pour les routes avec paramètres)
func ParseIDFromURL(r *http.Request, paramName string) (uint, error) {
	vars := mux.Vars(r)
	idStr, exists := vars[paramName]
	if !exists {
		return 0, utils.ErrInvalidInput
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, utils.ErrInvalidInput
	}

	return uint(id), nil
}

// RequireAdmin vérifie que l'utilisateur est admin (helper pour les controllers admin)
func RequireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !IsAdminFromContext(r) {
		utils.Forbidden(w, "Droits administrateur requis")
		return false
	}
	return true
}

// GetPaginationParams extrait les paramètres de pagination depuis la query string
func GetPaginationParams(r *http.Request) services.PaginationParams {
	params := services.DefaultPagination()

	// Page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Per page
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if perPage, err := strconv.Atoi(perPageStr); err == nil && perPage > 0 && perPage <= 100 {
			params.PerPage = perPage
		}
	}

	// Sort
	if sort := r.URL.Query().Get("sort"); sort != "" {
		params.Sort = sort
	}

	// Order
	if order := r.URL.Query().Get("order"); order == "ASC" || order == "DESC" {
		params.Order = order
	}

	services.ValidatePagination(&params)
	return params
}
