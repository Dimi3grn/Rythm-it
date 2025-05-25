# Makefile pour Rythmit Backend

# Variables
BINARY_NAME=rythmit
MAIN_PATH=cmd/server/main.go
DB_NAME=rythmit_db

# Couleurs pour output
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: help
help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Lance le serveur
	@echo "$(GREEN)🚀 Démarrage du serveur Rythmit...$(NC)"
	go run $(MAIN_PATH)

.PHONY: dev
dev: ## Lance le serveur avec hot reload (nécessite air)
	@echo "$(GREEN)🔄 Démarrage en mode développement avec hot reload...$(NC)"
	air

.PHONY: build
build: ## Compile l'application
	@echo "$(GREEN)🔨 Compilation de Rythmit...$(NC)"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✅ Binaire créé: bin/$(BINARY_NAME)$(NC)"

.PHONY: test
test: ## Lance les tests
	@echo "$(GREEN)🧪 Lancement des tests...$(NC)"
	go test -v ./...

.PHONY: test-cover
test-cover: ## Lance les tests avec couverture
	@echo "$(GREEN)📊 Tests avec couverture...$(NC)"
	go test -v -cover ./...

.PHONY: lint
lint: ## Lance le linter
	@echo "$(GREEN)🔍 Analyse du code...$(NC)"
	golangci-lint run

.PHONY: fmt
fmt: ## Formate le code
	@echo "$(GREEN)✨ Formatage du code...$(NC)"
	go fmt ./...

.PHONY: tidy
tidy: ## Nettoie les dépendances
	@echo "$(GREEN)🧹 Nettoyage des dépendances...$(NC)"
	go mod tidy

.PHONY: deps
deps: ## Installe les dépendances
	@echo "$(GREEN)📦 Installation des dépendances...$(NC)"
	go mod download

.PHONY: db-test
db-test: ## Teste la connexion DB
	@echo "$(GREEN)🗄️  Test de connexion MySQL...$(NC)"
	go run cmd/test_db/main.go

.PHONY: db-reset
db-reset: ## Reset la base de données (ATTENTION: efface tout!)
	@echo "$(RED)⚠️  ATTENTION: Cette commande va effacer toute la base de données!$(NC)"
	@read -p "Êtes-vous sûr? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		mysql -u root -e "DROP DATABASE IF EXISTS $(DB_NAME); CREATE DATABASE $(DB_NAME);" && \
		mysql -u root $(DB_NAME) < migrations/001_initial_schema.sql && \
		echo "$(GREEN)✅ Base de données réinitialisée$(NC)"; \
	fi

.PHONY: migrate
migrate: ## Applique les migrations
	@echo "$(GREEN)🔄 Application des migrations...$(NC)"
	mysql -u root $(DB_NAME) < migrations/001_initial_schema.sql
	@echo "$(GREEN)✅ Migrations appliquées$(NC)"

.PHONY: install-tools
install-tools: ## Installe les outils de développement
	@echo "$(GREEN)🛠️  Installation des outils de développement...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✅ Outils installés$(NC)"

.PHONY: clean
clean: ## Nettoie les fichiers générés
	@echo "$(GREEN)🧹 Nettoyage...$(NC)"
	rm -rf bin/
	rm -rf tmp/
	go clean

.PHONY: all
all: clean deps fmt lint test build ## Lance tout

# Commandes Docker (pour plus tard)
.PHONY: docker-build
docker-build: ## Construit l'image Docker
	@echo "$(GREEN)🐳 Construction de l'image Docker...$(NC)"
	docker build -t rythmit-backend .

.PHONY: docker-run
docker-run: ## Lance le conteneur Docker
	@echo "$(GREEN)🐳 Lancement du conteneur...$(NC)"
	docker run -p 8085:8085 --env-file .env rythmit-backend