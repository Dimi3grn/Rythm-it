// JavaScript pour la page thread
document.addEventListener('DOMContentLoaded', function() {
    
    // Gestion du formulaire de commentaire via API JSON
    const commentForm = document.getElementById('comment-form');
    if (commentForm) {
        commentForm.addEventListener('submit', async function(e) {
            e.preventDefault(); // Empêcher le submit normal
            
            const textarea = this.querySelector('.comment-input');
            const submitBtn = this.querySelector('.comment-btn');
            const content = textarea.value.trim();
            
            if (!content) {
                showNotification('Le commentaire ne peut pas être vide', 'error');
                return;
            }
            
            // Désactiver le bouton pendant l'envoi
            const originalText = submitBtn.textContent;
            submitBtn.disabled = true;
            submitBtn.textContent = 'Envoi...';
            
            try {
                // Extraire l'ID du thread depuis l'URL
                const threadId = window.location.pathname.split('/thread/')[1];
                
                const response = await fetch(`/threads/${threadId}/messages`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    credentials: 'same-origin',
                    body: JSON.stringify({ content: content })
                });
                
                const data = await response.json();
                
                if (data.success) {
                    // Vider le textarea
                    textarea.value = '';
                    
                    // Ajouter le nouveau commentaire au DOM
                    addCommentToDOM(data.data);
                    
                    // Mettre à jour le compteur de commentaires
                    updateCommentsCount();
                    
                    showNotification('Commentaire ajouté avec succès !', 'success');
                } else {
                    throw new Error(data.message || 'Erreur lors de l\'ajout du commentaire');
                }
                
            } catch (error) {
                console.error('Erreur ajout commentaire:', error);
                
                // En cas d'erreur, utiliser le fallback classique
                console.log('Fallback vers POST classique...');
                commentForm.removeEventListener('submit', arguments.callee);
                commentForm.submit();
                
            } finally {
                // Réactiver le bouton
                submitBtn.disabled = false;
                submitBtn.textContent = originalText;
            }
        });
    }
    
    // Fonction pour ajouter un commentaire au DOM
    function addCommentToDOM(commentData) {
        const commentsList = document.querySelector('.comments-list');
        if (!commentsList) return;
        
        // Générer les initiales depuis le nom d'utilisateur
        const initials = generateInitials(commentData.author.username);
        
        // Créer l'élément HTML du commentaire
        const commentHTML = `
            <div class="comment-item">
                <div class="comment-avatar">
                    <div class="user-pic">${initials}</div>
                </div>
                <div class="comment-content">
                    <div class="comment-header">
                        <h4>${commentData.author.username}</h4>
                        <span class="comment-time">à l'instant</span>
                    </div>
                    <div class="comment-text">
                        ${commentData.content}
                    </div>
                    <div class="comment-actions">
                        <button class="comment-action">❤️ 0</button>
                        <button class="comment-action">💬 Répondre</button>
                        <button class="comment-action">📤 Partager</button>
                    </div>
                </div>
            </div>
        `;
        
        // Ajouter le commentaire au début de la liste
        commentsList.insertAdjacentHTML('afterbegin', commentHTML);
        
        // Animation d'apparition
        const newComment = commentsList.firstElementChild;
        newComment.style.opacity = '0';
        newComment.style.transform = 'translateY(-20px)';
        
        requestAnimationFrame(() => {
            newComment.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
            newComment.style.opacity = '1';
            newComment.style.transform = 'translateY(0)';
        });
    }
    
    // Fonction pour mettre à jour le compteur de commentaires
    function updateCommentsCount() {
        const commentsHeader = document.querySelector('.comments-header h3');
        if (commentsHeader) {
            const currentText = commentsHeader.textContent;
            const currentCount = parseInt(currentText.match(/\d+/)[0]) || 0;
            const newCount = currentCount + 1;
            commentsHeader.textContent = `Commentaires (${newCount})`;
        }
        
        // Mettre à jour aussi le compteur dans le bouton d'engagement
        const engagementBtn = document.querySelector('.engagement-btn .btn-count');
        if (engagementBtn) {
            const currentCount = parseInt(engagementBtn.textContent) || 0;
            engagementBtn.textContent = currentCount + 1;
        }
    }
    
    // Fonction pour générer les initiales
    function generateInitials(username) {
        if (!username) return '??';
        
        const words = username.split(/[\s_-]+/);
        if (words.length >= 2) {
            return (words[0].charAt(0) + words[1].charAt(0)).toUpperCase();
        } else {
            return username.substring(0, 2).toUpperCase();
        }
    }
    
    // Fonction pour afficher les notifications
    function showNotification(message, type = 'info') {
        // Vérifier si la fonction globale existe
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
            return;
        }
        
        // Fallback simple
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: ${type === 'error' ? '#ff6b6b' : type === 'success' ? '#4ade80' : '#6c63ff'};
            color: white;
            padding: 16px 20px;
            border-radius: 8px;
            z-index: 10000;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            transition: all 0.3s ease;
            transform: translateX(100%);
        `;
        
        document.body.appendChild(notification);
        
        // Animation d'entrée
        requestAnimationFrame(() => {
            notification.style.transform = 'translateX(0)';
        });
        
        // Retirer après 4 secondes
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 4000);
    }
    
    // Auto-resize du textarea
    const textarea = document.querySelector('.comment-input');
    if (textarea) {
        textarea.addEventListener('input', function() {
            this.style.height = 'auto';
            this.style.height = Math.min(this.scrollHeight, 150) + 'px';
        });
    }
    
    console.log('🧵 Thread.js chargé avec succès');
});

// Fonction pour supprimer un thread
function deleteThread(threadId) {
    // Demander confirmation
    if (!confirm('Êtes-vous sûr de vouloir supprimer ce thread ? Cette action est irréversible.')) {
        return;
    }
    
    // Afficher un loader
    const deleteBtn = document.querySelector('.delete-btn');
    if (deleteBtn) {
        deleteBtn.disabled = true;
        deleteBtn.innerHTML = '⏳';
    }
    
    // Envoyer la requête de suppression
    fetch(`/thread/${threadId}/delete`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        credentials: 'same-origin'
    })
    .then(response => {
        if (response.ok) {
            // Rediriger vers la page d'accueil avec message de succès
            window.location.href = '/?success=thread_deleted';
        } else {
            throw new Error('Erreur lors de la suppression');
        }
    })
    .catch(error => {
        console.error('Erreur suppression thread:', error);
        
        // Restaurer le bouton
        if (deleteBtn) {
            deleteBtn.disabled = false;
            deleteBtn.innerHTML = '🗑️';
        }
        
        // Afficher une erreur
        if (typeof showNotification === 'function') {
            showNotification('Erreur lors de la suppression du thread', 'error');
        } else {
            alert('Erreur lors de la suppression du thread');
        }
    });
} 