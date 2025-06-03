// JavaScript pour la page Messages - messages.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let currentConversation = 'mixmaster';
    let typingTimeout;
    let isTyping = false;
    
    // Données des conversations (simulation)
    const conversations = {
        mixmaster: {
            name: 'MixMaster',
            avatar: 'MX',
            status: 'online',
            activity: 'Écoute: Techno Vibes',
            messages: [
                { type: 'received', content: 'Salut ! Tu as écouté le dernier album de Daft Punk ?', time: '14:32' },
                { type: 'sent', content: 'Oui ! Complètement fou 🤯', time: '14:35' },
                { type: 'received', content: 'Je te partage ma playlist techno !', time: '14:37' },
                { type: 'sent', content: 'Excellent ! Je l\'ajoute à ma bibliothèque 📚', time: '14:42' },
                { type: 'received', content: 'On fait une session d\'écoute ce soir ?', time: 'il y a 5 min' }
            ]
        },
        soundbliss: {
            name: 'SoundBliss',
            avatar: 'SB',
            status: 'online',
            activity: 'Écoute: Jazz Evening',
            messages: [
                { type: 'received', content: 'Cette playlist jazz est incroyable ! 🎷', time: '13:45' },
                { type: 'sent', content: 'Merci ! J\'ai passé des heures à la peaufiner', time: '13:47' },
                { type: 'received', content: 'Ça se sent, chaque morceau s\'enchaîne parfaitement', time: '13:50' }
            ]
        },
        rhythmhunter: {
            name: 'RhythmHunter',
            avatar: 'RH',
            status: 'away',
            activity: 'Absent',
            messages: [
                { type: 'received', content: 'On fait une session d\'écoute ce soir ?', time: '16:30' },
                { type: 'sent', content: 'Parfait ! À quelle heure ?', time: '16:32' },
                { type: 'received', content: 'Vers 20h ça te va ?', time: '16:35' }
            ]
        }
    };
    
    // Éléments DOM
    const conversationItems = document.querySelectorAll('.conversation-item');
    const chatMessages = document.getElementById('chatMessages');
    const chatInput = document.querySelector('.chat-input');
    const sendBtn = document.querySelector('.send-message-btn');
    const typingStatus = document.querySelector('.typing-status');
    const newConversationBtn = document.querySelector('.new-conversation-btn');
    const musicBtn = document.querySelector('.music-btn');
    const searchInput = document.querySelector('.search-conversations-input');
    
    // Initialisation
    init();
    
    function init() {
        // Charger la conversation active
        loadConversation(currentConversation);
        
        // Attacher les événements
        attachEventListeners();
        
        // Démarrer les simulations
        startRealTimeFeatures();
        
        // Scroll automatique vers le bas
        scrollToBottom();
    }
    
    // Gestion des événements
    function attachEventListeners() {
        // Sélection de conversation
        conversationItems.forEach(item => {
            item.addEventListener('click', function() {
                const conversationId = this.getAttribute('data-conversation');
                selectConversation(conversationId);
            });
        });
        
        // Envoi de message
        sendBtn.addEventListener('click', sendMessage);
        chatInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
        
        // Détection de frappe
        chatInput.addEventListener('input', handleTyping);
        
        // Recherche de conversations
        if (searchInput) {
            searchInput.addEventListener('input', searchConversations);
        }
        
        // Nouvelle conversation
        if (newConversationBtn) {
            newConversationBtn.addEventListener('click', openNewConversationModal);
        }
        
        // Partage de musique
        if (musicBtn) {
            musicBtn.addEventListener('click', openMusicShareModal);
        }
        
        // Actions du chat
        document.querySelectorAll('.chat-action-btn').forEach(btn => {
            btn.addEventListener('click', handleChatAction);
        });
        
        // Actions de playlist
        document.querySelectorAll('.playlist-btn').forEach(btn => {
            btn.addEventListener('click', handlePlaylistAction);
        });
        
        // Écouter ensemble
        const listenTogetherBtn = document.querySelector('.listen-together-btn');
        if (listenTogetherBtn) {
            listenTogetherBtn.addEventListener('click', startListenTogether);
        }
    }
    
    // Sélection de conversation
    function selectConversation(conversationId) {
        // Retirer la classe active de toutes les conversations
        conversationItems.forEach(item => {
            item.classList.remove('active');
        });
        
        // Ajouter la classe active à la conversation sélectionnée
        const selectedItem = document.querySelector(`[data-conversation="${conversationId}"]`);
        if (selectedItem) {
            selectedItem.classList.add('active');
            
            // Retirer le badge non lu
            const unreadBadge = selectedItem.querySelector('.unread-count');
            if (unreadBadge) {
                unreadBadge.style.transform = 'scale(0)';
                setTimeout(() => {
                    unreadBadge.remove();
                }, 200);
            }
        }
        
        // Charger la conversation
        currentConversation = conversationId;
        loadConversation(conversationId);
        
        // Mettre à jour l'en-tête du chat
        updateChatHeader(conversationId);
    }
    
    // Chargement d'une conversation
    function loadConversation(conversationId) {
        const conversation = conversations[conversationId];
        if (!conversation) return;
        
        // Vider les messages actuels
        chatMessages.innerHTML = '';
        
        // Créer le groupe de messages
        const messageGroup = document.createElement('div');
        messageGroup.className = 'message-group';
        
        // Date
        const dateDiv = document.createElement('div');
        dateDiv.className = 'message-date';
        dateDiv.textContent = 'Aujourd\'hui';
        messageGroup.appendChild(dateDiv);
        
        // Ajouter les messages
        conversation.messages.forEach(msg => {
            const messageElement = createMessageElement(msg, conversation);
            messageGroup.appendChild(messageElement);
        });
        
        chatMessages.appendChild(messageGroup);
        scrollToBottom();
    }
    
    // Création d'un élément message
    function createMessageElement(message, conversation) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${message.type}`;
        
        let content = '';
        
        if (message.type === 'received') {
            content = `
                <div class="message-avatar">
                    <div class="user-pic tiny">${conversation.avatar}</div>
                </div>
                <div class="message-content">
                    <div class="message-bubble">${message.content}</div>
                    <span class="message-timestamp">${message.time}</span>
                </div>
            `;
        } else {
            content = `
                <div class="message-content">
                    <div class="message-bubble">${message.content}</div>
                    <span class="message-timestamp">${message.time}</span>
                </div>
            `;
        }
        
        messageDiv.innerHTML = content;
        return messageDiv;
    }
    
    // Mise à jour de l'en-tête du chat
    function updateChatHeader(conversationId) {
        const conversation = conversations[conversationId];
        if (!conversation) return;
        
        // Mettre à jour les informations utilisateur
        const userPic = document.querySelector('.chat-header .user-pic');
        const userName = document.querySelector('.chat-header h3');
        const userStatus = document.querySelector('.user-status');
        const profileCard = document.querySelector('.user-profile-card');
        
        if (userPic) userPic.textContent = conversation.avatar;
        if (userName) userName.textContent = conversation.name;
        if (userStatus) {
            const statusIcon = conversation.status === 'online' ? '🟢' : 
                             conversation.status === 'away' ? '🟡' : '⚫';
            userStatus.textContent = `${statusIcon} ${conversation.activity}`;
        }
        
        // Mettre à jour la carte de profil
        if (profileCard) {
            const profilePic = profileCard.querySelector('.user-pic');
            const profileName = profileCard.querySelector('h3');
            
            if (profilePic) profilePic.textContent = conversation.avatar;
            if (profileName) profileName.textContent = conversation.name;
        }
    }
    
    // Envoi de message
    function sendMessage() {
        const messageText = chatInput.value.trim();
        if (!messageText) return;
        
        // Créer le message
        const newMessage = {
            type: 'sent',
            content: messageText,
            time: getCurrentTime()
        };
        
        // Ajouter à la conversation
        if (conversations[currentConversation]) {
            conversations[currentConversation].messages.push(newMessage);
        }
        
        // Afficher le message
        const messageElement = createMessageElement(newMessage, {avatar: 'MO'});
        const messageGroup = chatMessages.querySelector('.message-group');
        messageGroup.appendChild(messageElement);
        
        // Vider l'input
        chatInput.value = '';
        
        // Scroll vers le bas
        scrollToBottom();
        
        // Animation d'envoi
        sendBtn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            sendBtn.style.transform = 'scale(1)';
        }, 150);
        
        // Simuler une réponse
        setTimeout(() => {
            simulateResponse();
        }, 1000 + Math.random() * 2000);
        
        // Mettre à jour la liste des conversations
        updateConversationPreview(currentConversation, messageText);
    }
    
    // Gestion de la frappe
    function handleTyping() {
        if (!isTyping) {
            isTyping = true;
            showTypingIndicator();
        }
        
        clearTimeout(typingTimeout);
        typingTimeout = setTimeout(() => {
            isTyping = false;
            hideTypingIndicator();
        }, 1000);
    }
    
    // Afficher l'indicateur de frappe
    function showTypingIndicator() {
        if (typingStatus) {
            typingStatus.textContent = `${conversations[currentConversation]?.name} est en train d'écrire...`;
        }
    }
    
    // Masquer l'indicateur de frappe
    function hideTypingIndicator() {
        if (typingStatus) {
            typingStatus.textContent = '';
        }
    }
    
    // Simuler une réponse
    function simulateResponse() {
        const responses = [
            "Intéressant ! 🤔",
            "Complètement d'accord avec toi",
            "J'ai hâte d'écouter ça !",
            "Tu as vraiment bon goût musical 👌",
            "On en reparle ce soir ?",
            "Excellente recommandation !",
            "Je vais tester ça tout de suite 🎧"
        ];
        
        const randomResponse = responses[Math.floor(Math.random() * responses.length)];
        const conversation = conversations[currentConversation];
        
        if (conversation) {
            const responseMessage = {
                type: 'received',
                content: randomResponse,
                time: getCurrentTime()
            };
            
            conversation.messages.push(responseMessage);
            
            const messageElement = createMessageElement(responseMessage, conversation);
            const messageGroup = chatMessages.querySelector('.message-group');
            messageGroup.appendChild(messageElement);
            
            scrollToBottom();
            updateConversationPreview(currentConversation, randomResponse);
        }
    }
    
    // Mettre à jour l'aperçu de conversation
    function updateConversationPreview(conversationId, lastMessage) {
        const conversationItem = document.querySelector(`[data-conversation="${conversationId}"]`);
        if (conversationItem) {
            const lastMessageElement = conversationItem.querySelector('.last-message');
            const timeElement = conversationItem.querySelector('.message-time');
            
            if (lastMessageElement) {
                lastMessageElement.textContent = lastMessage;
            }
            if (timeElement) {
                timeElement.textContent = 'À l\'instant';
            }
        }
    }
    
    // Recherche de conversations
    function searchConversations() {
        const searchTerm = searchInput.value.toLowerCase();
        
        conversationItems.forEach(item => {
            const name = item.querySelector('h4').textContent.toLowerCase();
            const lastMessage = item.querySelector('.last-message').textContent.toLowerCase();
            
            if (name.includes(searchTerm) || lastMessage.includes(searchTerm)) {
                item.style.display = 'flex';
                item.style.opacity = '1';
            } else {
                item.style.opacity = '0';
                setTimeout(() => {
                    if (!name.includes(searchTerm) && !lastMessage.includes(searchTerm)) {
                        item.style.display = 'none';
                    }
                }, 200);
            }
        });
    }
    
    // Actions du chat
    function handleChatAction(e) {
        const btn = e.currentTarget;
        const title = btn.getAttribute('title');
        
        // Animation de feedback
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(title) {
            case 'Appel vocal':
                showNotification('Appel vocal démarré 📞', 'info');
                break;
            case 'Écoute partagée':
                startListenTogether();
                break;
            case 'Partager une playlist':
                openMusicShareModal();
                break;
            case 'Plus d\'options':
                showChatOptionsMenu(btn);
                break;
        }
    }
    
    // Actions de playlist
    function handlePlaylistAction(e) {
        const btn = e.currentTarget;
        const action = btn.textContent.includes('Écouter') ? 'play' : 'save';
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        if (action === 'play') {
            showNotification('Lecture de la playlist démarrée 🎵', 'music');
            btn.innerHTML = '⏸️ Pause';
        } else {
            showNotification('Playlist sauvegardée ! 💾', 'success');
            btn.innerHTML = '✅ Sauvegardée';
        }
    }
    
    // Démarrer l'écoute ensemble
    function startListenTogether() {
        showNotification(`Session d'écoute partagée avec ${conversations[currentConversation]?.name} ! 🎧`, 'music');
        
        // Mettre à jour le bouton
        const btn = document.querySelector('.listen-together-btn');
        if (btn) {
            btn.innerHTML = '🔴 En session';
            btn.style.background = 'linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%)';
        }
    }
    
    // Ouvrir le modal nouvelle conversation
    function openNewConversationModal() {
        const modal = document.getElementById('newConversationModal');
        if (modal) {
            modal.classList.add('active');
            
            // Focus sur la recherche
            const searchInput = modal.querySelector('.search-users-input');
            if (searchInput) {
                searchInput.focus();
            }
            
            // Attacher les événements
            attachModalEvents(modal);
        }
    }
    
    // Ouvrir le modal de partage de musique
    function openMusicShareModal() {
        const modal = document.getElementById('musicShareModal');
        if (modal) {
            modal.classList.add('active');
            
            // Focus sur la recherche
            const searchInput = modal.querySelector('.music-search-input');
            if (searchInput) {
                searchInput.focus();
            }
            
            // Attacher les événements
            attachModalEvents(modal);
        }
    }
    
    // Attacher les événements des modaux
    function attachModalEvents(modal) {
        // Fermeture du modal
        const closeBtn = modal.querySelector('.close-modal');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                modal.classList.remove('active');
            });
        }
        
        // Fermeture en cliquant à l'extérieur
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
        
        // Sélection d'utilisateur ou de musique
        const items = modal.querySelectorAll('.user-item, .music-item');
        items.forEach(item => {
            item.addEventListener('click', () => {
                if (item.classList.contains('user-item')) {
                    const userId = item.getAttribute('data-user');
                    selectUser(userId);
                } else {
                    const musicTitle = item.querySelector('h5').textContent;
                    shareMusic(musicTitle);
                }
                modal.classList.remove('active');
            });
        });
        
        // Boutons de sélection
        const selectBtns = modal.querySelectorAll('.select-music-btn');
        selectBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const musicTitle = btn.closest('.music-item').querySelector('h5').textContent;
                shareMusic(musicTitle);
                modal.classList.remove('active');
            });
        });
    }
    
    // Sélectionner un utilisateur pour nouvelle conversation
    function selectUser(userId) {
        showNotification(`Nouvelle conversation avec ${userId} créée !`, 'success');
        // Ici on pourrait ajouter la logique pour créer une nouvelle conversation
    }
    
    // Partager une musique
    function shareMusic(musicTitle) {
        const conversation = conversations[currentConversation];
        if (conversation) {
            const musicMessage = {
                type: 'sent',
                content: `🎵 Je partage avec toi: "${musicTitle}"`,
                time: getCurrentTime()
            };
            
            conversation.messages.push(musicMessage);
            
            const messageElement = createMessageElement(musicMessage, {avatar: 'MO'});
            const messageGroup = chatMessages.querySelector('.message-group');
            messageGroup.appendChild(messageElement);
            
            scrollToBottom();
            showNotification('Musique partagée ! 🎵', 'music');
        }
    }
    
    // Fonctionnalités en temps réel
    function startRealTimeFeatures() {
        // Simulation de nouveaux messages
        setInterval(() => {
            if (Math.random() < 0.1) { // 10% de chance toutes les 5 secondes
                simulateIncomingMessage();
            }
        }, 5000);
        
        // Mise à jour des statuts en ligne
        setInterval(updateOnlineStatus, 30000);
        
        // Simulation d'activité de frappe
        setInterval(() => {
            if (Math.random() < 0.05 && !isTyping) {
                showTypingIndicator();
                setTimeout(hideTypingIndicator, 3000);
            }
        }, 10000);
    }
    
    // Simuler un message entrant
    function simulateIncomingMessage() {
        const otherConversations = Object.keys(conversations).filter(id => id !== currentConversation);
        const randomConversation = otherConversations[Math.floor(Math.random() * otherConversations.length)];
        
        if (randomConversation) {
            const conversation = conversations[randomConversation];
            const messages = [
                "Tu as vu le nouveau clip ? 🎬",
                "Cette mélodie me reste en tête !",
                "Parfait pour travailler 👌",
                "Tu connais cet artiste ?",
                "Session d'écoute ce soir ?"
            ];
            
            const randomMessage = messages[Math.floor(Math.random() * messages.length)];
            
            // Ajouter un badge non lu
            const conversationItem = document.querySelector(`[data-conversation="${randomConversation}"]`);
            if (conversationItem && !conversationItem.querySelector('.unread-count')) {
                const statusDiv = conversationItem.querySelector('.conversation-status');
                const unreadBadge = document.createElement('span');
                unreadBadge.className = 'unread-count';
                unreadBadge.textContent = '1';
                statusDiv.appendChild(unreadBadge);
                
                // Animation d'apparition
                unreadBadge.style.transform = 'scale(0)';
                setTimeout(() => {
                    unreadBadge.style.transform = 'scale(1)';
                }, 100);
            }
            
            updateConversationPreview(randomConversation, randomMessage);
            showNotification(`Nouveau message de ${conversation.name}`, 'message');
        }
    }
    
    // Mettre à jour les statuts en ligne
    function updateOnlineStatus() {
        document.querySelectorAll('.online-status').forEach(status => {
            if (Math.random() < 0.1) { // 10% de chance de changer de statut
                const statuses = ['online', 'away', 'offline'];
                const currentStatus = Array.from(status.classList).find(cls => statuses.includes(cls));
                const newStatus = statuses[Math.floor(Math.random() * statuses.length)];
                
                if (currentStatus !== newStatus) {
                    status.classList.remove(currentStatus);
                    status.classList.add(newStatus);
                }
            }
        });
    }
    
    // Fonctions utilitaires
    function getCurrentTime() {
        const now = new Date();
        return now.getHours().toString().padStart(2, '0') + ':' + 
               now.getMinutes().toString().padStart(2, '0');
    }
    
    function scrollToBottom() {
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    
    function showNotification(message, type = 'info') {
        // Utiliser la fonction de notification globale si disponible
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
        } else {
            console.log(`Notification: ${message}`);
        }
    }
    
    // Raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // Échap pour fermer les modaux
        if (e.key === 'Escape') {
            document.querySelectorAll('.new-conversation-modal, .music-share-modal').forEach(modal => {
                modal.classList.remove('active');
            });
        }
        
        // Ctrl + K pour rechercher
        if (e.ctrlKey && e.key === 'k') {
            e.preventDefault();
            if (searchInput) {
                searchInput.focus();
            }
        }
        
        // Ctrl + N pour nouvelle conversation
        if (e.ctrlKey && e.key === 'n') {
            e.preventDefault();
            openNewConversationModal();
        }
    });
    
    // Gestion du redimensionnement
    window.addEventListener('resize', function() {
        scrollToBottom();
    });
    
    console.log('💬 Page Messages Rythm\'it initialisée avec succès !');
    console.log('🚀 Fonctionnalités: Chat en temps réel, Partage de musique, Écoute partagée');
});