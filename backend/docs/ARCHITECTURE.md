# Architecture Backend Rythmit

## Vue d'ensemble

Le backend Rythmit suit une architecture MVC (Model-View-Controller) stricte avec une séparation claire des responsabilités.

## Flux de données

```
Client HTTP → Router → Middleware → Controller → Service → Repository → Database
                ↓                        ↓            ↓           ↓
              Response ← Utils ← ← ← ← Model ← ← ← ← ← ← ← ← ← SQL
```

## Couches de l'application

### 1. Router (`internal/router`)
- Définit toutes les routes HTTP
- Applique les middlewares globaux (CORS, Logger, Recovery)
- Distribue les requêtes vers les controllers

### 2. Middleware (`internal/middleware`)
- **Logger** : Log toutes les requêtes HTTP
- **Recovery** : Récupère des panics
- **Auth** : Vérifie les tokens JWT
- **Admin** : Vérifie les droits administrateur
- **CORS** : Gère les requêtes cross-origin

### 3. Controllers (`internal/controllers`)
- Gèrent les requêtes HTTP
- Valident les entrées
- Appellent les services
- Formatent les réponses

### 4. Services (`internal/services`)
- Contiennent la logique métier
- Orchestrent les appels aux repositories
- Gèrent les transactions

### 5. Repositories (`internal/repositories`)
- Accès aux données (CRUD)
- Requêtes SQL
- Aucune logique métier

### 6. Models (`internal/models`)
- Structures de données
- Tags de validation
- Constantes métier

## Patterns utilisés

### Singleton pour la configuration
```go
var instance *Config

func Load() *Config {
    if instance != nil {
        return instance
    }
    // ...
}
```

### Repository Pattern
```go
type UserRepository interface {
    Create(user *models.User) error
    FindByID(id uint) (*models.User, error)
    FindByEmail(email string) (*models.User, error)
}
```

### Dependency Injection
```go
type UserService struct {
    repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

## Sécurité

### Authentification
- JWT pour les tokens d'authentification
- Bcrypt pour le hashing des mots de passe
- Tokens avec expiration configurable

### Validation
- Validation des entrées avec `go-playground/validator`
- Sanitisation des données utilisateur
- Protection contre les injections SQL

### Middlewares de sécurité
- Rate limiting (à implémenter)
- CORS configuré
- Headers de sécurité

## Base de données

### Structure
- 11 tables principales
- Relations many-to-many pour les tags
- Index sur les colonnes fréquemment recherchées

### Transactions
```go
func (s *Service) CreateWithTags(thread *models.Thread, tags []string) error {
    return s.repo.Transaction(func(tx *sql.Tx) error {
        // Créer le thread
        // Associer les tags
        // Tout est rollback si erreur
    })
}
```

## Configuration

### Variables d'environnement
- Chargées via `godotenv`
- Valeurs par défaut sécurisées
- Configuration typée

### Gestion des erreurs
- Erreurs custom définies
- Codes d'erreur standardisés
- Réponses JSON cohérentes

## Tests

### Structure des tests
- Tests unitaires par package
- Tests d'intégration pour les endpoints
- Mocks pour les dépendances externes

### Coverage cible
- Controllers : 80%+
- Services : 90%+
- Repositories : 70%+
- Utils : 100%

## Performance

### Optimisations
- Pool de connexions MySQL (25 max)
- Timeouts configurés
- Pagination obligatoire

### Monitoring (à venir)
- Métriques Prometheus
- Logs structurés
- Health checks

## Déploiement

### Environnements
- **Development** : Hot reload, logs debug
- **Production** : Optimisé, logs JSON

### Docker (à venir)
- Multi-stage build
- Image Alpine minimale
- Health checks intégrés 