# 📷 Fonctionnalité d'Upload d'Images

Cette fonctionnalité permet aux utilisateurs d'ajouter des images à leurs threads et commentaires sur Rythm'it.

## ✨ Fonctionnalités

### 🎯 Images dans les Threads
- **Upload d'images** lors de la création d'un nouveau thread
- **Prévisualisation** de l'image avant publication
- **Affichage** des images dans le feed principal
- **Support** des formats : JPG, PNG, GIF, WebP
- **Limite de taille** : 5MB maximum

### 💬 Images dans les Commentaires
- **Upload d'images** dans les commentaires des threads
- **Prévisualisation** avant envoi du commentaire
- **Affichage** des images dans les commentaires
- **Même contraintes** que pour les threads

## 🛠️ Implémentation Technique

### Backend (Go)
```
📁 backend/
├── internal/
│   ├── models/base_model.go          # Ajout champ ImageURL aux Thread & Message
│   ├── services/thread_service.go    # Support des images dans les DTOs
│   └── handlers/
│       ├── upload_handler.go         # Gestion upload avec type (thread/profile)
│       └── page_handler.go           # Traitement des images dans formulaires
├── migrations/
│   ├── 004_add_image_to_threads.sql  # Migration pour field image_url dans threads
│   └── 005_add_image_to_messages.sql # Migration pour field image_url dans messages
└── uploads/
    ├── threads/                      # Stockage images de threads
    └── profiles/                     # Stockage images de profils
```

### Frontend (HTML/CSS/JS)
```
📁 frontend/
├── styles/
│   ├── css/index.css                 # Styles pour images et prévisualisations
│   └── js/image-upload.js            # Logique d'upload et prévisualisation
├── index.html                        # Affichage images dans feed + script upload
└── thread.html                       # Affichage images dans threads + script upload
```

## 🚀 Comment Utiliser

### Pour les Utilisateurs

1. **Créer un thread avec image :**
   - Allez sur la page d'accueil
   - Cliquez sur le bouton 📷 dans le formulaire
   - Sélectionnez votre image (max 5MB)
   - Une prévisualisation apparaît
   - Écrivez votre message et publiez

2. **Ajouter une image à un commentaire :**
   - Ouvrez un thread
   - Dans le formulaire de commentaire, cliquez sur 📷
   - Sélectionnez votre image
   - Écrivez votre commentaire et envoyez

3. **Supprimer une image avant publication :**
   - Cliquez sur le bouton ❌ dans la prévisualisation

### Pour les Développeurs

1. **Exécuter les migrations :**
   ```bash
   cd backend
   go run cmd/main.go migrate
   ```

2. **Tester l'upload :**
   ```bash
   cd backend
   chmod +x test_upload.sh
   ./test_upload.sh
   ```

3. **Structure des répertoires :**
   - Les images sont stockées dans `backend/uploads/threads/`
   - URLs publiques : `/uploads/threads/FILENAME`

## 🔧 Configuration

### Limites d'Upload
```go
// Dans upload_handler.go
const MAX_FILE_SIZE = 5 << 20  // 5MB
const MAX_FORM_SIZE = 10 << 20 // 10MB

// Types acceptés
var ALLOWED_TYPES = []string{
    "image/jpeg",
    "image/jpg", 
    "image/png",
    "image/gif",
    "image/webp",
}
```

### Sécurité
- ✅ Vérification de l'authentification
- ✅ Validation du type de fichier
- ✅ Limite de taille de fichier
- ✅ Génération de noms uniques
- ✅ Dossiers séparés par type

## 🎨 Interface Utilisateur

### Design
- **Bouton 📷** : Stylé avec hover effects
- **Prévisualisation** : Bordure pointillée avec bouton de suppression
- **Images publiées** : Coins arrondis avec ombre légère
- **Responsive** : Adaptation mobile automatique

### Feedback Utilisateur
- **Notifications** : Messages de succès/erreur en temps réel
- **États de chargement** : Indicateur ⏳ pendant l'upload
- **Bouton actif** : Indication visuelle quand une image est attachée

## 📝 API

### Endpoint d'Upload
```
POST /upload/image
Content-Type: multipart/form-data

Paramètres:
- image: fichier image (obligatoire)
- type: "thread" ou "profile" (optionnel, défaut: thread)

Réponse:
{
    "success": true,
    "imageUrl": "/uploads/threads/1735134567_a1b2c3d4e5f6.jpg",
    "message": "Image uploadée avec succès"
}
```

### Champs de Formulaire
```html
<!-- Thread -->
<input type="hidden" name="image_url" value="/uploads/threads/...">

<!-- Commentaire -->  
<input type="hidden" name="comment_image_url" value="/uploads/threads/...">
```

## 🐛 Dépannage

### Problèmes Courants

1. **Erreur "Dossier non trouvé"**
   ```bash
   mkdir -p backend/uploads/threads
   chmod 755 backend/uploads/threads
   ```

2. **Images non affichées**
   - Vérifier que le serveur sert `/uploads/` dans le router
   - Vérifier les permissions des fichiers

3. **Upload échoue**
   - Vérifier la taille du fichier (< 5MB)
   - Vérifier le type MIME
   - Vérifier l'authentification

### Debug
```bash
# Logs du serveur
go run cmd/main.go 

# Vérifier les uploads
ls -la backend/uploads/threads/
```

## 🔄 Évolutions Futures

- [ ] Compression automatique des images
- [ ] Support du drag & drop
- [ ] Galerie d'images multiples
- [ ] Recadrage/rotation d'images
- [ ] Intégration avec CDN
- [ ] Génération de miniatures

---

✅ **Fonctionnalité complète et prête à l'utilisation !** 