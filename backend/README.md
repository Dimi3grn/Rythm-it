# Rythmit Backend

Forum musical interactif avec système de battles d'artistes - Backend Go

## 🚀 Installation et lancement

### Prérequis
- Go 1.21+
- MySQL 5.7+ (ou WAMP/XAMPP)
- Git

### Installation

1. **Cloner le repository**
```bash
git clone https://github.com/Dimi3grn/rythmit.git
cd rythmit/backend
```

2. **Installer les dépendances**
```bash
go mod download
```

3. **Configuration**
```bash
# Copier le fichier d'exemple
cp .env.example .env

# Éditer .env avec vos paramètres MySQL
```

4. **Base de données**
```bash
# Créer la base de données
mysql -u root -p < migrations/001_initial_schema.sql
```

### Lancement

**Mode développement (avec hot reload)**
```bash
# Windows PowerShell
.\dev.ps1 dev

# Linux/Mac
make dev
```

**Mode production**
```bash
# Windows
.\dev.ps1 run

# Linux/Mac
make run
```

## 📁 Structure du projet

```
backend/
├── cmd/
│   └── server/         # Point d'entrée principal
├── configs/            # Configuration
├── internal/
│   ├── controllers/    # Logique métier
│   ├── middleware/     # Middlewares (auth, CORS, logging)
│   ├── models/         # Structures de données
│   ├── repositories/   # Accès base de données
│   ├── router/         # Routes HTTP
│   ├── services/       # Services métier
│   └── utils/          # Utilitaires
├── pkg/
│   └── database/       # Connexion MySQL
├── migrations/         # Scripts SQL
└── tests/             # Tests

```

## 🔗 Routes API

### Routes publiques

| Méthode | Route | Description | Status |
|---------|-------|-------------|---------|
| GET | `/` | Page d'accueil API | ✅ |
| GET | `/health` | Health check | ✅ |
| GET | `/api/health` | Health check API | ✅ |
| GET | `/api/ready` | Readiness check | ✅ |
| POST | `/api/public/register` | Inscription | 🚧 |
| POST | `/api/public/login` | Connexion | 🚧 |
| GET | `/api/public/threads` | Liste des threads publics | 🚧 |

### Routes protégées (Auth JWT requise)

| Méthode | Route | Description | Status |
|---------|-------|-------------|---------|
| GET | `/api/v1/profile` | Profil utilisateur | 🚧 |
| POST | `/api/v1/threads` | Créer un thread | 🚧 |
| POST | `/api/v1/messages/{id}/fire` | Fire un message | 🚧 |
| POST | `/api/v1/messages/{id}/skip` | Skip un message | 🚧 |

### Routes admin

| Méthode | Route | Description | Status |
|---------|-------|-------------|---------|
| GET | `/api/v1/admin/dashboard` | Dashboard admin | 🚧 |
| POST | `/api/v1/admin/users/{id}/ban` | Bannir un utilisateur | 🚧 |

## 🧪 Tests

```bash
# Lancer tous les tests
.\dev.ps1 test

# Tests avec couverture
go test -cover ./...

# Tests sans base de données
go test -short ./...
```

## 🛠️ Commandes de développement

### Windows (PowerShell)
```bash
.\dev.ps1 help     # Affiche l'aide
.\dev.ps1 dev      # Mode développement avec hot reload
.\dev.ps1 build    # Compile l'application
.\dev.ps1 test     # Lance les tests
.\dev.ps1 db-test  # Teste la connexion MySQL
```

### Linux/Mac (Make)
```bash
make help          # Affiche l'aide
make dev           # Mode développement
make build         # Compile
make test          # Tests
```

## 🔐 Configuration

Variables d'environnement principales (`.env`) :

```env
# Application
APP_NAME=Rythmit
APP_PORT=8085

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=rythmit_db

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24

# Security
BCRYPT_COST=12
MIN_PASSWORD_LENGTH=12
```

## 👥 Équipe

- **Dimitri** - Backend Go
- **Romain** - Frontend JS/HTML/CSS

## 📝 Conventions de code

- Architecture MVC stricte
- Noms en anglais
- Tests pour chaque fonctionnalité
- Commits conventionnels (feat:, fix:, docs:, etc.)

## 🚧 Roadmap

- [x] Phase 0 : Architecture et infrastructure
- [ ] Phase 1 : Système d'authentification JWT
- [ ] Phase 2 : CRUD Threads et messages
- [ ] Phase 3 : Système Fire/Skip
- [ ] Phase 4 : Battle system
- [ ] Phase 5 : Musical Twins

---

**Version** : 0.1.0  
**Dernière mise à jour** : 26/05/2025