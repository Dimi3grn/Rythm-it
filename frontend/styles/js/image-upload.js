// Gestionnaire d'upload d'images pour les threads et commentaires
class ImageUploader {
    constructor() {
        this.setupImageUpload();
        this.setupCommentImageUpload();
    }

    setupImageUpload() {
        // Créer un input file caché pour les threads
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = 'image/*';
        fileInput.style.display = 'none';
        fileInput.id = 'image-upload-input';
        document.body.appendChild(fileInput);

        // Vérifier si le champ caché existe déjà, sinon le créer
        let imageUrlInput = document.getElementById('image-url-input');
        if (!imageUrlInput) {
            imageUrlInput = document.createElement('input');
            imageUrlInput.type = 'hidden';
            imageUrlInput.name = 'image_url';
            imageUrlInput.id = 'image-url-input';

            // Ajouter le champ caché au formulaire principal
            const form = document.querySelector('.composer');
            if (form) {
                form.appendChild(imageUrlInput);
            }
        }

        // Gestionnaire pour le bouton d'image dans le formulaire principal
        const imageButton = document.querySelector('.composer .tool-btn[title="Joindre une image"]');
        if (imageButton) {
            imageButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.openImageSelector('thread');
            });
        }

        // Gestionnaire pour la sélection de fichier
        fileInput.addEventListener('change', (e) => {
            this.handleImageSelection(e, 'thread');
        });

        // DEBUG: Ajouter un listener sur le formulaire pour vérifier les données avant envoi
        const form = document.querySelector('.composer');
        if (form) {
            form.addEventListener('submit', (e) => {
                const imageUrlInput = document.getElementById('image-url-input');
                console.log('🔍 DEBUG - Formulaire soumis avec:');
                console.log('  - image_url value:', imageUrlInput ? imageUrlInput.value : 'champ non trouvé');
                console.log('  - FormData contient:');
                const formData = new FormData(form);
                for (let [key, value] of formData.entries()) {
                    console.log(`    ${key}: ${value}`);
                }
            });
        }
    }

    setupCommentImageUpload() {
        // Créer un input file caché pour les commentaires
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = 'image/*';
        fileInput.style.display = 'none';
        fileInput.id = 'comment-image-upload-input';
        document.body.appendChild(fileInput);

        // Ajouter un champ caché pour l'URL de l'image des commentaires
        const imageUrlInput = document.createElement('input');
        imageUrlInput.type = 'hidden';
        imageUrlInput.name = 'comment_image_url';
        imageUrlInput.id = 'comment-image-url-input';

        // Ajouter le champ caché au formulaire de commentaire
        const commentForm = document.querySelector('#comment-form');
        if (commentForm) {
            commentForm.appendChild(imageUrlInput);
        }

        // Gestionnaire pour le bouton d'image dans les commentaires
        const commentImageButton = document.querySelector('#comment-form .tool-btn[title="Joindre une image"]');
        if (commentImageButton) {
            commentImageButton.addEventListener('click', (e) => {
                e.preventDefault();
                this.openImageSelector('comment');
            });
        }

        // Gestionnaire pour la sélection de fichier des commentaires
        fileInput.addEventListener('change', (e) => {
            this.handleImageSelection(e, 'comment');
        });
    }

    openImageSelector(type = 'thread') {
        const inputId = type === 'comment' ? 'comment-image-upload-input' : 'image-upload-input';
        const fileInput = document.getElementById(inputId);
        fileInput.click();
    }

    async handleImageSelection(event, type = 'thread') {
        const file = event.target.files[0];
        if (!file) return;

        // Vérifier le type de fichier
        if (!file.type.startsWith('image/')) {
            this.showNotification('❌ Veuillez sélectionner une image valide', 'error');
            return;
        }

        // Vérifier la taille (5MB max)
        if (file.size > 5 * 1024 * 1024) {
            this.showNotification('❌ L\'image est trop volumineuse (max 5MB)', 'error');
            return;
        }

        // Afficher l'état de chargement
        this.showUploadProgress(true, type);

        try {
            // Créer FormData
            const formData = new FormData();
            formData.append('image', file);
            formData.append('type', type); // Indiquer le type d'upload

            // Envoyer l'image au serveur
            const response = await fetch('/upload/image', {
                method: 'POST',
                body: formData
            });

            const result = await response.json();

            if (response.ok && result.success) {
                // Succès - mettre à jour l'interface
                this.handleImageUploadSuccess(result.imageUrl, type);
                this.showNotification('✅ Image uploadée avec succès !', 'success');
            } else {
                throw new Error(result.message || 'Erreur lors de l\'upload');
            }
        } catch (error) {
            console.error('Erreur upload:', error);
            this.showNotification('❌ Erreur lors de l\'upload: ' + error.message, 'error');
        } finally {
            this.showUploadProgress(false, type);
        }
    }

    handleImageUploadSuccess(imageUrl, type = 'thread') {
        // Mettre à jour le champ caché avec l'URL de l'image
        const inputId = type === 'comment' ? 'comment-image-url-input' : 'image-url-input';
        const imageUrlInput = document.getElementById(inputId);
        if (imageUrlInput) {
            imageUrlInput.value = imageUrl;
            
            // DEBUG: Vérifier que la valeur est bien mise à jour
            console.log('🔍 DEBUG - Image upload success:');
            console.log('  - imageUrl:', imageUrl);
            console.log('  - inputId:', inputId);
            console.log('  - imageUrlInput found:', !!imageUrlInput);
            console.log('  - imageUrlInput.value:', imageUrlInput.value);
        } else {
            console.error('❌ DEBUG - Champ caché non trouvé:', inputId);
        }

        // Afficher une prévisualisation de l'image
        this.showImagePreview(imageUrl, type);

        // Marquer le bouton comme actif
        const buttonSelector = type === 'comment' 
            ? '#comment-form .tool-btn[title="Joindre une image"]'
            : '.composer .tool-btn[title="Joindre une image"]';
        const imageButton = document.querySelector(buttonSelector);
        if (imageButton) {
            imageButton.classList.add('active');
            imageButton.style.backgroundColor = '#4CAF50';
            imageButton.style.color = 'white';
        }
    }

    showImagePreview(imageUrl, type = 'thread') {
        // Supprimer toute prévisualisation existante de ce type
        const previewClass = type === 'comment' ? 'comment-image-preview' : 'image-preview';
        const existingPreview = document.querySelector(`.${previewClass}`);
        if (existingPreview) {
            existingPreview.remove();
        }

        // Créer la prévisualisation
        const preview = document.createElement('div');
        preview.className = previewClass;
        preview.innerHTML = `
            <div style="margin: 10px 0; padding: 10px; border: 2px dashed #667eea; border-radius: 8px; position: relative;">
                <img src="${imageUrl}" alt="Image attachée" style="max-width: 200px; max-height: 150px; border-radius: 6px; object-fit: cover;">
                <button type="button" class="remove-image-btn" style="position: absolute; top: 5px; right: 5px; background: #ff4757; color: white; border: none; border-radius: 50%; width: 24px; height: 24px; font-size: 14px; cursor: pointer;">×</button>
                <p style="margin: 5px 0 0 0; font-size: 12px; color: #666;">Image attachée</p>
            </div>
        `;

        // Ajouter le gestionnaire de suppression
        const removeButton = preview.querySelector('.remove-image-btn');
        removeButton.addEventListener('click', () => {
            this.removeImage(type);
        });

        // Insérer la prévisualisation
        const targetSelector = type === 'comment' ? '.comment-input' : '.composer-input';
        const targetInput = document.querySelector(targetSelector);
        if (targetInput) {
            targetInput.parentNode.insertBefore(preview, targetInput.nextSibling);
        }
    }

    removeImage(type = 'thread') {
        // Supprimer l'URL du champ caché
        const inputId = type === 'comment' ? 'comment-image-url-input' : 'image-url-input';
        const imageUrlInput = document.getElementById(inputId);
        if (imageUrlInput) {
            imageUrlInput.value = '';
        }

        // Supprimer la prévisualisation
        const previewClass = type === 'comment' ? 'comment-image-preview' : 'image-preview';
        const preview = document.querySelector(`.${previewClass}`);
        if (preview) {
            preview.remove();
        }

        // Réinitialiser l'apparence du bouton
        const buttonSelector = type === 'comment' 
            ? '#comment-form .tool-btn[title="Joindre une image"]'
            : '.composer .tool-btn[title="Joindre une image"]';
        const imageButton = document.querySelector(buttonSelector);
        if (imageButton) {
            imageButton.classList.remove('active');
            imageButton.style.backgroundColor = '';
            imageButton.style.color = '';
        }

        // Réinitialiser l'input file
        const fileInputId = type === 'comment' ? 'comment-image-upload-input' : 'image-upload-input';
        const fileInput = document.getElementById(fileInputId);
        if (fileInput) {
            fileInput.value = '';
        }
    }

    showUploadProgress(show, type = 'thread') {
        const buttonSelector = type === 'comment' 
            ? '#comment-form .tool-btn[title="Joindre une image"]'
            : '.composer .tool-btn[title="Joindre une image"]';
        const imageButton = document.querySelector(buttonSelector);
        if (imageButton) {
            if (show) {
                imageButton.style.opacity = '0.6';
                imageButton.textContent = '⏳';
            } else {
                imageButton.style.opacity = '1';
                imageButton.textContent = '📷';
            }
        }
    }

    showNotification(message, type) {
        // Créer ou mettre à jour la notification
        let notification = document.querySelector('.upload-notification');
        if (!notification) {
            notification = document.createElement('div');
            notification.className = 'upload-notification';
            notification.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                padding: 12px 20px;
                border-radius: 6px;
                font-weight: 500;
                z-index: 10000;
                max-width: 300px;
                word-wrap: break-word;
                box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            `;
            document.body.appendChild(notification);
        }

        // Définir le style selon le type
        if (type === 'success') {
            notification.style.backgroundColor = '#4CAF50';
            notification.style.color = 'white';
        } else if (type === 'error') {
            notification.style.backgroundColor = '#f44336';
            notification.style.color = 'white';
        }

        notification.textContent = message;
        notification.style.display = 'block';

        // Masquer après 3 secondes
        setTimeout(() => {
            if (notification) {
                notification.style.display = 'none';
            }
        }, 3000);
    }
}

// Initialiser l'uploader d'images quand le DOM est prêt
document.addEventListener('DOMContentLoaded', () => {
    new ImageUploader();
});

// Également l'initialiser si le DOM est déjà prêt
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        new ImageUploader();
    });
} else {
    new ImageUploader();
} 