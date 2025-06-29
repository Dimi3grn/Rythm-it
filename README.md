# 🎵 Rythmit - Forum Musical Interactif

Forum musical avec système de battles d'artistes, profils utilisateurs et interactions sociales.

## 🚀 Quick Start (Nouveau développeur)

Pour lancer le projet après avoir cloné le repo :

### 1. Prérequis
- **Go 1.21+** - [Télécharger ici](https://golang.org/dl/)
- **WAMP/XAMPP** - Pour MySQL ([WAMP](https://www.wampserver.com/) ou [XAMPP](https://www.apachefriends.org/))
- **Git** - [Télécharger ici](https://git-scm.com/)

### 2. Installation automatique

```powershell
# Aller dans le dossier backend
cd backend

# Installation complète (dépendances + outils + configuration)
.\dev.ps1 install

# Créer la base de données dans WAMP/phpMyAdmin
# - Créer une base 'rythmit_db'
# - Importer le fichier SQL du repo

# Lancer le serveur avec hot reload
.\dev.ps1 run
```

**C'est tout ! 🎉** Le site sera accessible sur `http://localhost:8085`

### 3. Commandes utiles

```powershell
.\dev.ps1 help      # Voir toutes les commandes
.\dev.ps1 run       # Lancer le serveur
.\dev.ps1 test      # Lancer les tests
.\dev.ps1 build     # Compiler l'application
```

---

## 📁 Structure du projet

```
Rythm-it/
├── backend/           # API Go avec Gorilla Mux
│   ├── cmd/          # Points d'entrée
│   ├── internal/     # Code métier
│   └── dev.ps1       # Script de développement
├── frontend/         # Interface HTML/CSS/JS
└── README.md         # Ce fichier
```

## 🛠️ Technologies

- **Backend**: Go, Gorilla Mux, MySQL, JWT
- **Frontend**: HTML5, CSS3, JavaScript Vanilla
- **Base de données**: MySQL
- **Outils**: Air (hot reload), PowerShell (dev scripts)

## 👥 Équipe

- **Dimitri** - Backend Go
- **Romain** - Frontend JS/HTML/CSS

---

**Version**: 0.2.0  
**Dernière mise à jour**: Décembre 2024 