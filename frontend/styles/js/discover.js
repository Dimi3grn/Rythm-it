// JavaScript pour la page Découverte - discover.js

// JavaScript pour la page Découverte - discover.js (VERSION MISE À JOUR)

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let currentFilter = 'all';
    let currentSearchType = 'all'; // 'all', 'discussions', 'users'
    let currentCategory = 'all';
    let currentGenre = '';
    let searchTimeout;
    let isVoiceSearchActive = false;
    let selectedTags = [];
    
    // Éléments DOM
    const globalSearch = document.querySelector('.global-search');
    const searchBtn = document.querySelector('.search-btn');
    const voiceSearchBtn = document.querySelector('.voice-search-btn');
    const filterBtns = document.querySelectorAll('.filter-btn');
    const categoryFilters = document.querySelectorAll('.filter-category');
    const genreFilters = document.querySelectorAll('.filter-genre');
    const refreshBtn = document.querySelector('.refresh-recommendations');
    const tagSelector = document.getElementById('tag-selector');
    const selectedTagsContainer = document.getElementById('selected-tags');
    const searchResults = document.getElementById('search-results');
    const discoverSections = document.getElementById('discover-sections');
    
    // Initialisation
    init();
    
    function init() {
        // Attacher les événements
        attachEventListeners();
        
        // Animations d'entrée
        animateOnLoad();
        
        // Démarrer les mises à jour
        startDynamicUpdates();
        
        // Charger le contenu initial
        loadDiscoverContent();
        
        // Initialiser la navigation améliorée
        if (typeof initializeEnhancedNavigation === 'function') {
            initializeEnhancedNavigation();
        }
        
        // Gérer les paramètres URL pour les genres
        handleURLParams();
        
        // Charger les tags depuis l'API
        loadTagsFromAPI();
    }
    
    // NOUVELLE FONCTION: Charger les tags depuis l'API
    function loadTagsFromAPI() {
        fetch('/api/public/tags')
            .then(response => response.json())
            .then(data => {
                console.log('🏷️ Réponse API tags:', data);
                
                // L'API retourne {success: true, data: [...]}
                const tags = data.success && data.data ? data.data : [];
                
                const tagButtons = document.getElementById('tag-buttons');
                if (tagButtons && tags.length > 0) {
                    tagButtons.innerHTML = '';
                    
                    // Créer une grille de tags
                    const tagGrid = document.createElement('div');
                    tagGrid.className = 'tag-grid';
                    
                    tags.forEach(tag => {
                        const button = document.createElement('button');
                        button.type = 'button';
                        button.className = 'tag-option';
                        button.textContent = tag.name || tag.Name || tag;
                        button.onclick = () => addTag(tag.name || tag.Name || tag);
                        tagGrid.appendChild(button);
                    });
                    
                    tagButtons.appendChild(tagGrid);
                    console.log('🏷️ Tags chargés depuis l\'API:', tags.length);
                    showNotification(`🏷️ ${tags.length} tags chargés`, 'success');
                } else {
                    console.warn('⚠️ Aucun tag trouvé dans la réponse');
                    showNotification('⚠️ Aucun tag disponible', 'warning');
                }
            })
            .catch(error => {
                console.error('❌ Erreur lors du chargement des tags:', error);
                showNotification('❌ Erreur lors du chargement des tags', 'error');
            });
    }
    
    // NOUVELLE FONCTION: Toggle du sélecteur de tags
    window.toggleTagSelector = function() {
        if (!tagSelector) {
            console.error('❌ Élément tag-selector non trouvé');
            return;
        }
        
        const isVisible = tagSelector.style.display !== 'none';
        
        if (isVisible) {
            tagSelector.style.display = 'none';
            showNotification('🏷️ Sélecteur de tags fermé', 'info');
        } else {
            tagSelector.style.display = 'block';
            showNotification('🏷️ Sélecteur de tags ouvert', 'info');
        }
    };
    
    // NOUVELLE FONCTION: Ajouter un tag
    window.addTag = function(tagName) {
        if (!tagName || selectedTags.includes(tagName)) {
            return;
        }
        
        selectedTags.push(tagName);
        updateSelectedTagsDisplay();
        applyTagFilters();
        showNotification(`🏷️ Tag "${tagName}" ajouté`, 'success');
    };
    
    // NOUVELLE FONCTION: Supprimer un tag
    window.removeTag = function(tagName) {
        const index = selectedTags.indexOf(tagName);
        if (index > -1) {
            selectedTags.splice(index, 1);
            updateSelectedTagsDisplay();
            applyTagFilters();
            showNotification(`🏷️ Tag "${tagName}" supprimé`, 'info');
        }
    };
    
    // NOUVELLE FONCTION: Mettre à jour l'affichage des tags sélectionnés
    function updateSelectedTagsDisplay() {
        if (!selectedTagsContainer) return;
        
        if (selectedTags.length === 0) {
            selectedTagsContainer.innerHTML = '<p class="no-tags">Aucun tag sélectionné</p>';
        } else {
            selectedTagsContainer.innerHTML = selectedTags.map(tag => 
                `<span class="selected-tag">
                    ${tag}
                    <button onclick="removeTag('${tag}')" class="remove-tag-btn">×</button>
                </span>`
            ).join('');
        }
    }
    
    // NOUVELLE FONCTION: Appliquer les filtres par tags
    function applyTagFilters() {
        if (selectedTags.length === 0) {
            // Réinitialiser l'affichage si aucun tag sélectionné
            resetSearch();
            return;
        }
        
        // Rechercher automatiquement les threads avec les tags sélectionnés
        searchThreadsByTags(selectedTags);
    }
    
    // NOUVELLE FONCTION: Rechercher les threads par tags
    async function searchThreadsByTags(tags) {
        try {
            showNotification(`🔍 Recherche de threads avec TOUS les tags: ${tags.join(', ')}`, 'info');
            
            // Afficher la section des résultats et masquer les sections de découverte
            if (searchResults) {
                searchResults.style.display = 'block';
                const resultsTitle = document.getElementById('results-title');
                if (resultsTitle) {
                    resultsTitle.textContent = `Résultats pour les tags: ${tags.join(' + ')}`;
                }
            }
            
            if (discoverSections) {
                discoverSections.style.display = 'none';
            }
            
            // Afficher le loading
            const searchLoading = document.getElementById('search-loading');
            if (searchLoading) {
                searchLoading.style.display = 'block';
            }
            
            // Construire l'URL avec les tags séparés par des virgules
            const tagsParam = tags.join(',');
            const url = `/api/public/threads/search?tags=${encodeURIComponent(tagsParam)}`;
            
            console.log('🔍 Recherche par tags URL:', url);
            
            const response = await fetch(url);
            const data = await response.json();
            
            console.log('🔍 Réponse recherche par tags:', data);
            
            // Masquer le loading
            if (searchLoading) {
                searchLoading.style.display = 'none';
            }
            
            if (data.success && data.data && data.data.threads) {
                const threads = data.data.threads;
                displayThreadSearchResults(threads, `Tags: ${tags.join(' + ')}`);
                showNotification(`🏷️ ${threads.length} thread(s) trouvé(s) avec TOUS ces tags`, 'success');
            } else {
                displayNoResults(`Tags: ${tags.join(' + ')}`);
                showNotification('🏷️ Aucun thread trouvé avec TOUS ces tags', 'warning');
            }
        } catch (error) {
            console.error('❌ Erreur lors de la recherche par tags:', error);
            
            // Masquer le loading en cas d'erreur
            const searchLoading = document.getElementById('search-loading');
            if (searchLoading) {
                searchLoading.style.display = 'none';
            }
            
            showNotification('❌ Erreur lors de la recherche par tags', 'error');
            displayNoResults(`Tags: ${tags.join(' + ')}`);
        }
    }
    
    // NOUVELLE FONCTION: Réinitialiser la recherche
    function resetSearch() {
        if (searchResults) {
            searchResults.style.display = 'none';
        }
        
        if (discoverSections) {
            discoverSections.style.display = 'block';
        }
        
        showNotification('🔄 Affichage réinitialisé', 'info');
    }
    
    // NOUVELLE FONCTION: Effacer la recherche
    window.clearSearch = function() {
        if (globalSearch) {
            globalSearch.value = '';
        }
        
        selectedTags = [];
        updateSelectedTagsDisplay();
        
        resetSearch();
        showNotification('🔍 Recherche effacée', 'info');
    };
    
    // FONCTION MODIFIÉE: Recherche améliorée
    window.performSearch = function() {
        const query = globalSearch ? globalSearch.value.trim() : '';
        
        if (query) {
            performSearchWithQuery(query);
        } else if (selectedTags.length > 0) {
            // Recherche par tags seulement
            searchThreadsByTags(selectedTags);
        } else {
            showNotification('🔍 Veuillez saisir un terme de recherche ou sélectionner des tags', 'warning');
        }
    };
    
    // NOUVELLE FONCTION: Gérer les paramètres URL
    function handleURLParams() {
        const urlParams = new URLSearchParams(window.location.search);
        const genreParam = urlParams.get('genre');
        
        if (genreParam) {
            setTimeout(() => {
                setActiveGenre(genreParam);
                showNotification(`🎵 Filtre de genre "${genreParam}" appliqué`, 'info');
            }, 1000);
        }
        
        // Gérer les ancres
        if (window.location.hash) {
            setTimeout(() => {
                handlePageAnchor(window.location.hash);
            }, 1500);
        }
    }
    
    // NOUVELLE FONCTION: Gérer les ancres de page
    function handlePageAnchor(anchor) {
        switch(anchor) {
            case '#trending':
                scrollToTrending();
                break;
            case '#live-sessions':
                scrollToLiveSessions();
                break;
            default:
                const element = document.querySelector(anchor);
                if (element) {
                    element.scrollIntoView({ behavior: 'smooth' });
                }
        }
    }
    
    // NOUVELLE FONCTION: Scroll vers les tendances
    function scrollToTrending() {
        const trendingSection = document.getElementById('trending') || 
                              document.querySelector('.trending-carousel')?.closest('.discover-section');
        
        if (trendingSection) {
            trendingSection.scrollIntoView({ behavior: 'smooth' });
            
            // Mettre en évidence temporairement
            trendingSection.style.border = '2px solid rgba(255, 107, 107, 0.5)';
            setTimeout(() => {
                trendingSection.style.border = '';
            }, 3000);
            
            showNotification('🔥 Section Tendances mise en évidence', 'info');
        }
    }
    
    // NOUVELLE FONCTION: Scroll vers les sessions live
    function scrollToLiveSessions() {
        const liveSection = document.querySelector('.live-sessions') || 
                           document.querySelector('[id*="live"]');
        
        if (liveSection) {
            liveSection.scrollIntoView({ behavior: 'smooth' });
            showNotification('🎵 Sessions Live affichées', 'info');
        } else {
            showNotification('🎵 Sessions Live bientôt disponibles !', 'info');
        }
    }
    
    // FONCTION MODIFIÉE: Rendre setActiveGenre globale
    window.setActiveGenre = function(genre) {
        // Mettre à jour la navigation
        document.querySelectorAll('.nav-item[data-genre]').forEach(item => {
            item.classList.remove('active');
            if (item.getAttribute('data-genre') === genre) {
                item.classList.add('active');
            }
        });
        
        currentGenre = genre;
        applyFilters();
        
        // Mettre à jour l'URL sans recharger la page
        const url = new URL(window.location);
        url.searchParams.set('genre', genre);
        window.history.replaceState({}, '', url);
    };
    
    // Gestion des événements (code existant avec modifications)
    function attachEventListeners() {
        // Recherche globale
        if (globalSearch) {
            globalSearch.addEventListener('input', handleSearch);
            globalSearch.addEventListener('keypress', function(e) {
                if (e.key === 'Enter') {
                    performSearch();
                }
            });
        }
        
        if (searchBtn) {
            searchBtn.addEventListener('click', performSearch);
        }
        
        if (voiceSearchBtn) {
            voiceSearchBtn.addEventListener('click', toggleVoiceSearch);
        }
        
        // Filtres principaux
        filterBtns.forEach(btn => {
            btn.addEventListener('click', function() {
                const filter = this.getAttribute('data-filter');
                
                // Mettre à jour les boutons actifs
                filterBtns.forEach(b => b.classList.remove('active'));
                this.classList.add('active');
                
                // Mettre à jour le filtre actuel
                currentSearchType = filter;
                currentFilter = filter;
                
                console.log('🔍 Filtre sélectionné:', filter);
                
                // Si on a une recherche active, relancer avec le nouveau filtre
                const query = globalSearch ? globalSearch.value.trim() : '';
                if (query) {
                    performSearchWithQuery(query);
                } else {
                    // Pas de recherche active, afficher le contenu par défaut
                    if (filter === 'all') {
                        if (searchResults) searchResults.style.display = 'none';
                        if (discoverSections) discoverSections.style.display = 'block';
                    }
                }
            });
        });
        
        // MODIFIÉ: Filtres de genres avec gestion URL
        document.querySelectorAll('.nav-item[data-genre]').forEach(item => {
            item.addEventListener('click', function(e) {
                e.preventDefault();
                const genre = this.getAttribute('data-genre');
                setActiveGenre(genre);
                showNotification(`🎵 Filtre ${genre} appliqué`, 'info');
            });
        });
        
        // Filtres de catégories
        categoryFilters.forEach(filter => {
            filter.addEventListener('click', function(e) {
                e.preventDefault();
                const category = this.getAttribute('data-category');
                setActiveCategory(category);
            });
        });
        
        // Boutons de lecture
        document.querySelectorAll('.play-btn-large, .play-btn-medium, .play-rec-btn').forEach(btn => {
            btn.addEventListener('click', handlePlay);
        });
        
        // Actions des artistes
        document.querySelectorAll('.follow-btn').forEach(btn => {
            btn.addEventListener('click', handleFollow);
        });
        
        document.querySelectorAll('.play-artist-btn').forEach(btn => {
            btn.addEventListener('click', handlePlayArtist);
        });
        
        // Actions des playlists
        document.querySelectorAll('.like-playlist-btn').forEach(btn => {
            btn.addEventListener('click', handleLikePlaylist);
        });
        
        document.querySelectorAll('.save-playlist-btn').forEach(btn => {
            btn.addEventListener('click', handleSavePlaylist);
        });
        
        document.querySelectorAll('.share-playlist-btn').forEach(btn => {
            btn.addEventListener('click', handleSharePlaylist);
        });
        
        // Boutons d'exploration
        document.querySelectorAll('.explore-rec-btn').forEach(btn => {
            btn.addEventListener('click', handleExploreRecommendation);
        });
        
        // Actualisation des recommandations
        if (refreshBtn) {
            refreshBtn.addEventListener('click', refreshRecommendations);
        }
        
        // Actions de la sidebar
        const discoverGenreBtn = document.querySelector('.discover-genre-btn');
        if (discoverGenreBtn) {
            discoverGenreBtn.addEventListener('click', exploreGenre);
        }
        
        // Sessions live
        document.querySelectorAll('.join-live-btn').forEach(btn => {
            btn.addEventListener('click', joinLiveSession);
        });
        
        document.querySelectorAll('.notify-live-btn').forEach(btn => {
            btn.addEventListener('click', notifyLiveSession);
        });
        
        // Navigation hover effects
        document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card').forEach(card => {
            card.addEventListener('mouseenter', handleCardHover);
            card.addEventListener('mouseleave', handleCardLeave);
        });
    }
    
    // [RESTE DU CODE EXISTANT - toutes les autres fonctions restent identiques]
    // Gestion de la recherche
    function handleSearch() {
        const query = globalSearch.value;
        
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            if (query.length > 2) {
                performSearchWithQuery(query);
            } else if (query.length === 0) {
                resetSearch();
            }
        }, 300);
    }
    
    function performSearch() {
        const query = globalSearch ? globalSearch.value.trim() : '';
        
        if (query) {
            performSearchWithQuery(query);
        } else if (selectedTags.length > 0) {
            // Recherche par tags seulement
            searchThreadsByTags(selectedTags);
        } else {
            showNotification('🔍 Veuillez saisir un terme de recherche ou sélectionner des tags', 'warning');
        }
    }
    
    function performSearchWithQuery(query) {
        showNotification(`🔍 Recherche: "${query}"`, 'info');
        
        // Afficher la section des résultats et masquer les sections de découverte
        if (searchResults) {
            searchResults.style.display = 'block';
            const resultsTitle = document.getElementById('results-title');
            if (resultsTitle) {
                resultsTitle.textContent = `Résultats pour "${query}"`;
            }
        }
        
        if (discoverSections) {
            discoverSections.style.display = 'none';
        }
        
        // Afficher le loading
        const searchLoading = document.getElementById('search-loading');
        if (searchLoading) {
            searchLoading.style.display = 'block';
        }
        
        // Rechercher selon le filtre actif
        if (currentSearchType === 'users') {
            // Recherche d'utilisateurs
            searchUsersInDatabase(query)
                .then(results => {
                    if (searchLoading) {
                        searchLoading.style.display = 'none';
                    }
                    displayUserSearchResults(results, query);
                    showNotification(`✅ ${results.length} utilisateur(s) trouvé(s)`, 'success');
                })
                .catch(error => {
                    console.error('❌ Erreur lors de la recherche d\'utilisateurs:', error);
                    if (searchLoading) {
                        searchLoading.style.display = 'none';
                    }
                    showNotification('❌ Erreur lors de la recherche', 'error');
                    displayNoResults(query);
                });
        } else {
            // Recherche de discussions (threads)
            searchThreadsInDatabase(query)
                .then(results => {
                    if (searchLoading) {
                        searchLoading.style.display = 'none';
                    }
                    
                    // Construire le message de recherche
                    let searchMessage = `"${query}"`;
                    if (selectedTags.length > 0) {
                        searchMessage += ` + tags: ${selectedTags.join(' + ')}`;
                    }
                    
                    displayThreadSearchResults(results, searchMessage);
                    
                    if (selectedTags.length > 0) {
                        showNotification(`✅ ${results.length} thread(s) trouvé(s) avec le texte ET tous les tags`, 'success');
                    } else {
                        showNotification(`✅ ${results.length} thread(s) trouvé(s)`, 'success');
                    }
                })
                .catch(error => {
                    console.error('❌ Erreur lors de la recherche:', error);
                    if (searchLoading) {
                        searchLoading.style.display = 'none';
                    }
                    showNotification('❌ Erreur lors de la recherche', 'error');
                    displayNoResults(query);
                });
        }
    }
    
    // NOUVELLE FONCTION: Rechercher les threads dans la base de données
    async function searchThreadsInDatabase(query) {
        try {
            // Construire l'URL de recherche avec les paramètres
            const searchParams = new URLSearchParams();
            searchParams.append('q', query);
            
            // Ajouter les tags sélectionnés
            if (selectedTags.length > 0) {
                searchParams.append('tags', selectedTags.join(','));
            }
            
            const response = await fetch(`/api/public/threads/search?${searchParams.toString()}`);
            
            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            console.log('🔍 Résultats de recherche reçus:', data);
            
            // L'API retourne {success: true, data: {threads: [...], count: ...}}
            if (data.success && data.data && data.data.threads) {
                return data.data.threads;
            }
            
            return [];
        } catch (error) {
            console.error('❌ Erreur lors de la recherche de threads:', error);
            throw error;
        }
    }
    
    // NOUVELLE FONCTION: Afficher les résultats de recherche de threads
    function displayThreadSearchResults(threads, query) {
        const resultsList = document.getElementById('results-list');
        if (!resultsList) return;
        
        if (threads.length === 0) {
            displayNoResults(query);
            return;
        }
        
        let html = `
            <div class="results-section">
                <h3>🧵 Discussions trouvées (${threads.length})</h3>
                <div class="results-grid thread-results">
        `;
        
        threads.forEach(thread => {
            html += createThreadResultHTML(thread);
        });
        
        html += `
                </div>
            </div>
        `;
        
        resultsList.innerHTML = html;
        
        // Réattacher les événements aux nouveaux éléments
        reattachThreadEventListeners();
    }
    
    // NOUVELLE FONCTION: Créer le HTML pour un thread dans les résultats
    function createThreadResultHTML(thread) {
        const timeAgo = formatTimeAgo(thread.created_at || thread.CreatedAt);
        const likes = thread.likes || thread.Likes || 0;
        const comments = thread.comments || thread.Comments || 0;
        const author = thread.author || thread.Author || 'Utilisateur';
        const title = thread.title || thread.Title || 'Sans titre';
        const content = thread.content || thread.Content || '';
        const threadId = thread.id || thread.ID;
        
        // Extraire les tags si disponibles
        let tagsHTML = '';
        if (thread.tags && Array.isArray(thread.tags)) {
            tagsHTML = thread.tags.map(tag => `<span class="tag">${tag}</span>`).join('');
        }
        
        return `
            <div class="thread-result-item" data-thread-id="${threadId}">
                <div class="thread-result-header">
                    <div class="user-pic">${author.substring(0, 2).toUpperCase()}</div>
                    <div class="thread-result-meta">
                        <h4>${author}</h4>
                        <span class="thread-result-time">${timeAgo}</span>
                    </div>
                </div>
                <div class="thread-result-content">
                    <h3 class="thread-result-title">${title}</h3>
                    <p class="thread-result-text">${content.substring(0, 150)}${content.length > 150 ? '...' : ''}</p>
                    ${tagsHTML ? `<div class="thread-result-tags">${tagsHTML}</div>` : ''}
                </div>
                <div class="thread-result-stats">
                    <span class="stat">❤️ ${likes}</span>
                    <span class="stat">💬 ${comments}</span>
                    <a href="/thread/${threadId}" class="view-thread-btn">Voir le thread →</a>
                </div>
            </div>
        `;
    }
    
    // NOUVELLE FONCTION: Afficher "aucun résultat"
    function displayNoResults(query) {
        const resultsList = document.getElementById('results-list');
        if (!resultsList) return;
        
        const isUserSearch = currentSearchType === 'users';
        const itemType = isUserSearch ? 'utilisateur' : 'thread';
        
        resultsList.innerHTML = `
            <div class="no-results">
                <h3>Aucun ${itemType} trouvé</h3>
                <p>Aucun ${itemType} ne correspond à votre recherche "${query}".</p>
                <div class="search-suggestions">
                    <h4>Suggestions :</h4>
                    <ul>
                        <li>Vérifiez l'orthographe de vos mots-clés</li>
                        <li>Essayez des termes plus généraux</li>
                        ${!isUserSearch ? '<li>Utilisez les filtres par tags</li>' : ''}
                        ${!isUserSearch ? '<li>Recherchez par nom d\'artiste ou genre musical</li>' : ''}
                    </ul>
                </div>
            </div>
        `;
    }
    
    // NOUVELLE FONCTION: Rechercher les utilisateurs dans la base de données
    async function searchUsersInDatabase(query) {
        try {
            const response = await fetch(`/api/users/search?q=${encodeURIComponent(query)}&limit=20`, {
                credentials: 'include'
            });
            
            if (!response.ok) {
                throw new Error(`Erreur HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            console.log('👥 Résultats de recherche utilisateurs:', data);
            
            // L'API retourne {success: true, data: {users: [...]}}
            if (data.success && data.data && data.data.users) {
                return data.data.users;
            }
            
            return [];
        } catch (error) {
            console.error('❌ Erreur lors de la recherche d\'utilisateurs:', error);
            throw error;
        }
    }
    
    // NOUVELLE FONCTION: Afficher les résultats de recherche d'utilisateurs
    function displayUserSearchResults(users, query) {
        const resultsList = document.getElementById('results-list');
        if (!resultsList) return;
        
        if (users.length === 0) {
            displayNoResults(query);
            return;
        }
        
        let html = `
            <div class="results-section">
                <h3>👥 Utilisateurs trouvés (${users.length})</h3>
                <div class="results-grid user-results">
        `;
        
        users.forEach(user => {
            html += createUserResultHTML(user);
        });
        
        html += `
                </div>
            </div>
        `;
        
        resultsList.innerHTML = html;
        
        // Réattacher les événements aux nouveaux éléments
        reattachUserEventListeners();
    }
    
    // NOUVELLE FONCTION: Créer le HTML pour un utilisateur dans les résultats
    function createUserResultHTML(user) {
        const avatar = user.avatar || user.Avatar;
        const username = user.username || user.Username;
        const userId = user.id || user.ID;
        const friendshipStatus = user.friendship_status || user.FriendshipStatus;
        const mutualFriends = user.mutual_friends || user.MutualFriends || 0;
        
        // Déterminer le bouton d'action selon le statut d'amitié
        let actionButton = '';
        if (friendshipStatus === 'accepted') {
            actionButton = '<button class="user-action-btn friend" disabled>✓ Ami</button>';
        } else if (friendshipStatus === 'pending') {
            actionButton = '<button class="user-action-btn pending" disabled>⏳ Demande envoyée</button>';
        } else {
            actionButton = `<button class="user-action-btn add-friend" data-user-id="${userId}">+ Ajouter</button>`;
        }
        
        const avatarHTML = avatar 
            ? `<img src="${avatar}" alt="${username}" class="user-result-avatar">` 
            : `<div class="user-result-avatar-placeholder">${username.substring(0, 2).toUpperCase()}</div>`;
        
        return `
            <div class="user-result-item" data-user-id="${userId}">
                <div class="user-result-header">
                    ${avatarHTML}
                    <div class="user-result-info">
                        <h4 class="user-result-name">${username}</h4>
                        ${mutualFriends > 0 ? `<p class="user-result-mutual">👥 ${mutualFriends} ami(s) en commun</p>` : ''}
                    </div>
                </div>
                <div class="user-result-actions">
                    ${actionButton}
                    <a href="/profile?user=${userId}" class="user-action-btn view-profile">Voir profil</a>
                </div>
            </div>
        `;
    }
    
    // NOUVELLE FONCTION: Réattacher les événements pour les utilisateurs
    function reattachUserEventListeners() {
        // Gérer les clics sur les boutons "Ajouter ami"
        document.querySelectorAll('.add-friend').forEach(btn => {
            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                const userId = parseInt(this.getAttribute('data-user-id'));
                sendFriendRequest(userId, this);
            });
        });
    }
    
    // NOUVELLE FONCTION: Envoyer une demande d'ami
    async function sendFriendRequest(userId, buttonElement) {
        try {
            buttonElement.disabled = true;
            buttonElement.textContent = '⏳ Envoi...';
            
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
            
            const data = await response.json();
            
            if (data.success) {
                buttonElement.textContent = '✓ Demande envoyée';
                buttonElement.className = 'user-action-btn pending';
                showNotification('✅ Demande d\'ami envoyée avec succès !', 'success');
            } else {
                throw new Error(data.message || 'Erreur lors de l\'envoi de la demande');
            }
        } catch (error) {
            console.error('❌ Erreur envoi demande d\'ami:', error);
            buttonElement.disabled = false;
            buttonElement.textContent = '+ Ajouter';
            showNotification('❌ ' + error.message, 'error');
        }
    }
    
    // NOUVELLE FONCTION: Formater le temps relatif
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
    
    // NOUVELLE FONCTION: Réattacher les événements pour les threads
    function reattachThreadEventListeners() {
        // Gérer les clics sur les threads pour navigation
        document.querySelectorAll('.thread-result-item').forEach(item => {
            item.addEventListener('click', function(e) {
                // Ne pas naviguer si on clique sur le bouton "Voir le thread"
                if (e.target.classList.contains('view-thread-btn')) {
                    return;
                }
                
                const threadId = this.getAttribute('data-thread-id');
                if (threadId) {
                    window.location.href = `/thread/${threadId}`;
                }
            });
        });
        
        // Gérer les boutons "Voir le thread"
        document.querySelectorAll('.view-thread-btn').forEach(btn => {
            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                // La navigation se fait via le href
            });
        });
    }
    
    // Recherche vocale
    function toggleVoiceSearch() {
        if (!isVoiceSearchActive) {
            startVoiceSearch();
        } else {
            stopVoiceSearch();
        }
    }
    
    function startVoiceSearch() {
        if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
            const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
            const recognition = new SpeechRecognition();
            
            recognition.lang = 'fr-FR';
            recognition.continuous = false;
            recognition.interimResults = false;
            
            recognition.onstart = function() {
                isVoiceSearchActive = true;
                voiceSearchBtn.style.color = '#ff6b6b';
                voiceSearchBtn.style.animation = 'pulse 1s infinite';
                globalSearch.placeholder = 'Parlez maintenant...';
                showNotification('🎤 Écoute en cours...', 'info');
            };
            
            recognition.onresult = function(event) {
                const transcript = event.results[0][0].transcript;
                globalSearch.value = transcript;
                performSearchWithQuery(transcript);
                showNotification(`🎤 Recherche vocale: "${transcript}"`, 'success');
            };
            
            recognition.onerror = function(event) {
                showNotification('❌ Erreur de reconnaissance vocale', 'warning');
            };
            
            recognition.onend = function() {
                stopVoiceSearch();
            };
            
            recognition.start();
        } else {
            showNotification('❌ Recherche vocale non supportée', 'warning');
        }
    }
    
    function stopVoiceSearch() {
        isVoiceSearchActive = false;
        voiceSearchBtn.style.color = '';
        voiceSearchBtn.style.animation = '';
        globalSearch.placeholder = 'Rechercher des artistes, albums, playlists...';
    }
    
    // Gestion des filtres
    function setActiveFilter(filter) {
        // Mettre à jour les boutons
        filterBtns.forEach(btn => btn.classList.remove('active'));
        const filterBtn = document.querySelector(`[data-filter="${filter}"]`);
        if (filterBtn) {
            filterBtn.classList.add('active');
        }
        
        currentFilter = filter;
        currentSearchType = filter;
        
        // Ne pas appeler applyFilters() pour les nouveaux filtres (discussions, users)
        // Ces filtres sont gérés par la recherche
        if (filter !== 'discussions' && filter !== 'users') {
            applyFilters();
        }
    }
    
    function setActiveCategory(category) {
        // Mettre à jour la navigation
        categoryFilters.forEach(filter => filter.classList.remove('active'));
        document.querySelector(`[data-category="${category}"]`).classList.add('active');
        
        currentCategory = category;
        applyFilters();
    }
    
    function applyFilters() {
        // Ne rien faire si on est en mode recherche (discussions ou users)
        if (currentSearchType === 'discussions' || currentSearchType === 'users') {
            return;
        }
        
        // Filtrer les éléments visibles
        const allItems = document.querySelectorAll('.artist-card, .album-card, .playlist-discover-card');
        
        allItems.forEach(item => {
            let shouldShow = true;
            
            // Filtre par catégorie
            if (currentCategory !== 'all') {
                const itemCategory = item.getAttribute('data-category');
                if (itemCategory && itemCategory !== currentCategory) {
                    shouldShow = false;
                }
            }
            
            // Filtre par genre
            if (currentGenre && currentGenre !== '') {
                const itemGenre = item.getAttribute('data-genre');
                if (itemGenre && itemGenre !== currentGenre) {
                    shouldShow = false;
                }
            }
            
            // Appliquer la visibilité
            if (shouldShow) {
                item.style.display = 'block';
                setTimeout(() => {
                    item.style.opacity = '1';
                    item.style.transform = 'translateY(0)';
                }, 50);
            } else {
                item.style.opacity = '0';
                item.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    item.style.display = 'none';
                }, 300);
            }
        });
        
        showNotification(`Filtres appliqués: ${currentCategory} ${currentGenre ? '• ' + currentGenre : ''}`, 'info');
    }
    
    // [TOUTES LES AUTRES FONCTIONS RESTENT IDENTIQUES]
    // Actions de lecture, artistes, playlists, etc... (code existant)
    
    // Actions de lecture
    function handlePlay(e) {
        const btn = e.currentTarget;
        const card = btn.closest('.trending-item, .album-card, .rec-track');
        
        // Animation de lecture
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        // Changer l'état
        const isPlaying = btn.textContent.includes('⏸️');
        btn.textContent = isPlaying ? '▶️' : '⏸️';
        
        // Notification
        if (card) {
            const title = card.querySelector('h3, h4, h5')?.textContent || 'Piste';
            if (!isPlaying) {
                showNotification(`🎵 Lecture: ${title}`, 'music');
                animatePlayingCard(card);
            } else {
                showNotification('⏸️ Lecture en pause', 'info');
            }
        }
    }
    
    function animatePlayingCard(card) {
        card.style.border = '2px solid rgba(102, 126, 234, 0.5)';
        card.style.boxShadow = '0 0 30px rgba(102, 126, 234, 0.3)';
        
        // Ajouter une animation de pulsation
        card.style.animation = 'gentle-pulse 2s ease-in-out infinite';
    }
    
    // Actions des artistes
    function handleFollow(e) {
        const btn = e.currentTarget;
        const artistCard = btn.closest('.artist-card');
        const artistName = artistCard.querySelector('h4').textContent;
        
        // Animation
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        // Changer l'état
        const isFollowing = btn.textContent.includes('Suivi');
        if (isFollowing) {
            btn.textContent = '+ Suivre';
            btn.style.background = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
            showNotification(`❌ Vous ne suivez plus ${artistName}`, 'info');
        } else {
            btn.textContent = '✓ Suivi';
            btn.style.background = 'linear-gradient(135deg, #4ade80 0%, #22c55e 100%)';
            showNotification(`✅ Vous suivez maintenant ${artistName}`, 'success');
        }
    }
    
    function handlePlayArtist(e) {
        const btn = e.currentTarget;
        const artistCard = btn.closest('.artist-card');
        const artistName = artistCard.querySelector('h4').textContent;
        
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            btn.style.transform = 'scale(1.1)';
        }, 150);
        
        showNotification(`🎵 Lecture de ${artistName}`, 'music');
    }
    
    // Actions des playlists
    function handleLikePlaylist(e) {
        const btn = e.currentTarget;
        const isLiked = btn.style.color === 'rgb(255, 107, 107)';
        
        btn.style.transform = 'scale(1.2)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 200);
        
        if (isLiked) {
            btn.style.color = '';
            showNotification('💔 Like retiré', 'info');
        } else {
            btn.style.color = '#ff6b6b';
            showNotification('❤️ Playlist likée !', 'success');
        }
    }
    
    function handleSavePlaylist(e) {
        const btn = e.currentTarget;
        const playlistCard = btn.closest('.playlist-discover-card');
        const playlistName = playlistCard.querySelector('h4').textContent;
        
        btn.style.transform = 'scale(1.2)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 200);
        
        btn.style.color = '#4ade80';
        showNotification(`💾 "${playlistName}" sauvegardée !`, 'success');
    }
    
    function handleSharePlaylist(e) {
        const btn = e.currentTarget;
        const playlistCard = btn.closest('.playlist-discover-card');
        const playlistName = playlistCard.querySelector('h4').textContent;
        
        btn.style.transform = 'scale(1.2)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 200);
        
        // Simuler le partage
        if (navigator.share) {
            navigator.share({
                title: playlistName,
                text: `Découvrez cette playlist sur Rythm'it !`,
                url: window.location.href
            });
        } else {
            // Fallback
            navigator.clipboard.writeText(window.location.href);
            showNotification('🔗 Lien copié dans le presse-papier !', 'success');
        }
    }
    
    // Exploration des recommandations
    function handleExploreRecommendation(e) {
        const btn = e.currentTarget;
        
        btn.style.transform = 'scale(0.98)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        showNotification('🎯 Chargement des recommandations...', 'info');
        
        // Simuler le chargement
        setTimeout(() => {
            showNotification('✨ Nouvelles recommandations chargées !', 'success');
        }, 1500);
    }
    
    // Actualisation des recommandations
    function refreshRecommendations() {
        refreshBtn.style.transform = 'rotate(360deg)';
        refreshBtn.style.transition = 'transform 0.8s ease';
        
        showNotification('🔄 Actualisation des recommandations...', 'info');
        
        setTimeout(() => {
            refreshBtn.style.transform = 'rotate(0deg)';
            generateNewRecommendations();
            showNotification('✅ Recommandations actualisées !', 'success');
        }, 800);
    }
    
    function generateNewRecommendations() {
        const recCards = document.querySelectorAll('.recommendation-card');
        recCards.forEach(card => {
            // Animer le changement
            card.style.opacity = '0.5';
            setTimeout(() => {
                // Simuler de nouvelles données
                const confidence = card.querySelector('.rec-confidence');
                if (confidence) {
                    const newPercentage = Math.floor(Math.random() * 20) + 80;
                    confidence.textContent = `${newPercentage}% de correspondance`;
                }
                card.style.opacity = '1';
            }, 400);
        });
    }
    
    // Exploration de genre
    function exploreGenre() {
        const btn = document.querySelector('.discover-genre-btn');
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        showNotification('🚀 Exploration du genre Synthwave...', 'music');
        
        // Mettre en évidence les éléments Synthwave
        setTimeout(() => {
            highlightGenreElements('electronic');
            showNotification('✨ Contenu Synthwave mis en évidence !', 'success');
        }, 1000);
    }
    
    function highlightGenreElements(genre) {
        const genreItems = document.querySelectorAll(`[data-genre="${genre}"]`);
        genreItems.forEach(item => {
            item.style.border = '2px solid rgba(255, 107, 107, 0.5)';
            item.style.transform = 'scale(1.02)';
            
            setTimeout(() => {
                item.style.border = '';
                item.style.transform = '';
            }, 3000);
        });
    }
    
    // Sessions live
    function joinLiveSession(e) {
        const btn = e.currentTarget;
        const session = btn.closest('.live-session');
        const sessionName = session.querySelector('h5').textContent;
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        btn.textContent = '🔴 En session';
        btn.style.background = 'linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%)';
        
        showNotification(`🎧 Vous avez rejoint "${sessionName}"`, 'success');
    }
    
    function notifyLiveSession(e) {
        const btn = e.currentTarget;
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        btn.textContent = '🔔 Notifié';
        showNotification('🔔 Vous serez notifié du début de la session', 'info');
    }
    
    // Effets de hover sur les cartes
    function handleCardHover(e) {
        const card = e.currentTarget;
        card.style.transition = 'all 0.3s ease';
        
        // Ajouter un effet de lueur subtile
        if (card.classList.contains('trending-item')) {
            card.style.boxShadow = '0 20px 50px rgba(102, 126, 234, 0.2)';
        } else {
            card.style.boxShadow = '0 15px 35px rgba(0, 0, 0, 0.3)';
        }
    }
    
    function handleCardLeave(e) {
        const card = e.currentTarget;
        card.style.boxShadow = '';
    }
    
    // Animations au chargement
    function animateOnLoad() {
        // Animer les sections une par une
        const sections = document.querySelectorAll('.discover-section');
        sections.forEach((section, index) => {
            section.style.opacity = '0';
            section.style.transform = 'translateY(30px)';
            
            setTimeout(() => {
                section.style.transition = 'opacity 0.8s ease, transform 0.8s ease';
                section.style.opacity = '1';
                section.style.transform = 'translateY(0)';
            }, 200 + index * 300);
        });
        
        // Animer les cartes dans chaque section
        setTimeout(() => {
            animateCards();
        }, 1000);
    }
    
    function animateCards() {
        const cards = document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card');
        cards.forEach((card, index) => {
            card.style.opacity = '0';
            card.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                card.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                card.style.opacity = '1';
                card.style.transform = 'translateY(0)';
            }, index * 100);
        });
    }
    
    // Chargement du contenu initial
    function loadDiscoverContent() {
        showNotification('🔍 Chargement du contenu de découverte...', 'info');
        
        setTimeout(() => {
            showNotification('✅ Contenu chargé avec succès !', 'success');
        }, 1500);
    }
    
    // Mises à jour dynamiques
    function startDynamicUpdates() {
        // Mettre à jour les statistiques des tendances
        setInterval(updateTrendingStats, 10000);
        
        // Simuler de nouveaux contenus
        setInterval(addNewContent, 30000);
        
        // Mettre à jour les sessions live
        setInterval(updateLiveSessions, 20000);
    }
    
    function updateTrendingStats() {
        const trendingStats = document.querySelectorAll('.trending-stats span');
        trendingStats.forEach(stat => {
            if (stat.textContent.includes('écoutes')) {
                const currentCount = parseInt(stat.textContent.match(/[\d.]+/)[0] * 1000000);
                const newCount = currentCount + Math.floor(Math.random() * 10000);
                const formattedCount = (newCount / 1000000).toFixed(1);
                stat.textContent = stat.textContent.replace(/[\d.]+/, formattedCount);
            }
        });
    }
    
    function addNewContent() {
        // Simuler l'ajout de nouveau contenu
        if (Math.random() < 0.3) {
            showNotification('✨ Nouveau contenu disponible !', 'info');
        }
    }
    
    function updateLiveSessions() {
        const liveSessions = document.querySelectorAll('.live-session');
        liveSessions.forEach(session => {
            const listener = session.querySelector('.session-info p');
            if (listener && listener.textContent.includes('auditeurs')) {
                const currentCount = parseInt(listener.textContent.match(/\d+/)[0]);
                const newCount = currentCount + Math.floor(Math.random() * 20) - 10;
                listener.textContent = `${Math.max(newCount, 0)} auditeurs`;
            }
        });
    }
    
    // Fonction de notification
    function showNotification(message, type = 'info') {
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
        } else {
            console.log(`Notification: ${message}`);
        }
    }
    
    // Raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // Ctrl + F pour focus sur la recherche
        if (e.ctrlKey && e.key === 'f') {
            e.preventDefault();
            globalSearch.focus();
        }
        
        // Espace pour pause/play si focus sur une carte
        if (e.key === ' ' && document.activeElement.classList.contains('play-btn-large')) {
            e.preventDefault();
            document.activeElement.click();
        }
        
        // Flèches pour naviguer entre les cartes
        if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
            const focusedCard = document.querySelector('.trending-item:focus, .artist-card:focus, .album-card:focus');
            if (focusedCard) {
                const allCards = Array.from(document.querySelectorAll('.trending-item, .artist-card, .album-card'));
                const currentIndex = allCards.indexOf(focusedCard);
                const nextIndex = e.key === 'ArrowRight' ? 
                    (currentIndex + 1) % allCards.length : 
                    (currentIndex - 1 + allCards.length) % allCards.length;
                allCards[nextIndex].focus();
            }
        }
    });
    
    // Rendre les cartes focusables
    document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card').forEach(card => {
        card.setAttribute('tabindex', '0');
    });
    
    console.log('🔍 Page Découverte Rythm\'it initialisée avec succès !');
    console.log('🎵 Fonctionnalités: Recherche avancée, Filtres dynamiques, Navigation améliorée');
});

// Fonctions pour la navigation dans la sidebar
function scrollToSection(sectionId) {
    const section = document.getElementById(sectionId);
    if (section) {
        section.scrollIntoView({ 
            behavior: 'smooth',
            block: 'start'
        });
        showNotification(`📍 Navigation vers ${sectionId}`, 'info');
    }
}

function filterByGenre(genre) {
    // Activer le filtre de genre
    const filterBtns = document.querySelectorAll('.filter-btn');
    filterBtns.forEach(btn => btn.classList.remove('active'));
    
    // Chercher s'il y a un bouton de filtre pour ce genre
    const genreBtn = Array.from(filterBtns).find(btn => 
        btn.textContent.toLowerCase().includes(genre.toLowerCase())
    );
    
    if (genreBtn) {
        genreBtn.classList.add('active');
    }
    
    // Filtrer les éléments par genre
    const allItems = document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card');
    allItems.forEach(item => {
        const itemGenre = item.getAttribute('data-genre');
        if (!itemGenre || itemGenre.toLowerCase().includes(genre.toLowerCase())) {
            item.style.display = 'block';
            item.style.opacity = '1';
        } else {
            item.style.opacity = '0.3';
        }
    });
    
    showNotification(`🎵 Filtré par genre: ${genre}`, 'success');
    
    // Scroll vers la section des résultats
    setTimeout(() => {
        const firstSection = document.querySelector('.discover-section');
        if (firstSection) {
            firstSection.scrollIntoView({ 
                behavior: 'smooth',
                block: 'start'
            });
        }
    }, 300);
}

// Export pour utilisation dans d'autres scripts
window.initializeEnhancedNavigation = initializeEnhancedNavigation;
window.handleInternalAnchor = handleInternalAnchor;
window.updateActiveNavigation = updateActiveNavigation;
window.scrollToSection = scrollToSection;
window.filterByGenre = filterByGenre;