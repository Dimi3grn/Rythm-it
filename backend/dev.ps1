param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Script de développement pour Windows PowerShell
# Forcer l'encodage UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# Couleurs
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Info { Write-Host $args -ForegroundColor Cyan }

# Vérifier que WAMP est lancé
function Test-Wamp {
    try {
        $mysql = Get-Process "mysqld" -ErrorAction SilentlyContinue
        if ($mysql) {
            return $true
        }
        return $false
    } catch {
        return $false
    }
}

# Commandes disponibles
switch ($Command) {
    "help" {
        Write-Info "Rythmit Backend - Commandes disponibles:"
        Write-Host ""
        Write-Success "SETUP:"
        Write-Host "  install   - Installation complete"
        Write-Host ""
        Write-Success "DEVELOPPEMENT:"
        Write-Host "  run       - Lance le serveur avec hot reload"
        Write-Host "  dev       - Alias pour run"
        Write-Host "  build     - Compile l'application"
        Write-Host "  test      - Lance les tests"
        Write-Host ""
        Write-Success "MAINTENANCE:"
        Write-Host "  db-test   - Teste la connexion MySQL"
        Write-Host "  deps      - Met a jour les dependances"
        Write-Host "  clean     - Nettoie les fichiers temporaires"
        Write-Host ""
        Write-Info "Pour un nouveau projet:"
        Write-Host "   1. .\dev.ps1 install"
        Write-Host "   2. Importer le fichier SQL dans phpMyAdmin"
        Write-Host "   3. .\dev.ps1 run"
        Write-Host ""
    }
    
    "run" {
        if (-not (Test-Wamp)) {
            Write-Error "ERREUR: WAMP n est pas lance! Demarre WAMP d abord."
            exit 1
        }
        
        # Vérifier que Air est installé
        $air = Get-Command air -ErrorAction SilentlyContinue
        if (-not $air) {
            Write-Error "ERREUR: Air n est pas installe. Lance .\dev.ps1 install d abord"
            exit 1
        }
        
        Write-Success "Demarrage du serveur Rythmit avec hot reload..."
        Write-Info "Serveur accessible sur: http://localhost:8085"
        Write-Info "Appuie sur Ctrl+C pour arreter"
        Write-Host ""
        air
    }
    
    "dev" {
        # Alias pour la commande "run"
        & $MyInvocation.MyCommand.Path "run"
    }
    
    "build" {
        Write-Success "OK: Compilation de Rythmit..."
        New-Item -ItemType Directory -Force -Path "bin" | Out-Null
        go build -o bin/rythmit.exe cmd/server/main.go
        Write-Success "OK: Binaire cree: bin/rythmit.exe"
    }
    
    "test" {
        Write-Success "OK: Lancement des tests..."
        go test -v ./...
    }
    
    "db-test" {
        if (-not (Test-Wamp)) {
            Write-Error "ERREUR: WAMP n est pas lance!"
            exit 1
        }
        Write-Success "OK: Test de connexion MySQL..."
        go run cmd/test_db/main.go
    }
    
    "deps" {
        Write-Success "OK: Mise a jour des dependances Go..."
        go mod download
        go mod tidy
     
        Write-Success "OK: Dependances mises a jour"
    }
    
    "clean" {
        Write-Success "OK: Nettoyage..."
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue bin
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue tmp
        go clean
    }
    
    "install" {
        Write-Success "🚀 Rythmit - Installation complete du projet..."
        Write-Host ""
        
        # Vérifier Go
        $go = Get-Command go -ErrorAction SilentlyContinue
        if (-not $go) {
            Write-Error "ERREUR: Go n est pas installe!"
            Write-Info "Installe Go depuis https://golang.org/dl/"
            exit 1
        }
        Write-Success "OK: Go detecte"
        
        # Installer les dépendances Go
        Write-Info "Installation des dependances Go..."
        go mod download
        go mod tidy
        Write-Success "OK: Dependances Go installees"
        
        # Installer les outils de développement
        Write-Info "Installation des outils de developpement..."
        go install github.com/cosmtrek/air@latest
        Write-Success "OK: Air installe pour le hot reload"
        
        # Créer les dossiers nécessaires
        Write-Info "Creation des dossiers necessaires..."
        New-Item -ItemType Directory -Force -Path "uploads" | Out-Null
        New-Item -ItemType Directory -Force -Path "uploads/profiles" | Out-Null
        New-Item -ItemType Directory -Force -Path "uploads/threads" | Out-Null
        New-Item -ItemType Directory -Force -Path "bin" | Out-Null
        New-Item -ItemType Directory -Force -Path "tmp" | Out-Null
        Write-Success "OK: Dossiers crees"
        
        # Vérifier/Créer le fichier .env
        if (-not (Test-Path ".env")) {
            Write-Info "Creation du fichier .env..."
            $env = "# Configuration Rythmit Backend`n"
            $env += "APP_NAME=Rythmit`n"
            $env += "APP_PORT=8085`n"
            $env += "APP_ENV=development`n`n"
            $env += "# Database MySQL`n"
            $env += "DB_HOST=localhost`n"
            $env += "DB_PORT=3306`n"
            $env += "DB_USER=root`n"
            $env += "DB_PASSWORD=`n"
            $env += "DB_NAME=rythmit_db`n`n"
            $env += "# JWT Configuration`n"
            $env += "JWT_SECRET=your-super-secret-jwt-key`n"
            $env += "JWT_EXPIRATION_HOURS=24`n`n"
            $env += "# Security`n"
            $env += "BCRYPT_COST=12`n"
            $env += "MIN_PASSWORD_LENGTH=8`n`n"
            $env += "# Upload Configuration`n"
            $env += "MAX_UPLOAD_SIZE=5242880`n"
            $env += "UPLOAD_PATH=uploads"
            $env | Out-File -FilePath ".env" -Encoding UTF8
            Write-Success "Fichier .env cree"
        } else {
            Write-Info "Fichier .env existe deja"
        }
        
        Write-Host ""
        Write-Success "Installation terminee!"
        Write-Host ""
        Write-Info "Prochaines etapes:"
        Write-Host "  1. Demarre WAMP/XAMPP pour MySQL"
        Write-Host "  2. Cree une base rythmit_db et importe le fichier SQL"
        Write-Host "  3. Lance le projet: .\dev.ps1 run"
        Write-Host ""
    }
    
    default {
        Write-Error "Commande inconnue: $Command"
        Write-Host "Utilise .\dev.ps1 help pour voir les commandes disponibles"
    }
}