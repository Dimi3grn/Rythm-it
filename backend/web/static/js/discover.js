// JavaScript pour la page Découverte - discover.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let currentFilter = 'all';
    let currentCategory = 'all';
    let currentGenre = '';
    let searchTimeout;
    let isVoiceSearchActive = false;
    
    // Éléments DOM
    const globalSearch = document.querySelector('.global-search');
    const searchBtn = document.querySelector('.search-btn');
    const voiceSearchBtn = document.querySelector('.voice-search-btn');
    const filterBtns = document.querySelectorAll('.filter-btn');
    const categoryFilters = document.querySelectorAll('.filter-category');
    const genreFilters = document.querySelectorAll('.filter-genre');
    const refreshBtn = document.querySelector('.refresh-recommendations');
    
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
    }
    
    // Gestion des événements
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
                setActiveFilter(filter);
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
        
        // Filtres de genres
        genreFilters.forEach(filter => {
            filter.addEventListener('click', function(e) {
                e.preventDefault();
                const genre = this.getAttribute('data-genre');
                setActiveGenre(genre);
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
        const query = globalSearch.value.trim();
        if (query) {
            performSearchWithQuery(query);
        }
    }
    
    function performSearchWithQuery(query) {
        showNotification(`🔍 Recherche: "${query}"`, 'info');
        
        // Simuler les résultats de recherche
        setTimeout(() => {
            highlightSearchResults(query);
            showNotification(`✅ ${Math.floor(Math.random() * 50) + 10} résultats trouvés`, 'success');
        }, 800);
    }
    
    function highlightSearchResults(query) {
        // Animer les cartes qui correspondent à la recherche
        const allCards = document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card');
        allCards.forEach((card, index) => {
            if (Math.random() < 0.3) { // 30% de correspondance simulée
                card.style.border = '2px solid rgba(102, 126, 234, 0.5)';
                card.style.transform = 'scale(1.02)';
                
                setTimeout(() => {
                    card.style.border = '';
                    card.style.transform = '';
                }, 2000);
            }
        });
    }
    
    function resetSearch() {
        // Réinitialiser l'affichage
        const allCards = document.querySelectorAll('.trending-item, .artist-card, .album-card, .playlist-discover-card');
        allCards.forEach(card => {
            card.style.display = 'block';
            card.style.opacity = '1';
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
        document.querySelector(`[data-filter="${filter}"]`).classList.add('active');
        
        currentFilter = filter;
        applyFilters();
    }
    
    function setActiveCategory(category) {
        // Mettre à jour la navigation
        categoryFilters.forEach(filter => filter.classList.remove('active'));
        document.querySelector(`[data-category="${category}"]`).classList.add('active');
        
        currentCategory = category;
        applyFilters();
    }
    
    function setActiveGenre(genre) {
        // Mettre à jour la navigation
        genreFilters.forEach(filter => filter.classList.remove('active'));
        document.querySelector(`[data-genre="${genre}"]`).classList.add('active');
        
        currentGenre = genre;
        applyFilters();
    }
    
    function applyFilters() {
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
    console.log('🎵 Fonctionnalités: Recherche avancée, Filtres dynamiques, Recommandations personnalisées');
});