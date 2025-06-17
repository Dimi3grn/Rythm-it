// JavaScript pour l'édition de thread
document.addEventListener('DOMContentLoaded', function() {
    console.log('Edit Thread JS loaded');
    
    // Initialiser tous les composants
    initImageUpload();
    initTagSelector();
    initFormValidation();
    
    // Ajouter les événements pour les boutons d'image
    window.changeImage = changeImage;
    window.removeImage = removeImage;
    window.removeTag = removeTag;
});

// ===== GESTION DES IMAGES =====
function initImageUpload() {
    const fileInput = document.getElementById('image-upload-input');
    const imageUrlInput = document.getElementById('image-url-input');
    
    if (!fileInput) return;
    
    fileInput.addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (!file) return;
        
        // Validation du fichier
        if (!validateImageFile(file)) {
            return;
        }
        
        // Upload du fichier
        uploadImage(file);
    });
}

function validateImageFile(file) {
    const maxSize = 5 * 1024 * 1024; // 5MB
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    
    if (!allowedTypes.includes(file.type)) {
        showNotification('Format d\'image non supporté. Utilisez JPG, PNG, GIF ou WebP.', 'error');
        return false;
    }
    
    if (file.size > maxSize) {
        showNotification('L\'image est trop voluminuse. Taille maximum: 5MB.', 'error');
        return false;
    }
    
    return true;
}

function uploadImage(file) {
    const formData = new FormData();
    formData.append('image', file);
    formData.append('type', 'thread'); // Spécifier le type d'upload
    
    // Afficher un indicateur de chargement
    showUploadProgress();
    
    fetch('/upload/image', {
        method: 'POST',
        body: formData,
        credentials: 'same-origin'
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Erreur upload: ' + response.status);
        }
        return response.json();
    })
    .then(data => {
        if (data.success && data.imageUrl) {
            // Mettre à jour l'interface avec la nouvelle image
            updateImageDisplay(data.imageUrl);
            
            // Mettre à jour le champ caché
            const imageUrlInput = document.getElementById('image-url-input');
            if (imageUrlInput) {
                imageUrlInput.value = data.imageUrl;
            }
            
            showNotification('Image téléchargée avec succès !', 'success');
        } else {
            throw new Error(data.message || 'Erreur inconnue');
        }
    })
    .catch(error => {
        console.error('Erreur upload image:', error);
        showNotification('Erreur lors du téléchargement de l\'image: ' + error.message, 'error');
    })
    .finally(() => {
        hideUploadProgress();
    });
}

function updateImageDisplay(imageUrl) {
    const imageContainer = document.querySelector('.current-image-container, .upload-placeholder');
    if (!imageContainer) return;
    
    // Remplacer le contenu par la nouvelle image
    imageContainer.outerHTML = `
        <div class="current-image-container">
            <img src="${imageUrl}" alt="Image du thread" class="current-thread-image">
            <div class="image-overlay">
                <button type="button" class="image-btn change-btn" onclick="changeImage()">
                    🔄 Changer
                </button>
                <button type="button" class="image-btn remove-btn" onclick="removeImage()">
                    🗑️ Supprimer
                </button>
            </div>
        </div>
    `;
}

function changeImage() {
    const fileInput = document.getElementById('image-upload-input');
    if (fileInput) {
        fileInput.click();
    }
}

function removeImage() {
    if (!confirm('Êtes-vous sûr de vouloir supprimer cette image ?')) {
        return;
    }
    
    // Remplacer l'image par l'interface d'upload
    const imageContainer = document.querySelector('.current-image-container');
    if (imageContainer) {
        imageContainer.outerHTML = `
            <div class="upload-placeholder" onclick="document.getElementById('image-upload-input').click()">
                <div class="upload-icon">📷</div>
                <div class="upload-text">Cliquez pour ajouter une image</div>
                <div class="upload-hint">JPG, PNG, GIF - Max 5MB</div>
            </div>
        `;
    }
    
    // Vider le champ caché
    const imageUrlInput = document.getElementById('image-url-input');
    if (imageUrlInput) {
        imageUrlInput.value = '';
    }
    
    // Réinitialiser l'input file
    const fileInput = document.getElementById('image-upload-input');
    if (fileInput) {
        fileInput.value = '';
    }
    
    showNotification('Image supprimée', 'success');
}

function showUploadProgress() {
    const placeholder = document.querySelector('.upload-placeholder, .current-image-container');
    if (placeholder) {
        placeholder.style.opacity = '0.5';
        placeholder.style.pointerEvents = 'none';
        
        // Ajouter un indicateur de chargement
        if (!document.querySelector('.upload-progress')) {
            const progress = document.createElement('div');
            progress.className = 'upload-progress';
            progress.innerHTML = `
                <div style="text-align: center; padding: 20px;">
                    <div class="loading-spinner" style="margin: 0 auto 10px; width: 30px; height: 30px; border: 3px solid #f3f3f3; border-top: 3px solid var(--accent-primary); border-radius: 50%; animation: spin 1s linear infinite;"></div>
                    <div>Téléchargement en cours...</div>
                </div>
            `;
            placeholder.parentNode.appendChild(progress);
        }
    }
}

function hideUploadProgress() {
    const placeholder = document.querySelector('.upload-placeholder, .current-image-container');
    if (placeholder) {
        placeholder.style.opacity = '1';
        placeholder.style.pointerEvents = 'auto';
    }
    
    const progress = document.querySelector('.upload-progress');
    if (progress) {
        progress.remove();
    }
}

// ===== GESTION DES TAGS =====
let availableTags = [];
let selectedTags = [];

function initTagSelector() {
    const tagInput = document.getElementById('tag-input');
    const selectedTagsContainer = document.getElementById('selected-tags');
    const tagsHiddenInput = document.getElementById('tags-hidden');
    
    if (!tagInput) return;
    
    // Charger les tags existants depuis l'interface
    const existingTags = selectedTagsContainer.querySelectorAll('.selected-tag');
    existingTags.forEach(tag => {
        const tagName = tag.getAttribute('data-tag');
        if (tagName && !selectedTags.includes(tagName)) {
            selectedTags.push(tagName);
        }
    });
    
    // Charger les tags disponibles
    loadAvailableTags();
    
    // Événements
    tagInput.addEventListener('input', handleTagInput);
    tagInput.addEventListener('keydown', handleTagKeydown);
    tagInput.addEventListener('blur', hideTagSuggestions);
}

function loadAvailableTags() {
    fetch('/api/public/tags', {
        method: 'GET',
        credentials: 'same-origin'
    })
    .then(response => response.json())
    .then(data => {
        if (data.success && data.data) {
            availableTags = data.data.map(tag => ({
                name: tag.name,
                type: tag.type
            }));
        }
    })
    .catch(error => {
        console.error('Erreur chargement tags:', error);
    });
}

function handleTagInput(e) {
    const query = e.target.value.trim().toLowerCase();
    
    if (query.length < 1) {
        hideTagSuggestions();
        return;
    }
    
    const suggestions = availableTags
        .filter(tag => tag.name.toLowerCase().includes(query) && !selectedTags.includes(tag.name))
        .slice(0, 8);
    
    showTagSuggestions(suggestions, query);
}

function handleTagKeydown(e) {
    const suggestions = document.querySelector('.tag-suggestions');
    if (!suggestions || !suggestions.classList.contains('show')) return;
    
    const items = suggestions.querySelectorAll('.tag-suggestion');
    const current = suggestions.querySelector('.tag-suggestion.selected');
    
    switch (e.key) {
        case 'ArrowDown':
            e.preventDefault();
            if (current) {
                current.classList.remove('selected');
                const next = current.nextElementSibling || items[0];
                next.classList.add('selected');
            } else if (items.length > 0) {
                items[0].classList.add('selected');
            }
            break;
            
        case 'ArrowUp':
            e.preventDefault();
            if (current) {
                current.classList.remove('selected');
                const prev = current.previousElementSibling || items[items.length - 1];
                prev.classList.add('selected');
            } else if (items.length > 0) {
                items[items.length - 1].classList.add('selected');
            }
            break;
            
        case 'Enter':
            e.preventDefault();
            if (current) {
                const tagName = current.getAttribute('data-tag');
                addTag(tagName);
                e.target.value = '';
                hideTagSuggestions();
            }
            break;
            
        case 'Escape':
            hideTagSuggestions();
            break;
    }
}

function showTagSuggestions(suggestions, query) {
    const suggestionsContainer = document.getElementById('tag-suggestions');
    if (!suggestionsContainer) return;
    
    if (suggestions.length === 0) {
        // Permettre la création d'un nouveau tag
        suggestionsContainer.innerHTML = `
            <div class="tag-suggestion" data-tag="${query}" onclick="addTag('${query}')">
                <span>🆕</span>
                <span>Créer "${query}"</span>
            </div>
        `;
    } else {
        suggestionsContainer.innerHTML = suggestions.map(tag => `
            <div class="tag-suggestion" data-tag="${tag.name}" onclick="addTag('${tag.name}')">
                <span>${tag.type === 'genre' ? '🎵' : '🎤'}</span>
                <span>${tag.name}</span>
            </div>
        `).join('');
    }
    
    suggestionsContainer.classList.add('show');
}

function hideTagSuggestions() {
    setTimeout(() => {
        const suggestionsContainer = document.getElementById('tag-suggestions');
        if (suggestionsContainer) {
            suggestionsContainer.classList.remove('show');
        }
    }, 200);
}

function addTag(tagName) {
    if (!tagName || selectedTags.includes(tagName) || selectedTags.length >= 10) return;
    
    selectedTags.push(tagName);
    updateTagsDisplay();
    updateTagsHiddenInput();
    
    // Vider l'input
    const tagInput = document.getElementById('tag-input');
    if (tagInput) {
        tagInput.value = '';
    }
    
    hideTagSuggestions();
}

function removeTag(button) {
    const tagElement = button.parentElement;
    const tagName = tagElement.getAttribute('data-tag');
    
    // Retirer du tableau
    selectedTags = selectedTags.filter(tag => tag !== tagName);
    
    // Supprimer l'élément avec animation
    tagElement.style.animation = 'slideOut 0.3s ease forwards';
    setTimeout(() => {
        tagElement.remove();
        updateTagsHiddenInput();
    }, 300);
}

function updateTagsDisplay() {
    const selectedTagsContainer = document.getElementById('selected-tags');
    if (!selectedTagsContainer) return;
    
    const newTag = selectedTags[selectedTags.length - 1];
    const tagElement = document.createElement('span');
    tagElement.className = 'selected-tag';
    tagElement.setAttribute('data-tag', newTag);
    tagElement.innerHTML = `
        ${newTag}
        <button type="button" class="remove-tag" onclick="removeTag(this)">×</button>
    `;
    
    selectedTagsContainer.appendChild(tagElement);
}

function updateTagsHiddenInput() {
    const tagsHiddenInput = document.getElementById('tags-hidden');
    if (tagsHiddenInput) {
        tagsHiddenInput.value = selectedTags.join(',');
    }
}

// ===== VALIDATION DU FORMULAIRE =====
function initFormValidation() {
    const form = document.querySelector('.edit-form');
    if (!form) return;
    
    form.addEventListener('submit', function(e) {
        if (!validateForm()) {
            e.preventDefault();
        }
    });
    
    // Validation en temps réel
    const titleInput = document.getElementById('title');
    const descriptionInput = document.getElementById('description');
    
    if (titleInput) {
        titleInput.addEventListener('blur', () => validateTitle(titleInput));
        titleInput.addEventListener('input', () => clearFieldError(titleInput));
    }
    
    if (descriptionInput) {
        descriptionInput.addEventListener('blur', () => validateDescription(descriptionInput));
        descriptionInput.addEventListener('input', () => clearFieldError(descriptionInput));
    }
}

function validateForm() {
    let isValid = true;
    
    const titleInput = document.getElementById('title');
    const descriptionInput = document.getElementById('description');
    
    if (titleInput && !validateTitle(titleInput)) {
        isValid = false;
    }
    
    if (descriptionInput && !validateDescription(descriptionInput)) {
        isValid = false;
    }
    
    return isValid;
}

function validateTitle(input) {
    const value = input.value.trim();
    
    if (value.length < 5) {
        showFieldError(input, 'Le titre doit faire au moins 5 caractères');
        return false;
    }
    
    if (value.length > 200) {
        showFieldError(input, 'Le titre ne peut pas dépasser 200 caractères');
        return false;
    }
    
    clearFieldError(input);
    return true;
}

function validateDescription(input) {
    const value = input.value.trim();
    
    if (value.length < 10) {
        showFieldError(input, 'La description doit faire au moins 10 caractères');
        return false;
    }
    
    clearFieldError(input);
    return true;
}

function showFieldError(input, message) {
    input.classList.add('error');
    
    // Supprimer l'ancien message d'erreur
    const existingError = input.parentNode.querySelector('.field-error');
    if (existingError) {
        existingError.remove();
    }
    
    // Ajouter le nouveau message
    const errorDiv = document.createElement('div');
    errorDiv.className = 'field-error';
    errorDiv.style.color = '#ff6b6b';
    errorDiv.style.fontSize = '0.875rem';
    errorDiv.style.marginTop = '4px';
    errorDiv.textContent = message;
    
    input.parentNode.appendChild(errorDiv);
}

function clearFieldError(input) {
    input.classList.remove('error');
    
    const existingError = input.parentNode.querySelector('.field-error');
    if (existingError) {
        existingError.remove();
    }
}

// ===== NOTIFICATIONS =====
function showNotification(message, type = 'info') {
    // Supprimer les anciennes notifications
    const existingNotification = document.querySelector('.floating-notification');
    if (existingNotification) {
        existingNotification.remove();
    }
    
    const notification = document.createElement('div');
    notification.className = `floating-notification ${type}`;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 16px 20px;
        border-radius: 12px;
        color: white;
        font-weight: 500;
        z-index: 10000;
        animation: slideInRight 0.3s ease;
        max-width: 400px;
        box-shadow: 0 4px 20px rgba(0,0,0,0.15);
    `;
    
    // Couleurs selon le type
    switch (type) {
        case 'success':
            notification.style.background = 'linear-gradient(135deg, #48bb78 0%, #38a169 100%)';
            break;
        case 'error':
            notification.style.background = 'linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%)';
            break;
        case 'info':
        default:
            notification.style.background = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
            break;
    }
    
    notification.innerHTML = `
        <div style="display: flex; align-items: center; gap: 10px;">
            <span>${type === 'success' ? '✅' : type === 'error' ? '❌' : 'ℹ️'}</span>
            <span>${message}</span>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Supprimer automatiquement après 4 secondes
    setTimeout(() => {
        if (notification.parentNode) {
            notification.style.animation = 'slideOutRight 0.3s ease forwards';
            setTimeout(() => notification.remove(), 300);
        }
    }, 4000);
}

// Ajout des animations CSS
const style = document.createElement('style');
style.textContent = `
    @keyframes slideInRight {
        from {
            opacity: 0;
            transform: translateX(100%);
        }
        to {
            opacity: 1;
            transform: translateX(0);
        }
    }
    
    @keyframes slideOutRight {
        from {
            opacity: 1;
            transform: translateX(0);
        }
        to {
            opacity: 0;
            transform: translateX(100%);
        }
    }
    
    @keyframes slideOut {
        from {
            opacity: 1;
            transform: scale(1);
        }
        to {
            opacity: 0;
            transform: scale(0.8);
        }
    }
    
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
`;
document.head.appendChild(style); 