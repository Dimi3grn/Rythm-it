# Rythm'it - Contexte Claude

## 🎵 Vue d'ensemble du projet

**Nom** : Rythm'it  
**Type** : Forum musical interactif avec système de battles d'artistes  
**Équipe** : Dimitri (Backend Go) + Romain (Frontend JS/HTML/CSS)  
**Deadline** : 16/06/2025 23h59  
**Repo** : https://github.com/Dimi3grn/rythmit

## 🚀 Status actuel du projet

### ✅ Phase 0 COMPLÈTE (26/05/2025)
- Structure MVC complète avec Go
- Module : `rythmitbackend`
- Port : 8085
- Base de données MySQL via WAMP
- 15 tests unitaires qui passent
- Hot reload avec Air configuré
- Documentation complète

### 🔄 Phase 1 EN ATTENTE
Prochaine étape : Système d'authentification JWT

## 📂 Structure actuelle du projet

```
rythmit/
├── backend/
│   ├── cmd/
│   │   ├── server/
│   │   │   ├── main.go
│   │   │   └── main_test.go
│   │   └── test_db/
│   │       └── main.go
│   ├── configs/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   └── base_controller.go
│   │   ├── middleware/
│   │   │   └── middleware.go
│   │   ├── models/
│   │   │   └── base_model.go
│   │   ├── repositories/
│   │   │   └── repository.go
│   │   ├── router/
│   │   │   └── router.go
│   │   ├── services/
│   │   │   └── service.go
│   │   └── utils/
│   │       ├── errors.go
│   │       ├── response.go
│   │       └── validation.go
│   ├── pkg/
│   │   └── database/
│   │       ├── database.go
│   │       └── database_test.go
│   ├── migrations/
│   │   ├── 000_create_database.sql
│   │   └── 001_initial_schema.sql
│   ├── docs/
│   │   └── ARCHITECTURE.md
│   ├── .env
│   ├── .env.example
│   ├── .gitignore
│   ├── .air.toml
│   ├── dev.ps1
│   ├── Makefile
│   ├── go.mod
│   ├── go.sum
│   └── README.md
├── frontend/ (à venir - Romain)
├── database/ (scripts SQL dans backend/migrations)
└── claude_context.md (ce fichier)
```

## 🔧 Configuration technique actuelle

### Environnement de développement
- **OS** : Windows avec PowerShell
- **Base de données** : MySQL via WAMP (pas de mot de passe root)
- **Hot reload** : Air configuré et fonctionnel
- **Scripts** : `dev.ps1` pour Windows, Makefile pour Linux/Mac

### Dépendances installées
```go
// go.mod principal
module rythmitbackend

require:
- github.com/gorilla/mux         // Router HTTP
- github.com/rs/cors            // CORS middleware
- github.com/go-sql-driver/mysql // Driver MySQL
- github.com/joho/godotenv      // Variables d'environnement
- github.com/golang-jwt/jwt/v5  // JWT (installé, pas encore utilisé)
- golang.org/x/crypto/bcrypt    // Hash passwords (installé, pas encore utilisé)
- github.com/go-playground/validator/v10 // Validation (installé, utilisé dans utils)
```

### Routes implémentées
- ✅ `GET /` - Page d'accueil API
- ✅ `GET /health` - Health check direct
- ✅ `GET /api/health` - Health check API
- ✅ `GET /api/ready` - Readiness check avec status DB
- 🚧 Toutes les autres routes retournent 501 Not Implemented

### Base de données
- **11 tables créées** : users, threads, messages, tags, battles, etc.
- **Admin par défaut** : username=admin, password=ChangeThisPassword123!
- **20 tags musicaux** préchargés (genres + artistes)
- **Connexion** : Pool de 25 connexions max, timeouts configurés

## 🎯 Concept unique (rappel)

Forum musical où les utilisateurs peuvent :
- Créer des discussions sur des artistes/albums/genres
- Lancer des battles musicales (ex: Drake vs Kendrick)
- Voter avec Fire 🔥 (like) ou Skip ⏭️ (dislike)  
- Découvrir leur compatibilité musicale avec d'autres users
- Intégrer des embeds YouTube/Spotify dans les messages

## 📋 Spécifications techniques obligatoires

### Contraintes imposées
- **Backend** : Go obligatoire (architecture MVC) ✅
- **Frontend** : HTML/CSS/JavaScript (pas de framework imposé)
- **Auth** : JWT + hash SHA512 minimum pour mots de passe
- **BDD** : Persistance obligatoire (MySQL) ✅
- **Refresh** : JavaScript pour éviter les rechargements de page

### Fonctionnalités obligatoires (FT-1 à FT-12)
1. **FT-1** : Inscription (username unique, email unique, mdp 12+ chars + majuscule + spécial)
2. **FT-2** : Connexion (username/email + mdp → JWT)
3. **FT-3** : Création threads (titre, desc, tags, état: ouvert/fermé/archivé)
4. **FT-4** : Consultation threads (même non-auth)
5. **FT-5** : Messages dans threads (auth uniquement)
6. **FT-6** : Fire/Skip sur messages (pas les deux simultanément)
7. **FT-7** : CRUD propriétaire + admin peut tout supprimer
8. **FT-8** : Tri messages (chronologique OU popularité via Fire/Skip)
9. **FT-9** : Pagination (10/20/30/tout, défaut: 10)
10. **FT-10** : Filtrage par tags
11. **FT-11** : Recherche (titre OU tags)
12. **FT-12** : Dashboard admin (modérer, bannir, changer états)

## 🗄️ Structure Base de Données

### Tables actuelles (créées dans MySQL)
- `users` - Utilisateurs avec is_admin
- `threads` - Discussions avec état et visibilité
- `messages` - Messages dans les threads
- `tags` - Tags musicaux (genre/artist/album)
- `thread_tags` - Liaison threads-tags (N:N)
- `message_votes` - Fire/Skip sur messages
- `friendships` - Musical Twins
- `battles` - Battles musicales
- `battle_options` - Options de vote (2 par battle)
- `battle_votes` - Votes des utilisateurs
- `user_music_preferences` - Préférences musicales

## 🔗 API Endpoints prévus

Les endpoints sont définis dans `router.go` mais pas encore implémentés.

### Authentification
```
POST /api/public/register   - 501 Not Implemented
POST /api/public/login       - 501 Not Implemented
GET  /api/v1/profile        - 501 Not Implemented (auth requise)
```

### Threads
```
GET  /api/public/threads          - 501 Not Implemented
GET  /api/public/threads/{id}     - 501 Not Implemented
POST /api/v1/threads              - 501 Not Implemented (auth requise)
PUT  /api/v1/threads/{id}         - 501 Not Implemented (auth requise)
DELETE /api/v1/threads/{id}       - 501 Not Implemented (auth requise)
```

### Messages
```
GET  /api/v1/threads/{id}/messages - 501 Not Implemented (auth requise)
POST /api/v1/threads/{id}/messages - 501 Not Implemented (auth requise)
POST /api/v1/messages/{id}/fire    - 501 Not Implemented (auth requise)
POST /api/v1/messages/{id}/skip    - 501 Not Implemented (auth requise)
```

### Battles
```
GET  /api/public/battles/active  - 501 Not Implemented
GET  /api/public/battles/{id}    - 501 Not Implemented
POST /api/v1/battles             - 501 Not Implemented (auth requise)
POST /api/v1/battles/{id}/vote   - 501 Not Implemented (auth requise)
```

### Admin
```
GET  /api/v1/admin/dashboard         - 403 Forbidden (middleware admin actif)
POST /api/v1/admin/users/{id}/ban    - 403 Forbidden
PUT  /api/v1/admin/threads/{id}/state - 403 Forbidden
```

## 🔐 Sécurité et validation

### Validation mise en place
- Helper de validation dans `utils/validation.go`
- Validation custom pour passwords (12 chars, majuscule, spécial)
- Validation username (alphanumérique + underscore)
- Interface `validator/v10` intégrée

### Middlewares actifs
- ✅ Logger : Log toutes les requêtes
- ✅ Recovery : Récupère des panics
- ✅ CORS : Configuré pour localhost:3000 et 5173
- ✅ JSON : Force Content-Type JSON sur /api
- 🚧 Auth : Structure en place mais pas de vérification JWT
- ✅ Admin : Bloque tous les accès admin (403)

### Gestion des erreurs
- Erreurs définies dans `utils/errors.go`
- Réponses standardisées dans `utils/response.go`
- Codes d'erreur cohérents

## 🚀 Workflow de développement actuel

### Commandes disponibles
```powershell
# Windows PowerShell
.\dev.ps1 help    # Affiche l'aide
.\dev.ps1 dev     # Hot reload avec Air
.\dev.ps1 run     # Lance le serveur
.\dev.ps1 test    # Lance les tests
.\dev.ps1 build   # Compile l'exe
.\dev.ps1 db-test # Test connexion MySQL
```

### Git
- Branche actuelle : `feature/dimitri-backend`
- Commits réguliers avec messages conventionnels
- Phase 0 complète et pushée

### Tests
- 15 tests unitaires écrits et fonctionnels
- Coverage : cmd/server, configs, pkg/database
- Commande : `.\dev.ps1 test`

## 📊 Formats JSON (définis mais pas utilisés)

Les structures sont définies dans `models/base_model.go` mais pas encore utilisées dans les endpoints.

## ⚠️ Points d'attention pour la suite

### Phase 1 (Authentification) à implémenter
1. Repository User avec CRUD
2. Service d'authentification
3. Hash bcrypt (remplacer SHA512 du cahier des charges)
4. Génération et validation JWT
5. Endpoints register/login fonctionnels
6. Middleware Auth qui vérifie vraiment les tokens

### Décisions techniques prises
- Bcrypt au lieu de SHA512 (plus sécurisé)
- Validator v10 pour la validation
- Structure MVC stricte avec separation of concerns
- Tests first approach

### Points de vigilance
- Le middleware Auth laisse tout passer actuellement
- Pas de rate limiting
- Pas de gestion des sessions/refresh tokens
- Sanitisation XSS à implémenter

## 📝 Documentation

- `README.md` : Guide d'installation et utilisation
- `docs/ARCHITECTURE.md` : Documentation technique
- `claude_context.md` : Ce fichier (contexte IA)

---

**Dernière mise à jour** : 26/05/2025 - Phase 0 complète, prêt pour Phase 1 (Auth)
**Prochaine étape** : Implémenter le système d'authentification JWT complet