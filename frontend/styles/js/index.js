// JavaScript mis à jour pour index.js avec fonctionnalités amis et notifications

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let notificationCount = 3;
    let onlineFriends = ['MixMaster', 'SoundBliss', 'RhythmHunter'];
    let lastActivity = Date.now();
    
    // Animation d'apparition pour les threads
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, { threshold: 0.1 });

    document.querySelectorAll('.thread-item').forEach(thread => {
        thread.style.opacity = '0';
        thread.style.transform = 'translateY(20px)';
        thread.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
        observer.observe(thread);
    });

    // Gestion des contrôles de lecture
    document.querySelectorAll('.play-control').forEach(btn => {
        btn.addEventListener('click', function() {
            const isPlaying = this.textContent.includes('⏸️');
            
            // Arrêter tous les autres lecteurs
            document.querySelectorAll('.play-control').forEach(otherBtn => {
                if (otherBtn !== this) {
                    otherBtn.textContent = '▶️';
                    otherBtn.closest('.music-card').classList.remove('playing');
                }
            });
            
            // Basculer l'état du lecteur actuel
            this.textContent = isPlaying ? '▶️' : '⏸️';
            const musicCard = this.closest('.music-card');
            
            if (isPlaying) {
                musicCard.classList.remove('playing');
            } else {
                musicCard.classList.add('playing');
                // Simuler le partage d'activité avec les amis
                broadcastListeningActivity(musicCard);
            }
            
            // Animation
            this.style.transform = 'scale(0.9)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 150);
        });
    });

    // Fonction pour diffuser l'activité d'écoute
    function broadcastListeningActivity(musicCard) {
        const trackTitle = musicCard.querySelector('.track-info h5').textContent;
        const artist = musicCard.querySelector('.track-info p').textContent.split('•')[0].trim();
        
        showNotification(`🎵 Vous écoutez "${trackTitle}" par ${artist}`, 'music');
        
        // Mettre à jour le statut dans la sidebar
        updateListeningStatus(trackTitle, artist);
    }

    // Mettre à jour le statut d'écoute
    function updateListeningStatus(track, artist) {
        const userStatus = document.querySelector('.sidebar-right .friend-online');
        if (userStatus) {
            const statusText = userStatus.querySelector('.friend-info p');
            statusText.textContent = `🎵 Écoute: ${track}`;
        }
    }

    // Gestion des likes avec animation coeur
    document.querySelectorAll('.engagement-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            if (this.textContent.includes('❤️')) {
                this.classList.toggle('liked');
                
                // Animation de coeur
                this.style.transform = 'scale(1.2)';
                setTimeout(() => {
                    this.style.transform = 'scale(1)';
                }, 200);
                
                // Mettre à jour le compteur
                const countSpan = this.textContent.match(/\d+/);
                if (countSpan) {
                    const currentCount = parseInt(countSpan[0]);
                    const newCount = this.classList.contains('liked') ? currentCount + 1 : currentCount - 1;
                    this.innerHTML = this.innerHTML.replace(/\d+/, newCount);
                }
            } else if (this.textContent.includes('💬')) {
                // Gestion des commentaires
                showCommentModal(this.closest('.thread-item'));
            } else if (this.textContent.includes('📩')) {
                // Envoyer un message privé
                const userName = this.closest('.thread-item').querySelector('.user-details h4').textContent;
                showPrivateMessageModal(userName);
            }
        });
    });

    // Auto-resize du textarea avec animation
    const textarea = document.querySelector('.composer-input');
    if (textarea) {
        textarea.addEventListener('input', function() {
            this.style.height = 'auto';
            this.style.height = Math.min(this.scrollHeight, 200) + 'px';
        });
        
        textarea.addEventListener('focus', function() {
            this.parentElement.style.transform = 'scale(1.02)';
            this.parentElement.style.transition = 'transform 0.3s ease';
        });

        textarea.addEventListener('blur', function() {
            this.parentElement.style.transform = 'scale(1)';
        });
    }

    // Gestion de la publication
    const publishBtn = document.querySelector('.publish-btn');
    if (publishBtn) {
        publishBtn.addEventListener('click', function() {
            const content = textarea.value.trim();
            if (content) {
                publishPost(content);
                textarea.value = '';
                textarea.style.height = '120px';
            }
        });
    }

    // Fonction pour publier un post
    function publishPost(content) {
        const newThread = document.createElement('article');
        newThread.className = 'thread-item';
        newThread.style.opacity = '0';
        newThread.style.transform = 'translateY(30px)';
        
        newThread.innerHTML = `
            <div class="thread-header">
                <div class="user-pic">MO</div>
                <div class="user-details">
                    <h4>Moi</h4>
                    <span class="meta">À l'instant • Personnel</span>
                </div>
            </div>
            <div class="thread-text">
                ${content}
            </div>
            <div class="thread-engagement">
                <button class="engagement-btn">❤️ 0</button>
                <button class="engagement-btn">💬 0</button>
                <button class="engagement-btn">🔄 0</button>
                <button class="engagement-btn">📩</button>
                <button class="engagement-btn">🔖</button>
            </div>
        `;
        
        // Insérer au début du contenu
        const contentArea = document.querySelector('.content-area');
        const firstThread = contentArea.querySelector('.thread-item');
        contentArea.insertBefore(newThread, firstThread);
        
        // Animation d'apparition
        setTimeout(() => {
            newThread.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
            newThread.style.opacity = '1';
            newThread.style.transform = 'translateY(0)';
        }, 100);
        
        // Réattacher les événements
        attachEventListeners(newThread);
        
        showNotification('Post publié avec succès ! 🎉', 'success');
    }

    // Fonction pour réattacher les événements aux nouveaux éléments
    function attachEventListeners(element) {
        element.querySelectorAll('.engagement-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                if (this.textContent.includes('❤️')) {
                    this.classList.toggle('liked');
                    this.style.transform = 'scale(1.2)';
                    setTimeout(() => {
                        this.style.transform = 'scale(1)';
                    }, 200);
                }
            });
        });
    }

    // Navigation active
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', function(e) {
            e.preventDefault();
            document.querySelectorAll('.nav-item').forEach(nav => nav.classList.remove('active'));
            this.classList.add('active');
            
            // Animation de feedback
            this.style.transform = 'scale(0.95)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 150);
        });
    });

    // Hover effects pour les widgets avec son
    document.querySelectorAll('.trend-item').forEach(item => {
        item.addEventListener('mouseenter', function() {
            this.style.transform = 'translateX(5px)';
            this.style.background = 'rgba(255, 255, 255, 0.05)';
        });
        
        item.addEventListener('mouseleave', function() {
            this.style.transform = 'translateX(0)';
            this.style.background = 'transparent';
        });
    });

    // Gestion des amis en ligne
    document.querySelectorAll('.friend-online .message-btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const friendName = this.closest('.friend-online').querySelector('h5').textContent;
            openQuickMessage(friendName);
        });
    });

    // Fonction pour ouvrir un message rapide
    function openQuickMessage(friendName) {
        const quickModal = document.createElement('div');
        quickModal.className = 'quick-message-modal';
        quickModal.innerHTML = `
            <div class="quick-message-content">
                <div class="quick-message-header">
                    <h4>Message à ${friendName}</h4>
                    <button class="close-quick">✕</button>
                </div>
                <textarea class="quick-message-input" placeholder="Tapez votre message..."></textarea>
                <div class="quick-message-actions">
                    <button class="send-quick-btn">Envoyer</button>
                </div>
            </div>
        `;
        
        document.body.appendChild(quickModal);
        
        // Styles inline pour le modal rapide
        quickModal.style.cssText = `
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: rgba(26, 26, 46, 0.95);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            padding: 20px;
            width: 320px;
            z-index: 1000;
            backdrop-filter: blur(10px);
            transform: translateY(100px);
            opacity: 0;
            transition: all 0.3s ease;
        `;
        
        // Animation d'entrée
        setTimeout(() => {
            quickModal.style.transform = 'translateY(0)';
            quickModal.style.opacity = '1';
        }, 10);
        
        // Focus sur l'input
        const input = quickModal.querySelector('.quick-message-input');
        input.focus();
        
        // Gestion de l'envoi
        const sendBtn = quickModal.querySelector('.send-quick-btn');
        sendBtn.addEventListener('click', () => {
            if (input.value.trim()) {
                showNotification(`Message envoyé à ${friendName} !`, 'success');
                closeQuickMessage(quickModal);
            }
        });
        
        // Fermeture
        quickModal.querySelector('.close-quick').addEventListener('click', () => {
            closeQuickMessage(quickModal);
        });
        
        // Envoyer avec Entrée
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && e.ctrlKey) {
                sendBtn.click();
            }
        });
    }

    function closeQuickMessage(modal) {
        modal.style.transform = 'translateY(100px)';
        modal.style.opacity = '0';
        setTimeout(() => {
            if (modal.parentNode) {
                modal.remove();
            }
        }, 300);
    }

    // Animation pour les éléments de la sidebar
    const sidebarObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry, index) => {
            if (entry.isIntersecting) {
                setTimeout(() => {
                    entry.target.style.opacity = '1';
                    entry.target.style.transform = 'translateX(0)';
                }, index * 100);
            }
        });
    }, { threshold: 0.2 });

    document.querySelectorAll('.nav-section, .widget').forEach(element => {
        element.style.opacity = '0';
        element.style.transform = 'translateX(-20px)';
        element.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        sidebarObserver.observe(element);
    });

    // Système de notifications
    function showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        const icons = {
            info: 'ℹ️',
            success: '✅',
            music: '🎵',
            message: '💬',
            warning: '⚠️'
        };
        
        notification.innerHTML = `
            <div class="notification-content">
                <span class="notification-icon">${icons[type] || icons.info}</span>
                <span class="notification-text">${message}</span>
                <button class="notification-close">✕</button>
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
            max-width: 350px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
        `;
        
        document.body.appendChild(notification);
        
        // Animation d'entrée
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);
        
        // Bouton de fermeture
        notification.querySelector('.notification-close').addEventListener('click', () => {
            closeNotification(notification);
        });
        
        // Suppression automatique
        setTimeout(() => {
            closeNotification(notification);
        }, 5000);
    }

    function closeNotification(notification) {
        notification.style.transform = 'translateX(400px)';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.remove();
            }
        }, 300);
    }

    // Mettre à jour le badge de notification
    function updateNotificationBadge() {
        notificationCount++;
        const badge = document.querySelector('.notification-badge');
        if (badge) {
            badge.textContent = notificationCount;
        }
    }

    // Gestion du clic sur les notifications
   document.querySelector('.notification-btn').addEventListener('click', function(e) {
    // Ne pas empêcher la navigation par défaut
    // e.preventDefault(); ← SUPPRIMÉ
    
    // Si on est déjà sur la page messages, alors on gère les notifications
    if (window.location.pathname.includes('messages.html')) {
        e.preventDefault(); // Seulement si on est déjà sur la page
        notificationCount = 0;
        const badge = document.querySelector('.notification-badge');
        if (badge) {
            badge.style.transform = 'scale(0)';
            setTimeout(() => {
                badge.textContent = '0';
                badge.style.display = 'none';
            }, 200);
        }
        showNotification('Notifications marquées comme lues', 'info');
    } else {
        // Laisser la navigation normale se faire
        showNotification('Redirection vers Messages...', 'info');
    }
});

    // Détection d'inactivité
    function trackUserActivity() {
        const events = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
        
        events.forEach(event => {
            document.addEventListener(event, () => {
                lastActivity = Date.now();
            }, true);
        });
        
        // Vérifier l'inactivité toutes les minutes
        setInterval(() => {
            const inactiveTime = Date.now() - lastActivity;
            if (inactiveTime > 300000) { // 5 minutes
                // Marquer comme absent
                document.querySelectorAll('.online-status.online').forEach(status => {
                    status.classList.remove('online');
                    status.classList.add('away');
                });
            }
        }, 60000);
    }

    // Gestion du mode plein écran pour la musique
    document.querySelectorAll('.music-card').forEach(card => {
        card.addEventListener('dblclick', function() {
            this.classList.toggle('fullscreen-music');
            
            if (this.classList.contains('fullscreen-music')) {
                // Mode plein écran
                this.style.cssText = `
                    position: fixed;
                    top: 50%;
                    left: 50%;
                    transform: translate(-50%, -50%);
                    width: 80vw;
                    height: 60vh;
                    z-index: 9999;
                    background: rgba(26, 26, 46, 0.98);
                    backdrop-filter: blur(20px);
                    flex-direction: column;
                    justify-content: center;
                    text-align: center;
                `;
                
                // Effet de flou sur le reste
                document.body.style.overflow = 'hidden';
                document.querySelector('.app-container').style.filter = 'blur(5px)';
            } else {
                // Retour normal
                this.style.cssText = '';
                document.body.style.overflow = 'auto';
                document.querySelector('.app-container').style.filter = 'none';
            }
        });
    });

    // Smooth scroll pour la navigation
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Optimisation pour les performances
    let resizeTimer;
    window.addEventListener('resize', function() {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(function() {
            // Réinitialise certaines animations après redimensionnement
            document.querySelectorAll('.thread-item').forEach(item => {
                item.style.transition = 'all 0.3s ease';
            });
        }, 250);
    });

    // Raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // Ctrl + M pour ouvrir les messages
        if (e.ctrlKey && e.key === 'm') {
            e.preventDefault();
            document.querySelector('.notification-btn').click();
        }
        
        // Ctrl + N pour nouveau post
        if (e.ctrlKey && e.key === 'n') {
            e.preventDefault();
            document.querySelector('.composer-input').focus();
        }
        
        // Échap pour fermer les modaux plein écran
        if (e.key === 'Escape') {
            document.querySelectorAll('.fullscreen-music').forEach(card => {
                card.classList.remove('fullscreen-music');
                card.style.cssText = '';
            });
            document.body.style.overflow = 'auto';
            document.querySelector('.app-container').style.filter = 'none';
        }
    });

    // Initialisation des fonctionnalités
    simulateRealTimeActivity();
    trackUserActivity();
    
    // Message de démarrage
    setTimeout(() => {
        showNotification('Bienvenue sur Rythm\'it ! 🎵', 'success');
    }, 1000);
    
    console.log('🎵 Rythm\'it chargé avec succès !');
    console.log('🔥 Fonctionnalités activées: Amis, Messages, Notifications en temps réel');
});