// JavaScript pour la page Messages avec WebSocket - messages.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let ws = null;
    let currentConversationId = null;
    let currentUserId = null;
    let isTyping = false;
    let typingTimeout = null;
    let conversations = [];
    
    // Éléments DOM
    const conversationsList = document.querySelector('.conversations-list');
    const chatMessages = document.getElementById('chatMessages');
    const chatInput = document.querySelector('.chat-input');
    const sendBtn = document.querySelector('.send-message-btn');
    const chatUserInfo = document.querySelector('.chat-user-info');
    const chatStatusText = document.querySelector('.chat-user-info .user-status');
    const composerButtons = document.querySelectorAll('.chat-input-area button');
    let chatInfoPanel = document.getElementById('chatInfoPanel');
    const searchInput = document.querySelector('.search-conversations-input');
    
    // Notification sûre (évite la récursion si window.showNotification pointe ici)
    function safeNotify(message, type = 'info') {
        if (typeof window.showNotification === 'function' && window.showNotification !== safeNotify) {
            window.showNotification(message, type);
        } else {
            console.log(`[${type.toUpperCase()}] ${message}`);
        }
    }

    // Initialisation
    init();
    
    function init() {
        console.log('💬 Initialisation de la page Messages...');
        
        // Désactiver la saisie tant qu'aucune conversation n'est prête
        setComposerEnabled(false);

        // S'assurer que le panneau d'infos existe (au cas où le template ne l'aurait pas)
        ensureChatInfoPanel();

        // Charger les conversations
        loadConversations();
        
        // Connecter au WebSocket
        connectWebSocket();
        
        // Attacher les événements
        attachEventListeners();
        
        // Gérer les paramètres URL (si on vient d'un profil par exemple)
        handleURLParams();
    }

    function ensureChatInfoPanel() {
        if (!chatInfoPanel) {
            const sidebar = document.querySelector('.chat-sidebar');
            if (sidebar) {
                sidebar.innerHTML = `
                    <div class="chat-info-section" id="chatInfoPanel">
                        <div class="chat-info-empty">
                            <h4>Infos conversation</h4>
                            <p>Quand une conversation est ouverte, les infos apparaîtront ici.</p>
                        </div>
                    </div>
                `;
                chatInfoPanel = document.getElementById('chatInfoPanel');
            }
        }
    }

    function attachEventListeners() {
        if (sendBtn) {
            sendBtn.addEventListener('click', sendMessage);
        }
        
        if (chatInput) {
            chatInput.addEventListener('input', handleInputChange);
            chatInput.addEventListener('keypress', function(e) {
                if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    sendMessage();
                }
            });
        }
        if (searchInput) {
            searchInput.addEventListener('input', (e) => filterConversations(e.target.value));
        }
    }
    
    // Connexion WebSocket
    function connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/messages`;
        
        console.log('🔌 Connexion WebSocket:', wsUrl);
        
        ws = new WebSocket(wsUrl);
        
        ws.onopen = function() {
            console.log('✅ WebSocket connecté');
            showNotification('✅ Connexion établie', 'success');
        };
        
        ws.onmessage = function(event) {
            try {
                const message = JSON.parse(event.data);
                console.log('📥 Message WebSocket reçu:', message);
                handleWebSocketMessage(message);
            } catch (error) {
                console.error('❌ Erreur parsing message WebSocket:', error);
            }
        };
        
        ws.onerror = function(error) {
            console.error('❌ Erreur WebSocket:', error);
            showNotification('❌ Erreur de connexion', 'error');
        };
        
        ws.onclose = function() {
            console.log('❌ WebSocket déconnecté, reconnexion dans 5s...');
            showNotification('⚠️ Connexion perdue, reconnexion...', 'warning');
            setTimeout(connectWebSocket, 5000);
        };
    }
    
    // Gérer les messages WebSocket
    function handleWebSocketMessage(wsMsg) {
        switch (wsMsg.type) {
            case 'message':
                // Nouveau message reçu
                if (wsMsg.message) {
                    addMessageToChat(wsMsg.message);
                    updateConversationPreview(wsMsg.message);
                    moveConversationToTop(wsMsg.message.conversation_id);
                    updateChatHeader(wsMsg.message.conversation_id);
                    updateSidebarInfo(wsMsg.message.conversation_id);
                    
                    // Marquer comme lu si c'est la conversation active
                    if (wsMsg.message.conversation_id === currentConversationId) {
                        markConversationAsRead(currentConversationId);
                    } else {
                        // Notification si pas sur cette conversation
                        showDesktopNotification(wsMsg.message);
                    }
                }
                break;
                
            case 'typing':
                // L'autre personne est en train d'écrire
                if (wsMsg.data && wsMsg.data.conversation_id === currentConversationId) {
                    updateTypingIndicator(wsMsg.data.is_typing, wsMsg.data.user_id);
                }
                break;
                
            case 'read':
                // Messages marqués comme lus
                if (wsMsg.data && wsMsg.data.conversation_id === currentConversationId) {
                    updateReadStatus();
                }
                break;
                
            case 'status':
                // Changement de statut (en ligne, hors ligne)
                if (wsMsg.data) {
                    updateUserStatus(wsMsg.data.user_id, wsMsg.data.status);
                }
                break;
                
            case 'user_online':
                // Utilisateur vient de se connecter
                if (wsMsg.data && wsMsg.data.user_id) {
                    updateUserOnlineStatus(wsMsg.data.user_id, true);
                }
                break;
                
            case 'user_offline':
                // Utilisateur vient de se déconnecter
                if (wsMsg.data && wsMsg.data.user_id) {
                    updateUserOnlineStatus(wsMsg.data.user_id, false);
                }
                break;
        }
    }
    
    // Charger les conversations
    async function loadConversations() {
        try {
            console.log('📥 Chargement des conversations...');
            
            const response = await fetch('/api/conversations', {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            console.log('✅ Conversations chargées:', data);
            
            if (data.success && data.data && data.data.conversations) {
                conversations = data.data.conversations;

                // Déterminer l'ID utilisateur courant à partir de la première conversation (l'API ne le fournit pas directement)
                if (conversations.length > 0 && !currentUserId) {
                    const first = conversations[0];
                    const otherId = first.other_user?.id;
                    if (otherId && first.user1_id && first.user2_id) {
                        currentUserId = (first.user1_id === otherId) ? first.user2_id : first.user1_id;
                    }
                }

                displayConversations(conversations);
                
                // Charger la première conversation si elle existe
                if (conversations.length > 0 && !currentConversationId) {
                    await loadConversation(conversations[0].id);
                } else if (conversations.length === 0) {
                    // Aucune conversation existante : tenter de les initier à partir des amis
                    await ensureConversationsFromFriends();
                    // Recharger la liste après création
                    await loadConversations();
                    setComposerEnabled(false);
                }
            } else {
                displayEmptyConversations();
                setComposerEnabled(false);
            }
        } catch (error) {
            console.error('❌ Erreur chargement conversations:', error);
            showNotification('❌ Erreur de chargement', 'error');
            setComposerEnabled(false);
        }
    }
    
    // Afficher les conversations
    function displayConversations(conversations) {
        if (!conversationsList) return;
        
        if (conversations.length === 0) {
            displayEmptyConversations();
            return;
        }
        
        conversationsList.innerHTML = conversations.map(conv => {
            const otherUser = conv.other_user;
            const username = otherUser?.username || 'Utilisateur';
            const avatar = otherUser?.profile_pic;
            const unreadCount = conv.unread_count || 0;
            const lastMessage = conv.last_message_text || 'Nouvelle conversation';
            const timeAgo = formatTimeAgo(conv.last_message_at || conv.created_at);
            const isOnline = conv.is_online;
            
            const avatarHTML = avatar 
                ? `<img src="${avatar}" alt="${username}" class="user-avatar">` 
                : `<div class="user-pic small">${username.substring(0, 2).toUpperCase()}</div>`;
            
            return `
                <div class="conversation-item" data-conversation="${conv.id}">
                    <div class="conversation-avatar">
                        ${avatarHTML}
                        <span class="online-status ${isOnline ? 'online' : 'offline'}"></span>
                    </div>
                    <div class="conversation-info">
                        <h4>${username}</h4>
                        <p class="last-message">${lastMessage}</p>
                        <span class="message-time">${timeAgo}</span>
                    </div>
                    ${unreadCount > 0 ? `
                        <div class="conversation-status">
                            <span class="unread-count">${unreadCount}</span>
                        </div>
                    ` : ''}
                </div>
            `;
        }).join('');
        
        // Attacher les événements
        attachConversationListeners();
    }
    
    // Afficher état vide
    function displayEmptyConversations() {
        if (!conversationsList) return;
        
        conversationsList.innerHTML = `
            <div class="empty-state">
                <div class="empty-icon">💬</div>
                <h3>Aucune conversation</h3>
                <p>Ajoutez des amis pour démarrer une conversation !</p>
                <a href="/discover" class="btn-primary">Trouver des amis</a>
                </div>
            `;
        }
        
    // Attacher les événements aux conversations
    function attachConversationListeners() {
        document.querySelectorAll('.conversation-item').forEach(item => {
            item.addEventListener('click', function() {
                const convId = parseInt(this.getAttribute('data-conversation'));
                loadConversation(convId);
            });
        });
    }
    
    // Charger une conversation
    async function loadConversation(conversationId) {
        try {
            console.log('💬 Chargement conversation:', conversationId);
            setComposerEnabled(false);
            
            currentConversationId = conversationId;
            
            // Mettre à jour l'UI
            document.querySelectorAll('.conversation-item').forEach(item => {
                item.classList.remove('active');
                if (parseInt(item.getAttribute('data-conversation')) === conversationId) {
                    item.classList.add('active');
                }
            });
            
            // Charger les messages
            const response = await fetch(`/api/conversations/${conversationId}/messages?limit=50`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            console.log('✅ Messages chargés:', data);
            
            if (data.success && data.data && data.data.messages) {
                displayMessages(data.data.messages);
                
                // Marquer comme lu
                markConversationAsRead(conversationId);
                
                // Mettre à jour le header avec les infos de l'utilisateur
                updateChatHeader(conversationId);
                
                // Mettre à jour la sidebar avec les infos
                updateSidebarInfo(conversationId);

                // Activer l'input maintenant que la conversation est prête
                setComposerEnabled(true);
                return;
            }
        } catch (error) {
            console.error('❌ Erreur chargement conversation:', error);
            showNotification('❌ Erreur de chargement', 'error');
            setComposerEnabled(false);
        }

        // Fallback: si aucune donnée chargée mais la sélection a eu lieu, activer quand même l'UI
        setComposerEnabled(true);
        updateChatHeader(conversationId);
        updateSidebarInfo(conversationId);
    }
    
    // Afficher les messages
    function displayMessages(messages) {
        if (!chatMessages) return;
        
        if (messages.length === 0) {
            chatMessages.innerHTML = `
                <div class="empty-messages">
                    <p>🎵 Démarrez la conversation en envoyant un message !</p>
                </div>
            `;
            return;
        }
        
        chatMessages.innerHTML = messages.map(msg => createMessageHTML(msg)).join('');
        
        // Scroll vers le bas
        scrollToBottom();
    }
    
    // Créer le HTML d'un message
    function createMessageHTML(message) {
        const isSent = message.sender_id === currentUserId;
        const messageClass = isSent ? 'sent' : 'received';
        const timeAgo = formatTimeShort(message.created_at);
        
        return `
            <div class="message ${messageClass}" data-sender-id="${message.sender_id}" data-is-read="${message.is_read}">
                ${!isSent ? `
                    <div class="message-avatar">
                        <div class="user-pic tiny">U</div>
                    </div>
                ` : ''}
                <div class="message-content">
                    <div class="message-bubble">
                        ${escapeHTML(message.content)}
                    </div>
                    <span class="message-timestamp">${timeAgo}</span>
                </div>
            </div>
        `;
    }
    
    // Ajouter un message au chat
    function addMessageToChat(message) {
        if (!chatMessages) return;
        if (message.conversation_id !== currentConversationId) return;
        
        const messageHTML = createMessageHTML(message);
        chatMessages.insertAdjacentHTML('beforeend', messageHTML);
        scrollToBottom();
    }
    
    // Envoyer un message
    async function sendMessage() {
        if (!chatInput) return;
        if (!currentConversationId) {
            showNotification('Choisis une conversation avant d\'envoyer un message', 'warning');
            setComposerEnabled(false);
            return;
        }
        
        const content = chatInput.value.trim();
        if (!content) return;
        
        try {
            // Récupérer l'ID du destinataire
            const conv = conversations.find(c => c.id === currentConversationId);
            if (!conv) return;
            
            const receiverId = conv.other_user.id;
            
            // Envoyer via API (permet la persistance et la diffusion WS côté serveur)
            const response = await fetch('/api/messages/send', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({
                    receiver_id: receiverId,
                    content: content
                })
            });

            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }

            // Le backend diffusera le message via WebSocket
            chatInput.value = '';
            updateTypingStatus(false);
        } catch (error) {
            console.error('❌ Erreur envoi message:', error);
            showNotification('❌ Erreur d\'envoi', 'error');
        }
    }
    
    // Marquer comme lu
    async function markConversationAsRead(conversationId) {
        try {
            await fetch(`/api/conversations/${conversationId}/read`, {
                method: 'POST',
                credentials: 'include'
            });
            
            // Supprimer le badge non lu
            const convItem = document.querySelector(`[data-conversation="${conversationId}"]`);
            if (convItem) {
                const unreadBadge = convItem.querySelector('.unread-count');
                if (unreadBadge) {
                    unreadBadge.remove();
                }
            }
        } catch (error) {
            console.error('❌ Erreur marquage lu:', error);
        }
    }
    
    // Mettre à jour le statut "en train d'écrire"
    function updateTypingStatus(isTyping) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        if (!currentConversationId) return;
        
        const wsMsg = {
            type: 'typing',
            data: {
                conversation_id: currentConversationId,
                is_typing: isTyping
            }
        };
        
        ws.send(JSON.stringify(wsMsg));
    }
    
    // Afficher l'indicateur "en train d'écrire"
    function updateTypingIndicator(isTyping, userId) {
        // Ajouter ou supprimer l'indicateur dans le chat
        const typingIndicator = document.querySelector('.typing-indicator-message');
        
        if (isTyping && userId !== currentUserId) {
            if (!typingIndicator) {
                const html = `
                    <div class="message received typing-indicator-message">
                        <div class="message-avatar">
                            <div class="user-pic tiny">U</div>
                        </div>
                        <div class="message-content">
                            <div class="typing-indicator">
                                <div class="typing-dot"></div>
                                <div class="typing-dot"></div>
                                <div class="typing-dot"></div>
                            </div>
                        </div>
                    </div>
                `;
                chatMessages.insertAdjacentHTML('beforeend', html);
                scrollToBottom();
            }
        } else {
            if (typingIndicator) {
                typingIndicator.remove();
            }
        }
    }
    
    // Gérer la saisie dans l'input
    function handleInputChange() {
        if (!isTyping) {
            isTyping = true;
            updateTypingStatus(true);
        }
        
        clearTimeout(typingTimeout);
        typingTimeout = setTimeout(() => {
            isTyping = false;
            updateTypingStatus(false);
        }, 2000);
    }
    
    // Gérer les paramètres URL
    function handleURLParams() {
        const urlParams = new URLSearchParams(window.location.search);
        const userId = urlParams.get('user');
        
        if (userId) {
            // Créer une conversation avec cet utilisateur
            createConversationWithUser(parseInt(userId));
        }
    }
    
    // Créer une conversation avec un utilisateur
    async function createConversationWithUser(userId) {
        try {
            const response = await fetch(`/api/conversations/${userId}`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.success && data.data && data.data.conversation) {
                // Recharger les conversations et ouvrir celle-ci
                await loadConversations();
                loadConversation(data.data.conversation.id);
            }
        } catch (error) {
            console.error('❌ Erreur création conversation:', error);
            showNotification('❌ Impossible de démarrer la conversation', 'error');
        }
    }
    
    // Fonctions utilitaires
    function formatTimeAgo(dateString) {
        if (!dateString) return 'Récemment';
        
        const date = new Date(dateString);
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) return 'À l\'instant';
        if (diffInSeconds < 3600) return `il y a ${Math.floor(diffInSeconds / 60)}min`;
        if (diffInSeconds < 86400) return `il y a ${Math.floor(diffInSeconds / 3600)}h`;
        if (diffInSeconds < 2592000) return `il y a ${Math.floor(diffInSeconds / 86400)}j`;
        
        return date.toLocaleDateString('fr-FR');
    }
    
    function formatTimeShort(dateString) {
        if (!dateString) return '';
        const date = new Date(dateString);
        return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
    }
    
    function escapeHTML(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
    
    function scrollToBottom() {
        if (chatMessages) {
            chatMessages.scrollTop = chatMessages.scrollHeight;
        }
    }

    function setComposerEnabled(enabled) {
        if (chatInput) {
            chatInput.disabled = !enabled;
            chatInput.classList.toggle('disabled', !enabled);
        }
        if (sendBtn) {
            sendBtn.disabled = !enabled;
        }
        if (composerButtons) {
            composerButtons.forEach(btn => {
                btn.disabled = !enabled;
            });
        }
    }

    // Créer automatiquement des conversations pour les amis si aucune conversation n'existe
    async function ensureConversationsFromFriends() {
        try {
            const resp = await fetch('/api/friends', { credentials: 'include' });
            if (!resp.ok) return;
            const data = await resp.json();
            const friends = data?.data?.friends || data?.friends || [];
            if (!Array.isArray(friends) || friends.length === 0) return;

            // Créer une conversation pour chaque ami manquant (best effort)
            for (const friend of friends) {
                if (!friend.id) continue;
                try {
                    await fetch(`/api/conversations/${friend.id}`, { credentials: 'include' });
                } catch (e) {
                    console.warn('⚠️ Impossible de créer la conversation pour', friend.id, e);
                }
            }
        } catch (error) {
            console.warn('⚠️ Impossible de pré-créer les conversations', error);
        }
    }
    
    function updateChatHeader(conversationId) {
        const conv = conversations.find(c => c.id === conversationId);
        if (conv && chatUserInfo) {
            const username = conv.other_user?.username || 'Utilisateur';
            chatUserInfo.querySelector('h3').textContent = username;
            if (chatStatusText) {
                chatStatusText.textContent = computeDeliveryStatus();
            }
            return;
        }

        // Fallback: récupérer le titre depuis la liste DOM si présent
        const item = document.querySelector(`[data-conversation="${conversationId}"] h4`);
        if (item && chatUserInfo) {
            chatUserInfo.querySelector('h3').textContent = item.textContent.trim();
            if (chatStatusText) chatStatusText.textContent = 'En attente';
        }
    }
    
    function updateConversationPreview(message) {
        // Mettre à jour la preview dans la liste
        const convItem = document.querySelector(`[data-conversation="${message.conversation_id}"]`);
        if (convItem) {
            const lastMessageEl = convItem.querySelector('.last-message');
            if (lastMessageEl) {
                lastMessageEl.textContent = message.content.substring(0, 50) + '...';
            }
        }
        
        // Mettre à jour dans le tableau conversations aussi
        const conv = conversations.find(c => c.id === message.conversation_id);
        if (conv) {
            conv.last_message_text = message.content;
            conv.last_message_at = message.created_at;
        }
    }
    
    function moveConversationToTop(conversationId) {
        // Déplacer la conversation en haut de la liste
        const convIndex = conversations.findIndex(c => c.id === conversationId);
        if (convIndex > 0) {
            const [conv] = conversations.splice(convIndex, 1);
            conversations.unshift(conv);
            displayConversations(conversations);
        }
    }
    
    function showDesktopNotification(message) {
        // Notification navigateur si pas sur la page messages
        if (document.hidden && 'Notification' in window) {
            if (Notification.permission === 'granted') {
                const conv = conversations.find(c => c.id === message.conversation_id);
                const senderName = conv?.other_user?.username || 'Un utilisateur';
                new Notification(`💬 ${senderName}`, {
                    body: message.content,
                    icon: conv?.other_user?.profile_pic || '/styles/images/logo.png'
                });
            } else if (Notification.permission !== 'denied') {
                Notification.requestPermission();
            }
        }
        
        // Notification visuelle dans l'app
        if (message.sender_id !== currentUserId) {
            showNotification(`💬 Nouveau message reçu`, 'info');
        }
    }
    
    function updateUserOnlineStatus(userId, isOnline) {
        // Mettre à jour le statut en ligne dans toutes les conversations
        conversations.forEach(conv => {
            if (conv.other_user && conv.other_user.id === userId) {
                conv.is_online = isOnline;
            }
        });
        
        // Mettre à jour l'UI si c'est la conversation active
        const activeConv = conversations.find(c => c.id === currentConversationId);
        if (activeConv && activeConv.other_user && activeConv.other_user.id === userId) {
            updateSidebarInfo(currentConversationId);
            
            // Mettre à jour l'indicateur dans la liste
            const convItem = document.querySelector(`[data-conversation="${currentConversationId}"]`);
            if (convItem) {
                const statusIndicator = convItem.querySelector('.online-status');
                if (statusIndicator) {
                    statusIndicator.className = `online-status ${isOnline ? 'online' : 'offline'}`;
                }
            }
        }
    }
    
    function updateReadStatus() {
        // Mettre à jour les doubles checks ou autre indicateur
        console.log('✅ Messages marqués comme lus');
    }
    
    function updateUserStatus(userId, status) {
        // Mettre à jour le statut en ligne/hors ligne
        console.log(`👤 Statut utilisateur ${userId}: ${status}`);
        if (currentConversationId) {
            updateSidebarInfo(currentConversationId);
        }
    }
    
    function showNotification(message, type = 'info') {
        safeNotify(message, type);
    }

    // Calculer le statut "Envoyé/Lu/Reçu" sur la conversation active
    function computeDeliveryStatus() {
        try {
            const lastMessage = chatMessages?.querySelector('.message:last-child');
            let status = 'En attente';

            if (lastMessage && lastMessage.dataset) {
                const senderId = parseInt(lastMessage.dataset.senderId || '0');
                const isRead = lastMessage.dataset.isRead === 'true';
                if (senderId && currentUserId && senderId === currentUserId) {
                    status = isRead ? 'Lu' : 'Envoyé';
                } else if (senderId) {
                    status = 'Reçu';
                }
            }

            return status;
        } catch (e) {
            return 'En attente';
        }
    }

    // Mettre à jour la colonne d'info conversation
    function updateSidebarInfo(conversationId) {
        if (!chatInfoPanel) return;
        const conv = conversations.find(c => c.id === conversationId);
        if (!conv || !conv.other_user) {
            chatInfoPanel.innerHTML = `
                <div class="chat-info-empty">
                    <h4>Infos conversation</h4>
                    <p>Quand une conversation est ouverte, les infos apparaîtront ici.</p>
                </div>
            `;
            return;
        }

        const user = conv.other_user;
        const avatar = user.profile_pic
            ? `<img src="${user.profile_pic}" alt="${user.username}" class="user-avatar large">`
            : `<div class="user-pic large">${(user.username || '?').substring(0,2).toUpperCase()}</div>`;

        const statusText = computeDeliveryStatus();
        const onlineClass = conv.is_online ? 'online' : 'offline';
        const timeAgo = formatTimeAgo(conv.last_message_at || conv.created_at);

        chatInfoPanel.innerHTML = `
            <div class="chat-info-card">
                <div class="chat-info-header">
                    <div class="chat-info-avatar">
                        ${avatar}
                        <span class="online-status ${onlineClass}"></span>
                    </div>
                    <div class="chat-info-meta">
                        <h4>${user.username || 'Utilisateur'}</h4>
                        <p>@${user.username || 'user'}</p>
                        <span class="chat-info-status">${statusText}</span>
                    </div>
                </div>
                <div class="chat-info-body">
                    <div class="chat-info-row">
                        <span>Dernier message</span>
                        <strong>${conv.last_message_text || '—'}</strong>
                    </div>
                    <div class="chat-info-row">
                        <span>Activité</span>
                        <strong>${timeAgo}</strong>
                    </div>
                    <div class="chat-info-row">
                        <span>Statut</span>
                        <strong>${conv.is_online ? 'En ligne' : 'Hors ligne'}</strong>
                    </div>
                </div>
            </div>
        `;
    }

    // Recherche dans les conversations
    function filterConversations(query) {
        const q = (query || '').toLowerCase();
        if (!q) {
            displayConversations(conversations);
            return;
        }
        const filtered = conversations.filter(conv => {
            const name = conv.other_user?.username || '';
            return name.toLowerCase().includes(q);
        });
        if (filtered.length === 0) {
            conversationsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">🔍</div>
                    <h3>Aucune conversation</h3>
                    <p>Aucun résultat pour "${query}".</p>
                </div>
            `;
            return;
        }
        displayConversations(filtered);
    }
    
    console.log('💬 Page Messages initialisée avec succès !');
});
