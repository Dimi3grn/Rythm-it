package utils

import (
	"errors"
	"fmt"
	"net/http"
)

// Erreurs communes de l'application
var (
	// Erreurs d'authentification
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrUserNotFound       = errors.New("utilisateur non trouvé")
	ErrUserAlreadyExists  = errors.New("cet utilisateur existe déjà")
	ErrEmailAlreadyUsed   = errors.New("cette adresse email est déjà utilisée")
	ErrTokenInvalid       = errors.New("token invalide")
	ErrTokenExpired       = errors.New("token expiré")
	ErrUnauthorized       = errors.New("non autorisé")

	// Erreurs de validation
	ErrInvalidInput    = errors.New("données invalides")
	ErrPasswordTooWeak = errors.New("mot de passe trop faible")
	ErrUsernameTaken   = errors.New("nom d'utilisateur déjà pris")

	// Erreurs de threads
	ErrThreadNotFound = errors.New("thread non trouvé")
	ErrThreadClosed   = errors.New("ce thread est fermé")
	ErrThreadArchived = errors.New("ce thread est archivé")

	// Erreurs de messages
	ErrMessageNotFound = errors.New("message non trouvé")
	ErrAlreadyVoted    = errors.New("vous avez déjà voté pour ce message")

	// Erreurs de battles
	ErrBattleNotFound     = errors.New("battle non trouvée")
	ErrBattleEnded        = errors.New("cette battle est terminée")
	ErrAlreadyVotedBattle = errors.New("vous avez déjà voté dans cette battle")

	// Erreurs système
	ErrDatabaseConnection = errors.New("erreur de connexion à la base de données")
	ErrInternalServer     = errors.New("erreur interne du serveur")
)

// AppError structure d'erreur personnalisée
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implémente l'interface error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError crée une nouvelle erreur d'application
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Codes d'erreur standards
const (
	// Auth errors
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenInvalid       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeUnauthorized       = "UNAUTHORIZED"

	// Validation errors
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeDuplicate  = "DUPLICATE_ENTRY"

	// Resource errors
	ErrCodeNotFound  = "NOT_FOUND"
	ErrCodeForbidden = "FORBIDDEN"

	// System errors
	ErrCodeDatabase = "DATABASE_ERROR"
	ErrCodeInternal = "INTERNAL_ERROR"
)

// HandleError gère les erreurs et retourne la réponse appropriée
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		Unauthorized(w, "Email ou mot de passe incorrect")
	case errors.Is(err, ErrUserNotFound):
		NotFound(w, "Utilisateur non trouvé")
	case errors.Is(err, ErrTokenInvalid):
		Unauthorized(w, "Token invalide")
	case errors.Is(err, ErrTokenExpired):
		Unauthorized(w, "Votre session a expiré")
	case errors.Is(err, ErrThreadNotFound):
		NotFound(w, "Thread non trouvé")
	case errors.Is(err, ErrThreadClosed):
		Forbidden(w, "Ce thread est fermé aux nouveaux messages")
	case errors.Is(err, ErrThreadArchived):
		Forbidden(w, "Ce thread est archivé")
	case errors.Is(err, ErrAlreadyVoted):
		BadRequest(w, "Vous avez déjà voté pour ce message")
	case errors.Is(err, ErrBattleEnded):
		BadRequest(w, "Cette battle est terminée")
	default:
		// Log l'erreur réelle pour debug
		// TODO: Ajouter un logger
		InternalServerError(w, "Une erreur est survenue")
	}
}
