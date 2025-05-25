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
        Write-Info "[Rythmit Backend] - Commandes disponibles:"
        Write-Host ""
        Write-Host "  run       - Lance le serveur"
        Write-Host "  dev       - Lance avec hot reload (necessite air)"
        Write-Host "  build     - Compile l'application"
        Write-Host "  test      - Lance les tests"
        Write-Host "  db-test   - Teste la connexion MySQL"
        Write-Host "  deps      - Installe les dependances"
        Write-Host "  clean     - Nettoie les fichiers temporaires"
        Write-Host "  install   - Installe les outils de dev"
        Write-Host ""
    }
    
    "run" {
        if (-not (Test-Wamp)) {
            Write-Error "[!] WAMP n'est pas lance! Demarre WAMP d'abord."
            exit 1
        }
        Write-Success "[OK] Demarrage du serveur Rythmit..."
        go run cmd/server/main.go
    }
    
    "dev" {
        if (-not (Test-Wamp)) {
            Write-Error "[!] WAMP n'est pas lance! Demarre WAMP d'abord."
            exit 1
        }
        
        $air = Get-Command air -ErrorAction SilentlyContinue
        if (-not $air) {
            Write-Error "Air n'est pas installe. Lance './dev.ps1 install' d'abord"
            exit 1
        }
        
        Write-Success "[OK] Mode developpement avec hot reload..."
        air
    }
    
    "build" {
        Write-Success "[OK] Compilation de Rythmit..."
        New-Item -ItemType Directory -Force -Path "bin" | Out-Null
        go build -o bin/rythmit.exe cmd/server/main.go
        Write-Success "[OK] Binaire cree: bin/rythmit.exe"
    }
    
    "test" {
        Write-Success "[OK] Lancement des tests..."
        go test -v ./...
    }
    
    "db-test" {
        if (-not (Test-Wamp)) {
            Write-Error "[!] WAMP n'est pas lance! Demarre WAMP d'abord."
            exit 1
        }
        Write-Success "[OK] Test de connexion MySQL..."
        go run cmd/test_db/main.go
    }
    
    "deps" {
        Write-Success "[OK] Installation des dependances..."
        go mod download
        go mod tidy
    }
    
    "clean" {
        Write-Success "[OK] Nettoyage..."
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue bin
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue tmp
        go clean
    }
    
    "install" {
        Write-Success "[OK] Installation des outils de developpement..."
        go install github.com/cosmtrek/air@latest
        Write-Success "[OK] Air installe pour le hot reload"
        
        Write-Info "[Info] Note: Assure-toi que $env:GOPATH\bin est dans ton PATH"
    }
    
    default {
        Write-Error "Commande inconnue: $Command"
        Write-Host "Utilise './dev.ps1 help' pour voir les commandes disponibles"
    }
}