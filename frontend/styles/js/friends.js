// JavaScript pour la page amis - friends.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    const messageModal = document.getElementById('messageModal');
    const modalUserName = document.getElementById('modalUserName');
    const searchInput = document.querySelector('.search-input');
    const friendCards = document.querySelectorAll('.friend-card');
    const filterItems = document.querySelectorAll('.filter-item');
    
    // Gestion de la recherche d'amis
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            const searchTerm = this.value.toLowerCase();
            
            friendCards.forEach(card => {
                const friendName = card.querySelector('.friend-info h3').textContent.toLowerCase();
                const friendUsername = card.querySelector('.friend-username').textContent.toLowerCase();
                
                if (friendName.includes(searchTerm) || friendUsername.includes(searchTerm)) {
                    card.style.display = 'block';
                    card.style.opacity = '1';
                    card.style.transform = 'translateY(0)';
                } else {
                    card.style.opacity = '0';
                    card.style.transform = 'translateY(20px)';
                    setTimeout(() => {
                        if (!friendName.includes(searchTerm) && !friendUsername.includes(searchTerm)) {
                            card.style.display = 'none';
                        }
                    }, 200);
                }
            });
        });
    }
    
    // Gestion des filtres
    filterItems.forEach(filter => {
        filter.addEventListener('click', function(e) {
            e.preventDefault();
            
            // Retirer la classe active de tous les filtres
            filterItems.forEach(f => f.classList.remove('active'));
            
            // Ajouter la classe active au filtre cliqué
            this.classList.add('active');
            
            const filterType = this.getAttribute('data-filter');
            
            friendCards.forEach(card => {
                const status = card.getAttribute('data-status');
                const activity = card.getAttribute('data-activity');
                
                let shouldShow = false;
                
                switch(filterType) {
                    case 'all':
                        shouldShow = true;
                        break;
                    case 'online':
                        shouldShow = status === 'online';
                        break;
                    case 'music':
                        shouldShow = activity === 'music';
                        break;
                }
                
                if (shouldShow) {
                    card.style.display = 'block';
                    setTimeout(() => {
                        card.style.opacity = '1';
                        card.style.transform = 'translateY(0)';
                    }, 50);
                } else {
                    card.style.opacity = '0';
                    card.style.transform = 'translateY(20px)';
                    setTimeout(() => {
                        card.style.display = 'none';
                    }, 200);
                }
            });
        });
    });
    
    // Gestion des boutons de message
    document.querySelectorAll('.message-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const friendName = this.getAttribute('data-friend') || 'Ami';
            openMessageModal(friendName);
        });
    });
    
    // Fonction pour ouvrir le modal de message
    function openMessageModal(friendName) {
        modalUserName.textContent = friendName;
        messageModal.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        // Focus sur l'input de message
        setTimeout(() => {
            document.querySelector('.message-input').focus();
        }, 300);
    }
    
    // Fermer le modal
    document.querySelector('.close-modal').addEventListener('click', function() {
        closeMessageModal();
    });
    
    // Fermer en cliquant à l'extérieur
    messageModal.addEventListener('click', function(e) {
        if (e.target === messageModal) {
            closeMessageModal();
        }
    });
    
    // Fonction pour fermer le modal
    function closeMessageModal() {
        messageModal.classList.remove('active');
        document.body.style.overflow = 'auto';
    }
    
    // Gestion de l'envoi de messages
    const messageInput = document.querySelector('.message-input');
    const sendBtn = document.querySelector('.send-btn');
    const messageHistory = document.querySelector('.message-history');
    
    function sendMessage() {
        const messageText = messageInput.value.trim();
        if (messageText === '') return;
        
        // Créer le message
        const messageElement = document.createElement('div');
        messageElement.className = 'message sent';
        messageElement.innerHTML = `
            <div class="message-content">${messageText}</div>
            <span class="message-timestamp">${getCurrentTime()}</span>
        `;
        
        // Ajouter à l'historique
        messageHistory.appendChild(messageElement);
        
        // Faire défiler vers le bas
        messageHistory.scrollTop = messageHistory.scrollHeight;
        
        // Vider l'input
        messageInput.value = '';
        
        // Animation d'envoi
        sendBtn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            sendBtn.style.transform = 'scale(1)';
        }, 150);
        
        // Simuler une réponse (optionnel)
        setTimeout(() => {
            simulateResponse();
        }, 1000 + Math.random() * 2000);
    }
    
    // Fonction pour obtenir l'heure actuelle
    function getCurrentTime() {
        const now = new Date();
        return now.getHours().toString().padStart(2, '0') + ':' + 
               now.getMinutes().toString().padStart(2, '0');
    }
    
    // Simuler une réponse automatique
    function simulateResponse() {
        const responses = [
            "C'est une excellente recommandation ! 🎵",
            "Je vais écouter ça tout de suite !",
            "Tu as vraiment bon goût musical 👌",
            "Merci pour le partage !",
            "Cette playlist est parfaite pour travailler",
            "On fait une session d'écoute ensemble ? 🎧"
        ];
        
        const randomResponse = responses[Math.floor(Math.random() * responses.length)];
        
        const messageElement = document.createElement('div');
        messageElement.className = 'message received';
        messageElement.innerHTML = `
            <div class="message-content">${randomResponse}</div>
            <span class="message-timestamp">${getCurrentTime()}</span>
        `;
        
        messageHistory.appendChild(messageElement);
        messageHistory.scrollTop = messageHistory.scrollHeight;
    }
    
    // Envoyer avec le bouton
    sendBtn.addEventListener('click', sendMessage);
    
    // Envoyer avec Entrée
    messageInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            sendMessage();
        }
    });
    
    // Animations d'apparition pour les cartes
    const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry, index) => {
            if (entry.isIntersecting) {
                setTimeout(() => {
                    entry.target.style.opacity = '1';
                    entry.target.style.transform = 'translateY(0)';
                }, index * 100);
            }
        });
    }, { threshold: 0.1 });
    
    // Observer les cartes d'amis
    friendCards.forEach(card => {
        card.style.opacity = '0';
        card.style.transform = 'translateY(30px)';
        card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(card);
    });
    
    // Gestion des menus d'amis
    document.querySelectorAll('.friend-menu').forEach(menu => {
        menu.addEventListener('click', function(e) {
            e.stopPropagation();
            
            // Fermer tous les autres menus
            document.querySelectorAll('.friend-menu-dropdown').forEach(dropdown => {
                dropdown.remove();
            });
            
            // Créer le menu dropdown
            const dropdown = document.createElement('div');
            dropdown.className = 'friend-menu-dropdown';
            dropdown.innerHTML = `
                <div class="menu-item">👤 Voir le profil</div>
                <div class="menu-item">🎵 Voir les playlists</div>
                <div class="menu-item">🚫 Masquer les posts</div>
                <div class="menu-item danger">❌ Supprimer l'ami</div>
            `;
            
            // Positionner le menu
            const rect = this.getBoundingClientRect();
            dropdown.style.position = 'fixed';
            dropdown.style.top = (rect.bottom + 5) + 'px';
            dropdown.style.right = (window.innerWidth - rect.right) + 'px';
            dropdown.style.zIndex = '1000';
            
            document.body.appendChild(dropdown);
            
            // Fermer le menu en cliquant ailleurs
            setTimeout(() => {
                document.addEventListener('click', function closeMenu() {
                    dropdown.remove();
                    document.removeEventListener('click', closeMenu);
                });
            }, 100);
        });
    });
    
    // Gestion des boutons d'action rapide
    document.querySelectorAll('.friend-actions .friend-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            // Animation de feedback
            this.style.transform = 'scale(0.95)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 150);
        });
    });
    
    // Gestion du statut en temps réel (simulation)
    function updateOnlineStatus() {
        const onlineStatuses = document.querySelectorAll('.online-status');
        onlineStatuses.forEach(status => {
            if (status.classList.contains('online')) {
                // Petite animation de pulsation
                status.style.transform = 'scale(1.2)';
                setTimeout(() => {
                    status.style.transform = 'scale(1)';
                }, 200);
            }
        });
    }
    
    // Mettre à jour les statuts toutes les 5 secondes
    setInterval(updateOnlineStatus, 5000);
    
    // Gestion du scroll infini (simulation)
    let isLoading = false;
    window.addEventListener('scroll', function() {
        if (window.innerHeight + window.scrollY >= document.body.offsetHeight - 1000 && !isLoading) {
            isLoading = true;
            // Simuler le chargement de plus d'amis
            setTimeout(() => {
                isLoading = false;
                // Ici on pourrait ajouter plus de cartes d'amis
            }, 1000);
        }
    });
    
    // Gestion des notifications en temps réel
    function showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <span class="notification-icon">${type === 'message' ? '💬' : 'ℹ️'}</span>
                <span class="notification-text">${message}</span>
            </div>
        `;
        
        // Styles pour la notification
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: rgba(26, 26, 46, 0.95);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 15px 20px;
            color: #f0f0f0;
            z-index: 10000;
            transform: translateX(400px);
            transition: transform 0.3s ease;
            backdrop-filter: blur(10px);
        `;
        
        document.body.appendChild(notification);
        
        // Animation d'entrée
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        // Suppression automatique
        setTimeout(() => {
            notification.style.transform = 'translateX(400px)';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.remove();
                }
            }, 300);
        }, 4000);
    }
    
    // Simuler des notifications de messages
    setTimeout(() => {
        showNotification('MixMaster vous a envoyé un message', 'message');
        
        // Mettre à jour le badge de notification
        const badge = document.querySelector('.notification-badge');
        if (badge) {
            badge.textContent = parseInt(badge.textContent) + 1;
        }
    }, 10000);
    
    // Gestion des raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // Fermer le modal avec Échap
        if (e.key === 'Escape' && messageModal.classList.contains('active')) {
            closeMessageModal();
        }
        
        // Focus sur la recherche avec Ctrl+F
        if (e.ctrlKey && e.key === 'f' && searchInput) {
            e.preventDefault();
            searchInput.focus();
        }
    });
    
    // Performance: lazy loading des avatars
    const avatarObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const avatar = entry.target;
                // Ici on pourrait charger les vraies images d'avatar
                avatar.style.opacity = '1';
                avatarObserver.unobserve(avatar);
            }
        });
    });
    
    document.querySelectorAll('.user-pic').forEach(avatar => {
        avatarObserver.observe(avatar);
    });
    
    console.log('🎵 Page amis Rythm\'it initialisée avec succès !');
});

// CSS pour les styles de menu et notifications
const additionalStyles = `
.friend-menu-dropdown {
    background: rgba(26, 26, 46, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 8px 0;
    backdrop-filter: blur(10px);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
    min-width: 180px;
}

.menu-item {
    padding: 10px 16px;
    color: #f0f0f0;
    cursor: pointer;
    transition: background 0.2s ease;
    font-size: 14px;
    display: flex;
    align-items: center;
    gap: 8px;
}

.menu-item:hover {
    background: rgba(255, 255, 255, 0.05);
}

.menu-item.danger {
    color: #ff6b6b;
}

.menu-item.danger:hover {
    background: rgba(255, 107, 107, 0.1);
}

.notification {
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.notification-content {
    display: flex;
    align-items: center;
    gap: 10px;
}

.notification-icon {
    font-size: 16px;
}

.notification-text {
    font-size: 14px;
    font-weight: 500;
}
`;

// Ajouter les styles supplémentaires
if (!document.getElementById('friends-additional-styles')) {
    const styleSheet = document.createElement('style');
    styleSheet.id = 'friends-additional-styles';
    styleSheet.textContent = additionalStyles;
    document.head.appendChild(styleSheet);
}