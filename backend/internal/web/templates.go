package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"rythmitbackend/internal/utils"

	"github.com/gorilla/mux"
)

// TemplateManager gère le chargement et le rendu des templates
type TemplateManager struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
	funcMap   template.FuncMap
}

// NewTemplateManager crée une nouvelle instance de TemplateManager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*template.Template),
		funcMap: template.FuncMap{
			"add":         func(a, b int) int { return a + b },
			"sub":         func(a, b int) int { return a - b },
			"eq":          func(a, b interface{}) bool { return a == b },
			"ne":          func(a, b interface{}) bool { return a != b },
			"gt":          func(a, b int) bool { return a > b },
			"lt":          func(a, b int) bool { return a < b },
			"initials":    formatInitials,
			"formatDate":  formatTimeAgo,
			"joinStrings": strings.Join,
			"pluck":       pluck,
		},
	}
}

// formatInitials génère les initiales à partir d'un nom d'utilisateur
func formatInitials(username string) string {
	if len(username) == 0 {
		return "?"
	}
	// Prend les deux premières lettres majuscules, ou juste la première si une seule
	runes := []rune(username)
	if len(runes) >= 2 {
		return strings.ToUpper(string(runes[0])) + strings.ToUpper(string(runes[1]))
	}
	return strings.ToUpper(string(runes[0]))
}

// formatTimeAgo formate une date en "il y a X [unité]"
func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff.Hours() < 24 && time.Now().Day() == t.Day() {
		if diff.Minutes() < 1 {
			return "À l'instant"
		} else if diff.Minutes() < 60 {
			return fmt.Sprintf("il y a %.0f m", diff.Minutes())
		} else {
			return fmt.Sprintf("il y a %.0f h", diff.Hours())
		}
	} else if diff.Hours() < 24*7 && time.Now().Year() == t.Year() {
		return fmt.Sprintf("il y a %.0f j", diff.Hours()/24)
	} else if diff.Hours() < 24*30 { // Approximation d'un mois
		return fmt.Sprintf("il y a %.0f sem", diff.Hours()/24/7)
	} else if diff.Hours() < 24*365 { // Approximation d'un an
		return fmt.Sprintf("il y a %.0f mois", diff.Hours()/24/30)
	}
	return fmt.Sprintf("il y a %.0f ans", diff.Hours()/24/365)
}

// pluck extrait les valeurs d'un champ donné d'une slice de structs
func pluck(data interface{}, field string) ([]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("pluck requires a slice, got %T", data)
	}

	var results []interface{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem() // Déréférencer si c'est un pointeur
		}
		if elem.Kind() != reflect.Struct {
			return nil, fmt.Errorf("pluck requires a slice of structs or pointers to structs, got %T in slice", elem.Interface())
		}

		fieldValue := elem.FieldByName(field)
		if !fieldValue.IsValid() {
			// Gérer le cas où le champ n'existe pas ou n'est pas exporté
			// On pourrait retourner une erreur ou juste ignorer
			return nil, fmt.Errorf("champ \"%s\" non trouvé ou non exporté dans le struct %T", field, elem.Interface())
		}

		results = append(results, fieldValue.Interface())
	}

	return results, nil
}

// LoadTemplates charge tous les templates depuis le dossier templates
func (tm *TemplateManager) LoadTemplates() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Templates de base
	layouts, err := filepath.Glob("templates/layouts/*.html")
	if err != nil {
		return err
	}

	// Templates de pages
	pages, err := filepath.Glob("templates/pages/*.html")
	if err != nil {
		return err
	}

	// Charger chaque page avec les layouts
	for _, page := range pages {
		name := filepath.Base(page)
		files := append(layouts, page)
		tmpl, err := template.New(name).Funcs(tm.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("erreur chargement template %s: %w", name, err)
		}
		tm.templates[name] = tmpl
	}

	return nil
}

// RenderTemplate rend un template avec les données fournies
func (tm *TemplateManager) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tm.mu.RLock()
	tmpl, ok := tm.templates[name]
	tm.mu.RUnlock()

	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return fmt.Errorf("template %s not found", name)
	}

	// Exécuter le template avec le nom de la page
	return tmpl.ExecuteTemplate(w, name, data)
}

// PageData contient les données communes à toutes les pages
type PageData struct {
	Title       string
	Description string
	User        interface{}
	Data        interface{}
	Error       string
	Success     string
}

// BasePageData crée une nouvelle instance de PageData avec des valeurs par défaut
func BasePageData(title string) PageData {
	return PageData{
		Title:       title,
		Description: "Rythmit - La plateforme de battles de rap",
	}
}

// Handler est une interface pour les handlers de pages
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// TemplateHandler est un handler qui rend un template
type TemplateHandler struct {
	tmpl     *TemplateManager
	template string
	getData  func(r *http.Request) (PageData, error)
}

// NewTemplateHandler crée un nouveau TemplateHandler
func NewTemplateHandler(tmpl *TemplateManager, template string, getData func(r *http.Request) (PageData, error)) *TemplateHandler {
	return &TemplateHandler{
		tmpl:     tmpl,
		template: template,
		getData:  getData,
	}
}

// ServeHTTP implémente l'interface http.Handler
func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h.getData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.tmpl.RenderTemplate(w, h.template, data); err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// RequireAuth est un middleware qui vérifie l'authentification
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le token depuis le cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Vérifier le token JWT
		token, err := utils.ValidateJWTToken(cookie.Value)
		if err != nil {
			// Token invalide ou expiré
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		// Ajouter les claims du token au contexte
		ctx := context.WithValue(r.Context(), "user_id", token.Claims.(*utils.JWTClaims).UserID)
		ctx = context.WithValue(ctx, "is_admin", token.Claims.(*utils.JWTClaims).IsAdmin)

		// Continuer avec le handler suivant
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SetupWebRoutes configure les routes web
func SetupWebRoutes(r *mux.Router, tmpl *TemplateManager) {
	// Gestionnaire de fichiers statiques
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Routes publiques
	r.Handle("/", NewTemplateHandler(tmpl, "home.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Accueil"), nil
	}))

	r.Handle("/login", NewTemplateHandler(tmpl, "login.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Connexion"), nil
	}))

	r.Handle("/register", NewTemplateHandler(tmpl, "register.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Inscription"), nil
	}))

	// Routes protégées
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(RequireAuth)

	protected.Handle("/profile", NewTemplateHandler(tmpl, "profile.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Mon Profil"), nil
	}))

	protected.Handle("/threads", NewTemplateHandler(tmpl, "threads.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Threads"), nil
	}))

	protected.Handle("/battles", NewTemplateHandler(tmpl, "battles.html", func(r *http.Request) (PageData, error) {
		return BasePageData("Battles"), nil
	}))
}
