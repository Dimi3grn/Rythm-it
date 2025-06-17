// JavaScript pour la page amis - friends.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    const messageModal = document.getElementById('messageModal');
    const modalUserName = document.getElementById('modalUserName');
    const searchInput = document.querySelector('.search-input');
    const addFriendBtn = document.querySelector('.add-friend-btn');
    const friendsGrid = document.querySelector('.friends-grid');
    
    // Charger les données initiales
    loadFriends();
    loadFriendRequests();
    loadFriendshipStats();
    
    // Gestion de la recherche d'amis
    if (searchInput) {
        let searchTimeout;
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            const searchTerm = this.value.trim();
            
            searchTimeout = setTimeout(() => {
                if (searchTerm.length >= 2) {
                    searchUsers(searchTerm);
                } else {
                    loadFriends(); // Recharger les amis si recherche vide
                }
            }, 300);
        });
    }
    
    // Gestion du bouton "Ajouter un ami"
    if (addFriendBtn) {
        addFriendBtn.addEventListener('click', function() {
            showAddFriendModal();
        });
    }
    
    // Charger la liste des amis
    async function loadFriends() {
        try {
            const response = await fetch('/api/friends', {
                method: 'GET',
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                displayFriends(data.friends || []);
            } else {
                console.error('Erreur chargement amis:', response.statusText);
                showNotification('Erreur lors du chargement des amis', 'error');
            }
        } catch (error) {
            console.error('Erreur:', error);
            showNotification('Erreur de connexion', 'error');
        }
    }
    
    // Afficher la liste des amis
    function displayFriends(friends) {
        if (!friendsGrid) return;
        
        if (friends.length === 0) {
            friendsGrid.innerHTML = `
                <div class="no-friends">
                    <h3>Aucun ami pour le moment</h3>
                    <p>Utilisez la recherche pour trouver des amis !</p>
                </div>
            `;
            return;
        }
        
        friendsGrid.innerHTML = friends.map(friend => createFriendCard(friend)).join('');
        
        // Ajouter les event listeners
        addFriendCardListeners();
    }
    
    // Créer une carte d'ami
    function createFriendCard(friend) {
        const avatar = friend.avatar ? `<img src="${friend.avatar}" alt="${friend.username}">` : 
                      `<div class="user-pic">${friend.username.substring(0, 2).toUpperCase()}</div>`;
        
        const onlineStatusClass = friend.online_status || 'offline';
        const activity = friend.activity || '';
        const mutualFriends = friend.mutual_friends || 0;
        
        return `
            <div class="friend-card" data-user-id="${friend.id}" data-status="${onlineStatusClass}">
                <div class="friend-card-header">
                    <div class="friend-avatar">
                        ${avatar}
                        <span class="online-status ${onlineStatusClass}"></span>
                    </div>
                    <div class="friend-menu" data-friend-id="${friend.id}">⋯</div>
                </div>
                <div class="friend-info">
                    <h3>${friend.username}</h3>
                    <p class="friend-username">@${friend.username}</p>
                    ${activity ? `<p class="friend-activity">${activity}</p>` : ''}
                </div>
                <div class="friend-stats">
                    <div class="stat">
                        <span class="stat-number">${mutualFriends}</span>
                        <span class="stat-label">Amis mutuels</span>
                    </div>
                    <div class="stat">
                        <span class="stat-number">${formatDate(friend.friendship_date)}</span>
                        <span class="stat-label">Amis depuis</span>
                    </div>
                </div>
                <div class="friend-actions">
                    <button class="friend-btn message-btn" data-friend="${friend.username}">💬 Message</button>
                    <button class="friend-btn profile-btn" data-user-id="${friend.id}">👤 Profil</button>
                </div>
            </div>
        `;
    }
    
    // Rechercher des utilisateurs
    async function searchUsers(query) {
        try {
            const response = await fetch(`/api/users/search?q=${encodeURIComponent(query)}`, {
                method: 'GET',
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                displaySearchResults(data.users || []);
            } else {
                console.error('Erreur recherche:', response.statusText);
            }
        } catch (error) {
            console.error('Erreur recherche:', error);
        }
    }
    
    // Afficher les résultats de recherche
    function displaySearchResults(users) {
        if (!friendsGrid) return;
        
        if (users.length === 0) {
            friendsGrid.innerHTML = `
                <div class="no-results">
                    <h3>Aucun utilisateur trouvé</h3>
                    <p>Essayez avec un autre nom d'utilisateur</p>
                </div>
            `;
            return;
        }
        
        friendsGrid.innerHTML = users.map(user => createUserSearchCard(user)).join('');
        
        // Ajouter les event listeners
        addSearchCardListeners();
    }
    
    // Créer une carte de résultat de recherche
    function createUserSearchCard(user) {
        const avatar = user.avatar ? `<img src="${user.avatar}" alt="${user.username}">` : 
                      `<div class="user-pic">${user.username.substring(0, 2).toUpperCase()}</div>`;
        
        const mutualFriends = user.mutual_friends || 0;
        const friendshipStatus = user.friendship_status;
        
        let actionButton = '';
        if (!friendshipStatus) {
            actionButton = `<button class="friend-btn add-btn" data-user-id="${user.id}">➕ Ajouter</button>`;
        } else if (friendshipStatus === 'pending') {
            actionButton = `<button class="friend-btn pending-btn" disabled>⏳ En attente</button>`;
        } else if (friendshipStatus === 'accepted') {
            actionButton = `<button class="friend-btn friends-btn" disabled>✅ Amis</button>`;
        } else if (friendshipStatus === 'blocked') {
            actionButton = `<button class="friend-btn blocked-btn" disabled>🚫 Bloqué</button>`;
        }
        
        return `
            <div class="friend-card search-result" data-user-id="${user.id}">
                <div class="friend-card-header">
                    <div class="friend-avatar">
                        ${avatar}
                    </div>
                </div>
                <div class="friend-info">
                    <h3>${user.username}</h3>
                    <p class="friend-username">@${user.username}</p>
                    ${mutualFriends > 0 ? `<p class="mutual-friends">${mutualFriends} amis mutuels</p>` : ''}
                </div>
                <div class="friend-actions">
                    ${actionButton}
                    <button class="friend-btn profile-btn" data-user-id="${user.id}">👤 Profil</button>
                </div>
            </div>
        `;
    }
    
    // Ajouter les event listeners pour les cartes d'amis
    function addFriendCardListeners() {
        // Boutons de message
        document.querySelectorAll('.message-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const friendName = this.getAttribute('data-friend');
                openMessageModal(friendName);
            });
        });
        
        // Menus d'options
        document.querySelectorAll('.friend-menu').forEach(menu => {
            menu.addEventListener('click', function(e) {
                e.stopPropagation();
                const friendId = this.getAttribute('data-friend-id');
                showFriendMenu(this, friendId);
            });
        });
        
        // Boutons de profil
        document.querySelectorAll('.profile-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const userId = this.getAttribute('data-user-id');
                // Rediriger vers le profil de l'utilisateur
                window.location.href = `/profile?user=${userId}`;
            });
        });
    }
    
    // Ajouter les event listeners pour les cartes de recherche
    function addSearchCardListeners() {
        // Boutons d'ajout d'ami
        document.querySelectorAll('.add-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const userId = parseInt(this.getAttribute('data-user-id'));
                sendFriendRequest(userId, this);
            });
        });
        
        // Boutons de profil
        document.querySelectorAll('.profile-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const userId = this.getAttribute('data-user-id');
                window.location.href = `/profile?user=${userId}`;
            });
        });
    }
    
    // Envoyer une demande d'amitié
    async function sendFriendRequest(userId, button) {
        try {
            button.disabled = true;
            button.textContent = '⏳ Envoi...';
            
            const response = await fetch('/api/friends/request', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    addressee_id: userId
                })
            });
            
            if (response.ok) {
                button.textContent = '⏳ En attente';
                button.className = 'friend-btn pending-btn';
                showNotification('Demande d\'amitié envoyée !', 'success');
            } else {
                const data = await response.json();
                showNotification(data.error || 'Erreur lors de l\'envoi', 'error');
                button.disabled = false;
                button.textContent = '➕ Ajouter';
            }
        } catch (error) {
            console.error('Erreur envoi demande:', error);
            showNotification('Erreur de connexion', 'error');
            button.disabled = false;
            button.textContent = '➕ Ajouter';
        }
    }
    
    // Charger les demandes d'amitié
    async function loadFriendRequests() {
        try {
            const response = await fetch('/api/friends/requests', {
                method: 'GET',
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                displayFriendRequests(data.requests || []);
            }
        } catch (error) {
            console.error('Erreur chargement demandes:', error);
        }
    }
    
    // Afficher les demandes d'amitié
    function displayFriendRequests(requests) {
        if (requests.length === 0) return;
        
        // Créer une section pour les demandes si elle n'existe pas
        let requestsSection = document.querySelector('.friend-requests-section');
        if (!requestsSection) {
            requestsSection = document.createElement('div');
            requestsSection.className = 'friend-requests-section';
            requestsSection.innerHTML = `
                <h2>Demandes d'amitié (${requests.length})</h2>
                <div class="friend-requests-grid"></div>
            `;
            
            const mainContent = document.querySelector('.content-area');
            if (mainContent) {
                mainContent.insertBefore(requestsSection, friendsGrid);
            }
        }
        
        const requestsGrid = requestsSection.querySelector('.friend-requests-grid');
        if (requestsGrid) {
            requestsGrid.innerHTML = requests.map(request => createFriendRequestCard(request)).join('');
            addRequestCardListeners();
        }
    }
    
    // Créer une carte de demande d'amitié
    function createFriendRequestCard(request) {
        const avatar = request.requester_avatar ? 
                      `<img src="${request.requester_avatar}" alt="${request.requester_username}">` : 
                      `<div class="user-pic">${request.requester_username.substring(0, 2).toUpperCase()}</div>`;
        
        return `
            <div class="friend-request-card" data-request-id="${request.id}" data-requester-id="${request.requester_id}">
                <div class="request-avatar">
                    ${avatar}
                </div>
                <div class="request-info">
                    <h4>${request.requester_username}</h4>
                    <p>Demande d'amitié</p>
                    <small>${formatDate(request.created_at)}</small>
                </div>
                <div class="request-actions">
                    <button class="accept-btn" data-requester-id="${request.requester_id}">✅ Accepter</button>
                    <button class="reject-btn" data-requester-id="${request.requester_id}">❌ Refuser</button>
                </div>
            </div>
        `;
    }
    
    // Ajouter les event listeners pour les demandes
    function addRequestCardListeners() {
        document.querySelectorAll('.accept-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const requesterId = parseInt(this.getAttribute('data-requester-id'));
                acceptFriendRequest(requesterId, this);
            });
        });
        
        document.querySelectorAll('.reject-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const requesterId = parseInt(this.getAttribute('data-requester-id'));
                rejectFriendRequest(requesterId, this);
            });
        });
    }
    
    // Accepter une demande d'amitié
    async function acceptFriendRequest(requesterId, button) {
        try {
            const response = await fetch('/api/friends/request/accept', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    requester_id: requesterId
                })
            });
            
            if (response.ok) {
                // Supprimer la carte de demande
                const card = button.closest('.friend-request-card');
                if (card) card.remove();
                
                showNotification('Demande d\'amitié acceptée !', 'success');
                
                // Recharger les amis
                loadFriends();
                loadFriendshipStats();
            } else {
                const data = await response.json();
                showNotification(data.error || 'Erreur', 'error');
            }
        } catch (error) {
            console.error('Erreur acceptation:', error);
            showNotification('Erreur de connexion', 'error');
        }
    }
    
    // Rejeter une demande d'amitié
    async function rejectFriendRequest(requesterId, button) {
        try {
            const response = await fetch('/api/friends/request/reject', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include',
                body: JSON.stringify({
                    requester_id: requesterId
                })
            });
            
            if (response.ok) {
                // Supprimer la carte de demande
                const card = button.closest('.friend-request-card');
                if (card) card.remove();
                
                showNotification('Demande d\'amitié refusée', 'info');
                loadFriendshipStats();
            } else {
                const data = await response.json();
                showNotification(data.error || 'Erreur', 'error');
            }
        } catch (error) {
            console.error('Erreur rejet:', error);
            showNotification('Erreur de connexion', 'error');
        }
    }
    
    // Charger les statistiques d'amitié
    async function loadFriendshipStats() {
        try {
            const response = await fetch('/api/friends/stats', {
                method: 'GET',
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                updateFriendshipStats(data.stats);
            }
        } catch (error) {
            console.error('Erreur stats:', error);
        }
    }
    
    // Mettre à jour l'affichage des statistiques
    function updateFriendshipStats(stats) {
        // Mettre à jour le badge de notifications si il y a des demandes en attente
        const notificationBadge = document.querySelector('.notification-badge');
        if (notificationBadge && stats.pending_requests_count > 0) {
            notificationBadge.textContent = stats.pending_requests_count;
            notificationBadge.style.display = 'block';
        }
        
        // Mettre à jour le titre de la page avec le nombre d'amis
        const pageTitle = document.querySelector('.friends-header h1');
        if (pageTitle) {
            pageTitle.textContent = `Mes Amis (${stats.friends_count})`;
        }
    }
    
    // Afficher le menu d'options d'un ami
    function showFriendMenu(menuElement, friendId) {
        // Supprimer les menus existants
        document.querySelectorAll('.friend-context-menu').forEach(menu => menu.remove());
        
        const menu = document.createElement('div');
        menu.className = 'friend-context-menu';
        menu.innerHTML = `
            <div class="menu-item" data-action="message" data-friend-id="${friendId}">💬 Envoyer un message</div>
            <div class="menu-item" data-action="profile" data-friend-id="${friendId}">👤 Voir le profil</div>
            <div class="menu-item" data-action="mutual" data-friend-id="${friendId}">👥 Amis mutuels</div>
            <div class="menu-separator"></div>
            <div class="menu-item danger" data-action="remove" data-friend-id="${friendId}">🗑️ Supprimer ami</div>
            <div class="menu-item danger" data-action="block" data-friend-id="${friendId}">🚫 Bloquer</div>
        `;
        
        // Positionner le menu
        const rect = menuElement.getBoundingClientRect();
        menu.style.position = 'fixed';
        menu.style.top = rect.bottom + 'px';
        menu.style.left = (rect.left - 150) + 'px';
        menu.style.zIndex = '1000';
        
        document.body.appendChild(menu);
        
        // Ajouter les event listeners
        menu.querySelectorAll('.menu-item').forEach(item => {
            item.addEventListener('click', function() {
                const action = this.getAttribute('data-action');
                const friendId = this.getAttribute('data-friend-id');
                handleFriendMenuAction(action, friendId);
                menu.remove();
            });
        });
        
        // Fermer le menu en cliquant ailleurs
        setTimeout(() => {
            document.addEventListener('click', function closeMenu(e) {
                if (!menu.contains(e.target)) {
                    menu.remove();
                    document.removeEventListener('click', closeMenu);
                }
            });
        }, 100);
    }
    
    // Gérer les actions du menu d'ami
    function handleFriendMenuAction(action, friendId) {
        switch (action) {
            case 'message':
                // Ouvrir modal de message
                const friendCard = document.querySelector(`[data-user-id="${friendId}"]`);
                const friendName = friendCard ? friendCard.querySelector('h3').textContent : 'Ami';
                openMessageModal(friendName);
                break;
                
            case 'profile':
                window.location.href = `/profile?user=${friendId}`;
                break;
                
            case 'mutual':
                showMutualFriends(friendId);
                break;
                
            case 'remove':
                if (confirm('Êtes-vous sûr de vouloir supprimer cet ami ?')) {
                    removeFriend(friendId);
                }
                break;
                
            case 'block':
                if (confirm('Êtes-vous sûr de vouloir bloquer cet utilisateur ?')) {
                    blockUser(friendId);
                }
                break;
        }
    }
    
    // Supprimer un ami
    async function removeFriend(friendId) {
        try {
            const response = await fetch(`/api/friends/${friendId}`, {
                method: 'DELETE',
                credentials: 'include'
            });
            
            if (response.ok) {
                showNotification('Ami supprimé', 'info');
                loadFriends();
                loadFriendshipStats();
            } else {
                const data = await response.json();
                showNotification(data.error || 'Erreur', 'error');
            }
        } catch (error) {
            console.error('Erreur suppression ami:', error);
            showNotification('Erreur de connexion', 'error');
        }
    }
    
    // Bloquer un utilisateur
    async function blockUser(userId) {
        try {
            const response = await fetch(`/api/users/${userId}/block`, {
                method: 'POST',
                credentials: 'include'
            });
            
            if (response.ok) {
                showNotification('Utilisateur bloqué', 'info');
                loadFriends();
                loadFriendshipStats();
            } else {
                const data = await response.json();
                showNotification(data.error || 'Erreur', 'error');
            }
        } catch (error) {
            console.error('Erreur blocage:', error);
            showNotification('Erreur de connexion', 'error');
        }
    }
    
    // Afficher les amis mutuels
    async function showMutualFriends(userId) {
        try {
            const response = await fetch(`/api/friends/mutual/${userId}`, {
                method: 'GET',
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                displayMutualFriendsModal(data.mutual_friends || []);
            }
        } catch (error) {
            console.error('Erreur amis mutuels:', error);
        }
    }
    
    // Afficher le modal des amis mutuels
    function displayMutualFriendsModal(mutualFriends) {
        const modal = document.createElement('div');
        modal.className = 'modal mutual-friends-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Amis mutuels (${mutualFriends.length})</h3>
                    <button class="close-modal">&times;</button>
                </div>
                <div class="modal-body">
                    ${mutualFriends.length === 0 ? 
                        '<p>Aucun ami mutuel</p>' :
                        mutualFriends.map(friend => `
                            <div class="mutual-friend-item">
                                <div class="friend-avatar">
                                    ${friend.avatar ? 
                                        `<img src="${friend.avatar}" alt="${friend.username}">` :
                                        `<div class="user-pic">${friend.username.substring(0, 2).toUpperCase()}</div>`
                                    }
                                </div>
                                <span>${friend.username}</span>
                            </div>
                        `).join('')
                    }
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        modal.classList.add('active');
        
        // Fermer le modal
        modal.querySelector('.close-modal').addEventListener('click', () => {
            modal.remove();
        });
        
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
    }
    
    // Afficher le modal d'ajout d'ami
    function showAddFriendModal() {
        const modal = document.createElement('div');
        modal.className = 'modal add-friend-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Ajouter un ami</h3>
                    <button class="close-modal">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="search-section">
                        <input type="text" class="modal-search-input" placeholder="Rechercher par nom d'utilisateur...">
                        <div class="search-results"></div>
                    </div>
                    <div class="suggestions-section">
                        <h4>Suggestions</h4>
                        <div class="suggestions-list">Chargement...</div>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        modal.classList.add('active');
        
        const modalSearchInput = modal.querySelector('.modal-search-input');
        const searchResults = modal.querySelector('.search-results');
        const suggestionsList = modal.querySelector('.suggestions-list');
        
        // Recherche dans le modal
        let modalSearchTimeout;
        modalSearchInput.addEventListener('input', function() {
            clearTimeout(modalSearchTimeout);
            const query = this.value.trim();
            
            modalSearchTimeout = setTimeout(async () => {
                if (query.length >= 2) {
                    try {
                        const response = await fetch(`/api/users/search?q=${encodeURIComponent(query)}`, {
                            credentials: 'include'
                        });
                        
                        if (response.ok) {
                            const data = await response.json();
                            displayModalSearchResults(data.users || [], searchResults);
                        }
                    } catch (error) {
                        console.error('Erreur recherche modal:', error);
                    }
                } else {
                    searchResults.innerHTML = '';
                }
            }, 300);
        });
        
        // Charger les suggestions
        loadSuggestions(suggestionsList);
        
        // Fermer le modal
        modal.querySelector('.close-modal').addEventListener('click', () => {
            modal.remove();
        });
        
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
    }
    
    // Afficher les résultats de recherche dans le modal
    function displayModalSearchResults(users, container) {
        if (users.length === 0) {
            container.innerHTML = '<p>Aucun utilisateur trouvé</p>';
            return;
        }
        
        container.innerHTML = users.map(user => `
            <div class="modal-user-item" data-user-id="${user.id}">
                <div class="user-avatar">
                    ${user.avatar ? 
                        `<img src="${user.avatar}" alt="${user.username}">` :
                        `<div class="user-pic">${user.username.substring(0, 2).toUpperCase()}</div>`
                    }
                </div>
                <div class="user-info">
                    <span class="username">${user.username}</span>
                    ${user.mutual_friends > 0 ? `<small>${user.mutual_friends} amis mutuels</small>` : ''}
                </div>
                <div class="user-actions">
                    ${getUserActionButton(user)}
                </div>
            </div>
        `).join('');
        
        // Ajouter les event listeners
        container.querySelectorAll('.add-friend-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                const userId = parseInt(this.getAttribute('data-user-id'));
                sendFriendRequest(userId, this);
            });
        });
    }
    
    // Obtenir le bouton d'action pour un utilisateur
    function getUserActionButton(user) {
        if (!user.friendship_status) {
            return `<button class="add-friend-btn" data-user-id="${user.id}">Ajouter</button>`;
        } else if (user.friendship_status === 'pending') {
            return `<button class="pending-btn" disabled>En attente</button>`;
        } else if (user.friendship_status === 'accepted') {
            return `<button class="friends-btn" disabled>Amis</button>`;
        } else {
            return `<button class="blocked-btn" disabled>Bloqué</button>`;
        }
    }
    
    // Charger les suggestions d'amis
    async function loadSuggestions(container) {
        try {
            const response = await fetch('/api/friends/suggestions', {
                credentials: 'include'
            });
            
            if (response.ok) {
                const data = await response.json();
                displayModalSearchResults(data.suggestions || [], container);
            } else {
                container.innerHTML = '<p>Aucune suggestion disponible</p>';
            }
        } catch (error) {
            console.error('Erreur suggestions:', error);
            container.innerHTML = '<p>Erreur de chargement</p>';
        }
    }
    
    // Fonctions utilitaires
    function formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffTime = Math.abs(now - date);
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        
        if (diffDays === 1) return 'Hier';
        if (diffDays < 7) return `${diffDays} jours`;
        if (diffDays < 30) return `${Math.ceil(diffDays / 7)} semaines`;
        if (diffDays < 365) return `${Math.ceil(diffDays / 30)} mois`;
        return `${Math.ceil(diffDays / 365)} ans`;
    }
    
    function showNotification(message, type = 'info') {
        // Créer la notification
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <span>${message}</span>
            <button class="close-notification">&times;</button>
        `;
        
        // Ajouter au DOM
        document.body.appendChild(notification);
        
        // Afficher avec animation
        setTimeout(() => notification.classList.add('show'), 100);
        
        // Fermer automatiquement après 5 secondes
        setTimeout(() => {
            notification.classList.remove('show');
            setTimeout(() => notification.remove(), 300);
        }, 5000);
        
        // Bouton de fermeture
        notification.querySelector('.close-notification').addEventListener('click', () => {
            notification.classList.remove('show');
            setTimeout(() => notification.remove(), 300);
        });
    }
    
    // Fonction pour ouvrir le modal de message (existante)
    function openMessageModal(friendName) {
        let messageModal = document.getElementById('messageModal');
        
        if (!messageModal) {
            createMessageModal();
            messageModal = document.getElementById('messageModal');
        }
        
        const modalUserName = document.getElementById('modalUserName');
        if (modalUserName) {
            modalUserName.textContent = friendName;
        }
        
        messageModal.classList.add('active');
        document.body.style.overflow = 'hidden';
        
        setTimeout(() => {
            const messageInput = document.querySelector('.message-input');
            if (messageInput) {
                messageInput.focus();
            }
        }, 300);
        
        showNotification(`💬 Conversation avec ${friendName} ouverte`, 'info');
    }
    
    // Créer le modal de message si il n'existe pas
    function createMessageModal() {
        const modal = document.createElement('div');
        modal.id = 'messageModal';
        modal.className = 'modal message-modal';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Message à <span id="modalUserName"></span></h3>
                    <button class="close-modal">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="message-history"></div>
                    <div class="message-input-container">
                        <input type="text" class="message-input" placeholder="Tapez votre message...">
                        <button class="send-btn">Envoyer</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Event listeners pour le modal
        modal.querySelector('.close-modal').addEventListener('click', () => {
            modal.classList.remove('active');
            document.body.style.overflow = 'auto';
        });
        
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
                document.body.style.overflow = 'auto';
            }
        });
        
        // Gestion de l'envoi de messages
        const messageInput = modal.querySelector('.message-input');
        const sendBtn = modal.querySelector('.send-btn');
        
        function sendMessage() {
            const message = messageInput.value.trim();
            if (message) {
                // Ici vous pouvez ajouter la logique d'envoi de message
                console.log('Message envoyé:', message);
                messageInput.value = '';
                showNotification('Message envoyé !', 'success');
            }
        }
        
        sendBtn.addEventListener('click', sendMessage);
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    }
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
