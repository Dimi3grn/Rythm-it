// JavaScript pour la page Profil - profile.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let currentTab = 'overview';
    let isEditing = false;
    
    // Éléments DOM
    const tabItems = document.querySelectorAll('.profile-nav-item');
    const tabPanels = document.querySelectorAll('.tab-panel');
    const editProfileBtn = document.querySelector('.profile-action-btn.edit');
    const editModal = document.getElementById('editProfileModal');
    const playlistFilters = document.querySelectorAll('.filter-btn');
    const playlistCards = document.querySelectorAll('.playlist-card');
    
    // Initialisation
    init();
    
    function init() {
        // Charger l'onglet actuel
        showTab(currentTab);
        
        // Attacher les événements
        attachEventListeners();
        
        // Animer les éléments au chargement
        animateOnLoad();
        
        // Démarrer les mises à jour en temps réel
        startRealTimeUpdates();
    }
    
    // Gestion des événements
    function attachEventListeners() {
        // Navigation des onglets
        tabItems.forEach(item => {
            item.addEventListener('click', function(e) {
                e.preventDefault();
                const tabId = this.getAttribute('data-tab');
                showTab(tabId);
            });
        });
        
        // Bouton d'édition du profil
        if (editProfileBtn) {
            editProfileBtn.addEventListener('click', openEditModal);
        }
        
        // Actions du profil
        document.querySelectorAll('.profile-action-btn').forEach(btn => {
            btn.addEventListener('click', handleProfileAction);
        });
        
        // Lecture de pistes
        document.querySelectorAll('.play-track-btn, .play-btn').forEach(btn => {
            btn.addEventListener('click', handlePlayTrack);
        });
        
        // Filtres de playlists
        playlistFilters.forEach(filter => {
            filter.addEventListener('click', handlePlaylistFilter);
        });
        
        // Actions de cartes de playlists
        document.querySelectorAll('.card-action-btn').forEach(btn => {
            btn.addEventListener('click', handlePlaylistCardAction);
        });
        
        // Création de playlist
        const createPlaylistBtn = document.querySelector('.create-playlist-btn');
        if (createPlaylistBtn) {
            createPlaylistBtn.addEventListener('click', createNewPlaylist);
        }
        
        // Actions des playlists
        document.querySelectorAll('.playlist-action-btn').forEach(btn => {
            btn.addEventListener('click', handlePlaylistAction);
        });
        
        // Gestion du modal d'édition
        if (editModal) {
            attachModalEvents();
        }
        
        // ===== NOUVEAUX EVENT LISTENERS POUR LES UPLOADS =====
        console.log('🔧 Ajout des event listeners pour les uploads...');
        
        // Avatar upload
        const avatarUpload = document.getElementById('avatar-upload');
        if (avatarUpload) {
            console.log('✅ Avatar upload input trouvé');
            avatarUpload.addEventListener('change', function(e) {
                console.log('👤 ===== AVATAR UPLOAD CHANGE EVENT =====');
                console.log('📁 Fichier sélectionné:', e.target.files[0]);
                handleImageUpload(e, 'avatar');
            });
        } else {
            console.log('❌ Avatar upload input NOT FOUND');
        }
        
        // Banner upload
        const bannerUpload = document.getElementById('banner-upload');
        if (bannerUpload) {
            console.log('✅ Banner upload input trouvé');
            bannerUpload.addEventListener('change', function(e) {
                console.log('🖼️ ===== BANNER UPLOAD CHANGE EVENT =====');
                console.log('📁 Fichier sélectionné:', e.target.files[0]);
                handleImageUpload(e, 'banner');
            });
        } else {
            console.log('❌ Banner upload input NOT FOUND');
        }
        
        // Édition de couverture et avatar (anciens boutons - on les garde au cas où)
        const editCoverBtn = document.querySelector('.edit-cover-btn');
        const editAvatarBtn = document.querySelector('.edit-avatar-btn');
        
        if (editCoverBtn) {
            editCoverBtn.addEventListener('click', editCover);
        }
        
        if (editAvatarBtn) {
            editAvatarBtn.addEventListener('click', editAvatar);
        }
    }
    
    // Affichage des onglets
    function showTab(tabId) {
        // Mettre à jour la navigation
        tabItems.forEach(item => {
            item.classList.remove('active');
            if (item.getAttribute('data-tab') === tabId) {
                item.classList.add('active');
            }
        });
        
        // Mettre à jour les panneaux
        tabPanels.forEach(panel => {
            panel.classList.remove('active');
            if (panel.id === tabId) {
                panel.classList.add('active');
                
                // Animation d'entrée
                panel.style.opacity = '0';
                panel.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    panel.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                    panel.style.opacity = '1';
                    panel.style.transform = 'translateY(0)';
                }, 50);
            }
        });
        
        currentTab = tabId;
        
        // Charger le contenu spécifique à l'onglet
        loadTabContent(tabId);
    }
    
    // Chargement du contenu des onglets
    function loadTabContent(tabId) {
        switch(tabId) {
            case 'playlists':
                animatePlaylistCards();
                break;
            case 'stats':
                animateCharts();
                break;
            case 'activity':
                loadRecentActivity();
                break;
            case 'favorites':
                loadFavorites();
                break;
        }
    }
    
    // Actions du profil
    function handleProfileAction(e) {
        const btn = e.currentTarget;
        const action = btn.classList.contains('edit') ? 'edit' :
                      btn.classList.contains('share') ? 'share' : 'settings';
        
        // Animation de feedback
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(action) {
            case 'edit':
                openEditModal();
                break;
            case 'share':
                shareProfile();
                break;
            case 'settings':
                openSettings();
                break;
        }
    }
    
    // Lecture de pistes
    function handlePlayTrack(e) {
        const btn = e.currentTarget;
        const trackItem = btn.closest('.track-item, .favorite-item');
        
        if (trackItem) {
            const trackTitle = trackItem.querySelector('h5').textContent;
            const artist = trackItem.querySelector('p').textContent;
            
            // Changer l'icône
            const isPlaying = btn.textContent.includes('⏸️');
            btn.textContent = isPlaying ? '▶️' : '⏸️';
            
            // Animation
            btn.style.transform = 'scale(0.9)';
            setTimeout(() => {
                btn.style.transform = 'scale(1)';
            }, 150);
            
            // Notification
            if (!isPlaying) {
                showNotification(`🎵 Lecture: "${trackTitle}" par ${artist}`, 'music');
                updateCurrentlyPlaying(trackTitle, artist);
            } else {
                showNotification('⏸️ Lecture en pause', 'info');
            }
        }
    }
    
    // Mise à jour de la lecture en cours
    function updateCurrentlyPlaying(track, artist) {
        const currentTrackElement = document.querySelector('.current-track');
        if (currentTrackElement) {
            currentTrackElement.textContent = `${track} - ${artist}`;
        }
    }
    
    // Filtres de playlists
    function handlePlaylistFilter(e) {
        const filter = e.currentTarget;
        const filterType = filter.getAttribute('data-filter');
        
        // Mettre à jour les filtres actifs
        playlistFilters.forEach(f => f.classList.remove('active'));
        filter.classList.add('active');
        
        // Filtrer les playlists
        playlistCards.forEach(card => {
            const cardType = card.getAttribute('data-type');
            
            if (filterType === 'all' || cardType === filterType) {
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
    }
    
    // Actions des cartes de playlists
    function handlePlaylistCardAction(e) {
        const btn = e.currentTarget;
        const playlistCard = btn.closest('.playlist-card');
        const playlistTitle = playlistCard.querySelector('h4').textContent;
        
        // Animation de feedback
        btn.style.transform = 'scale(0.9)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        const action = btn.textContent.includes('▶️') ? 'play' :
                      btn.textContent.includes('📤') ? 'share' : 'menu';
        
        switch(action) {
            case 'play':
                playPlaylist(playlistTitle);
                break;
            case 'share':
                sharePlaylist(playlistTitle);
                break;
            case 'menu':
                showPlaylistMenu(btn, playlistTitle);
                break;
        }
    }
    
    // Lecture de playlist
    function playPlaylist(title) {
        showNotification(`🎵 Lecture de "${title}"`, 'music');
        
        // Mettre à jour l'icône de lecture
        const playBtn = document.querySelector('.card-action-btn');
        if (playBtn && playBtn.textContent.includes('▶️')) {
            playBtn.textContent = '⏸️';
        }
    }
    
    // Partage de playlist
    function sharePlaylist(title) {
        showNotification(`📤 "${title}" partagée !`, 'success');
    }
    
    // Menu de playlist
    function showPlaylistMenu(btn, title) {
        // Créer un menu contextuel
        const menu = document.createElement('div');
        menu.className = 'playlist-context-menu';
        menu.innerHTML = `
            <div class="menu-item" data-action="edit">✏️ Modifier</div>
            <div class="menu-item" data-action="duplicate">📋 Dupliquer</div>
            <div class="menu-item" data-action="export">📁 Exporter</div>
            <div class="menu-item danger" data-action="delete">🗑️ Supprimer</div>
        `;
        
        // Positionner le menu
        const rect = btn.getBoundingClientRect();
        menu.style.cssText = `
            position: fixed;
            top: ${rect.bottom + 5}px;
            right: ${window.innerWidth - rect.right}px;
            background: rgba(26, 26, 46, 0.95);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 8px 0;
            backdrop-filter: blur(10px);
            z-index: 1000;
            min-width: 150px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
        `;
        
        document.body.appendChild(menu);
        
        // Gestion des actions du menu
        menu.querySelectorAll('.menu-item').forEach(item => {
            item.addEventListener('click', () => {
                const action = item.getAttribute('data-action');
                handlePlaylistMenuAction(action, title);
                menu.remove();
            });
        });
        
        // Fermer en cliquant ailleurs
        setTimeout(() => {
            document.addEventListener('click', function closeMenu() {
                menu.remove();
                document.removeEventListener('click', closeMenu);
            });
        }, 100);
    }
    
    // Actions du menu de playlist
    function handlePlaylistMenuAction(action, title) {
        switch(action) {
            case 'edit':
                showNotification(`Édition de "${title}"`, 'info');
                break;
            case 'duplicate':
                showNotification(`"${title}" dupliquée !`, 'success');
                break;
            case 'export':
                showNotification(`"${title}" exportée !`, 'success');
                break;
            case 'delete':
                if (confirm(`Êtes-vous sûr de vouloir supprimer "${title}" ?`)) {
                    showNotification(`"${title}" supprimée`, 'info');
                }
                break;
        }
    }
    
    // Création de nouvelle playlist
    function createNewPlaylist() {
        const playlistName = prompt('Nom de la nouvelle playlist:');
        if (playlistName && playlistName.trim()) {
            showNotification(`Playlist "${playlistName}" créée !`, 'success');
            
            // Ajouter une nouvelle carte (simulation)
            setTimeout(() => {
                addNewPlaylistCard(playlistName);
            }, 500);
        }
    }
    
    // Ajouter une nouvelle carte de playlist
    function addNewPlaylistCard(name) {
        const playlistsGrid = document.querySelector('.playlists-grid');
        if (playlistsGrid) {
            const newCard = document.createElement('div');
            newCard.className = 'playlist-card';
            newCard.setAttribute('data-type', 'public');
            newCard.innerHTML = `
                <div class="playlist-cover-card"></div>
                <div class="playlist-card-info">
                    <h4>🎵 ${name}</h4>
                    <p>0 titres • 0min</p>
                    <div class="playlist-meta">
                        <span class="playlist-visibility">🌍 Publique</span>
                        <span class="playlist-plays">0 écoutes</span>
                    </div>
                </div>
                <div class="playlist-card-actions">
                    <button class="card-action-btn">▶️</button>
                    <button class="card-action-btn">📤</button>
                    <button class="card-action-btn">⋯</button>
                </div>
            `;
            
            // Animation d'entrée
            newCard.style.opacity = '0';
            newCard.style.transform = 'translateY(30px)';
            playlistsGrid.appendChild(newCard);
            
            setTimeout(() => {
                newCard.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                newCard.style.opacity = '1';
                newCard.style.transform = 'translateY(0)';
            }, 100);
            
            // Attacher les événements
            newCard.querySelectorAll('.card-action-btn').forEach(btn => {
                btn.addEventListener('click', handlePlaylistCardAction);
            });
        }
    }
    
    // Actions de playlist
    function handlePlaylistAction(e) {
        const btn = e.currentTarget;
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        if (btn.textContent.includes('Écouter')) {
            showNotification('🎵 Lecture de la playlist tendance', 'music');
            btn.innerHTML = '⏸️ Pause';
        }
    }
    
    // Modal d'édition
    function openEditModal() {
        if (editModal) {
            editModal.classList.add('active');
            isEditing = true;
            
            // Focus sur le premier champ
            const firstInput = editModal.querySelector('.form-input');
            if (firstInput) {
                setTimeout(() => firstInput.focus(), 300);
            }
        }
    }
    
    function closeEditModal() {
        if (editModal) {
            editModal.classList.remove('active');
            isEditing = false;
        }
    }
    
    // Événements du modal d'édition
    function attachModalEvents() {
        const closeBtn = editModal.querySelector('.close-modal');
        const cancelBtn = editModal.querySelector('.cancel-btn');
        const saveBtn = editModal.querySelector('.save-btn');
        
        if (closeBtn) {
            closeBtn.addEventListener('click', closeEditModal);
        }
        
        if (cancelBtn) {
            cancelBtn.addEventListener('click', closeEditModal);
        }
        
        if (saveBtn) {
            saveBtn.addEventListener('click', saveProfile);
        }
        
        // Fermer en cliquant à l'extérieur
        editModal.addEventListener('click', (e) => {
            if (e.target === editModal) {
                closeEditModal();
            }
        });
    }
    
    // Sauvegarde du profil
    function saveProfile() {
        console.log('💾 ===== DEBUT saveProfile =====');
        
        const inputs = editModal.querySelectorAll('.form-input, .form-textarea');
        const data = {};
        
        inputs.forEach(input => {
            const label = input.closest('.form-group').querySelector('label').textContent;
            data[label] = input.value;
            console.log('📝 Champ texte:', label, '=', input.value);
        });
        
        // Récupérer les URLs des images depuis les champs cachés
        const avatarUrl = document.getElementById('avatar-url')?.value;
        const bannerUrl = document.getElementById('banner-url')?.value;
        
        console.log('🖼️ Avatar URL:', avatarUrl);
        console.log('🖼️ Banner URL:', bannerUrl);
        
        // Animation de sauvegarde
        const saveBtn = editModal.querySelector('.save-btn');
        const originalText = saveBtn.innerHTML;
        saveBtn.innerHTML = '💾 Sauvegarde...';
        saveBtn.disabled = true;
        
        // Préparer les données pour l'envoi au backend
        const formData = new URLSearchParams();
        
        // Ajouter les champs texte
        if (data['Nom d\'affichage']) {
            formData.append('display_name', data['Nom d\'affichage']);
        }
        if (data['Nom d\'utilisateur']) {
            formData.append('username', data['Nom d\'utilisateur']);
        }
        if (data['Statut actuel']) {
            formData.append('status', data['Statut actuel']);
        }
        
        // Ajouter les images si elles existent
        if (avatarUrl) {
            formData.append('avatar_image', avatarUrl);
            console.log('➕ Avatar ajouté au formulaire');
        }
        if (bannerUrl) {
            formData.append('banner_image', bannerUrl);
            console.log('➕ Banner ajouté au formulaire');
        }
        
        console.log('📦 Données à envoyer:', formData.toString());
        
        // Envoyer au backend
        fetch('/profile', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData.toString(),
            credentials: 'same-origin'
        })
        .then(response => {
            console.log('📥 Réponse saveProfile - Status:', response.status, 'OK:', response.ok);
            if (!response.ok) {
                throw new Error('Erreur sauvegarde profil: ' + response.status);
            }
            return response.text();
        })
        .then((responseText) => {
            console.log('📥 REPONSE saveProfile (text):', responseText);
            
            showNotification('✅ Profil mis à jour avec succès !', 'success');
            closeEditModal();
            
            // Mettre à jour l'affichage
            updateProfileDisplay(data);
            
            // Mettre à jour l'affichage des images
            if (avatarUrl) {
                updateAvatarDisplay(avatarUrl);
            }
            if (bannerUrl) {
                updateBannerDisplay(bannerUrl);
            }
            
            saveBtn.innerHTML = originalText;
            saveBtn.disabled = false;
            
            // Recharger la page pour voir les changements
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        })
        .catch(error => {
            console.error('❌ ERREUR saveProfile:', error);
            showNotification('Erreur lors de la sauvegarde: ' + error.message, 'error');
            
            saveBtn.innerHTML = originalText;
            saveBtn.disabled = false;
        });
    }
    
    // Mise à jour de l'affichage du profil
    function updateProfileDisplay(data) {
        // Mettre à jour le nom d'affichage
        if (data['Nom d\'affichage']) {
            const nameElement = document.querySelector('.profile-basic-info h1');
            if (nameElement) {
                nameElement.textContent = data['Nom d\'affichage'];
            }
        }
        
        // Mettre à jour le nom d'utilisateur
        if (data['Nom d\'utilisateur']) {
            const usernameElement = document.querySelector('.profile-basic-info p');
            if (usernameElement) {
                usernameElement.textContent = '@' + data['Nom d\'utilisateur'];
            }
        }
        
        // Mettre à jour le statut
        if (data['Statut actuel']) {
            const statusElement = document.querySelector('.status-text');
            if (statusElement) {
                statusElement.textContent = data['Statut actuel'];
            }
        }
    }
    
    // Édition de couverture
    function editCover() {
        console.log('🎬 editCover() appelée');
        // Créer un input file temporaire
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = 'image/*';
        fileInput.style.display = 'none';
        
        fileInput.addEventListener('change', function(e) {
            console.log('📁 Fichier sélectionné pour cover:', e.target.files[0]);
            const file = e.target.files[0];
            if (!file) {
                console.log('❌ Aucun fichier sélectionné');
                return;
            }
            
            // Validation du fichier
            if (!validateImageFile(file)) {
                console.log('❌ Fichier invalide');
                return;
            }
            
            console.log('✅ Fichier valide, upload en cours...');
            // Upload du fichier
            uploadProfileImage(file, 'banner');
        });
        
        // Déclencher la sélection de fichier
        console.log('🔽 Ouverture du sélecteur de fichier...');
        fileInput.click();
    }
    
    // Édition d'avatar
    function editAvatar() {
        console.log('👤 editAvatar() appelée');
        // Créer un input file temporaire
        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.accept = 'image/*';
        fileInput.style.display = 'none';
        
        fileInput.addEventListener('change', function(e) {
            console.log('📁 Fichier sélectionné pour avatar:', e.target.files[0]);
            const file = e.target.files[0];
            if (!file) {
                console.log('❌ Aucun fichier sélectionné');
                return;
            }
            
            // Validation du fichier
            if (!validateImageFile(file)) {
                console.log('❌ Fichier invalide');
                return;
            }
            
            console.log('✅ Fichier valide, upload en cours...');
            // Upload du fichier
            uploadProfileImage(file, 'avatar');
        });
        
        // Déclencher la sélection de fichier
        console.log('🔽 Ouverture du sélecteur de fichier...');
        fileInput.click();
    }
    
    // Validation du fichier image
    function validateImageFile(file) {
        const maxSize = 5 * 1024 * 1024; // 5MB
        const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
        
        if (!allowedTypes.includes(file.type)) {
            showNotification('Format d\'image non supporté. Utilisez JPG, PNG, GIF ou WebP.', 'error');
            return false;
        }
        
        if (file.size > maxSize) {
            showNotification('L\'image est trop volumineuse. Taille maximum: 5MB.', 'error');
            return false;
        }
        
        return true;
    }
    
    // Upload d'image de profil
    function uploadProfileImage(file, type) {
        console.log('🚀 DEBUT uploadProfileImage - Type:', type, 'File:', file.name, 'Size:', file.size);
        
        const formData = new FormData();
        formData.append('image', file);
        formData.append('type', 'profile');
        
        console.log('📦 FormData créé avec image et type=profile');
        
        // Afficher un indicateur de chargement
        showNotification('📤 Téléchargement en cours...', 'info');
        
        console.log('📤 ENVOI requête vers /upload/image');
        
        fetch('/upload/image', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin'
        })
        .then(response => {
            console.log('📥 Réponse upload reçue - Status:', response.status, 'OK:', response.ok);
            if (!response.ok) {
                throw new Error('Erreur upload: ' + response.status);
            }
            return response.json();
        })
        .then(data => {
            console.log('📥 DONNEES upload reçues:', JSON.stringify(data, null, 2));
            if (data.success && data.imageUrl) {
                console.log('✅ Upload réussi! URL:', data.imageUrl);
                console.log('🔄 APPEL updateProfileImage avec type:', type, 'URL:', data.imageUrl);
                // Mettre à jour le profil avec la nouvelle image
                updateProfileImage(type, data.imageUrl);
                showNotification('✅ Image téléchargée avec succès !', 'success');
            } else {
                console.log('❌ Upload échoué - Data:', data);
                throw new Error(data.message || 'Erreur inconnue');
            }
        })
        .catch(error => {
            console.error('❌ ERREUR uploadProfileImage:', error);
            showNotification('Erreur lors du téléchargement de l\'image: ' + error.message, 'error');
        });
    }
    
    // Mise à jour de l'image de profil
    function updateProfileImage(type, imageUrl) {
        console.log('🔄 DEBUT updateProfileImage - Type:', type, 'URL:', imageUrl);
        
        // Créer les données du formulaire
        const formData = new URLSearchParams();
        const action = type === 'avatar' ? 'update_avatar' : 'update_banner';
        const fieldName = type === 'avatar' ? 'avatar_image' : 'banner_image';
        
        formData.append('action', action);
        formData.append(fieldName, imageUrl);
        
        console.log('📦 URLSearchParams créé:');
        console.log('  - action:', action);
        console.log('  - ' + fieldName + ':', imageUrl);
        console.log('📦 FormData string:', formData.toString());
        
        fetch('/profile/action', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: formData.toString(),
            credentials: 'same-origin'
        })
        .then(response => {
            console.log('📥 Réponse profile update - Status:', response.status, 'OK:', response.ok);
            if (!response.ok) {
                throw new Error('Erreur mise à jour profil: ' + response.status);
            }
            return response.text();
        })
        .then((responseText) => {
            console.log('📥 REPONSE profile update (text):', responseText);
            // Mettre à jour l'affichage
            if (type === 'avatar') {
                console.log('🖼️ Mise à jour affichage avatar');
                updateAvatarDisplay(imageUrl);
            } else {
                console.log('🖼️ Mise à jour affichage banner');
                updateBannerDisplay(imageUrl);
            }
            console.log('✅ updateProfileImage TERMINE avec succès');
        })
        .catch(error => {
            console.error('❌ ERREUR updateProfileImage:', error);
            showNotification('Erreur lors de la mise à jour du profil: ' + error.message, 'error');
        });
    }
    
    // Mise à jour de l'affichage de l'avatar
    function updateAvatarDisplay(imageUrl) {
        const avatarElements = document.querySelectorAll('.user-pic.profile-size');
        avatarElements.forEach(avatar => {
            // Remplacer le texte par une image
            avatar.innerHTML = `<img src="${imageUrl}" alt="Avatar" style="width: 100%; height: 100%; object-fit: cover; border-radius: 50%;">`;
        });
    }
    
    // Mise à jour de l'affichage de la bannière
    function updateBannerDisplay(imageUrl) {
        const coverElement = document.querySelector('.profile-cover');
        if (coverElement) {
            coverElement.style.background = `url(${imageUrl}) center/cover no-repeat`;
        }
    }
    
    // Partage de profil
    function shareProfile() {
        // Simulation de partage
        if (navigator.share) {
            navigator.share({
                title: 'Mon profil Rythm\'it',
                text: 'Découvrez ma musique sur Rythm\'it !',
                url: window.location.href
            });
        } else {
            // Fallback: copier dans le presse-papier
            navigator.clipboard.writeText(window.location.href).then(() => {
                showNotification('🔗 Lien du profil copié !', 'success');
            });
        }
    }
    
    // Paramètres
    function openSettings() {
        showNotification('⚙️ Ouverture des paramètres...', 'info');
    }
    
    // Animations au chargement
    function animateOnLoad() {
        // Animer les cartes de profil
        const profileCards = document.querySelectorAll('.profile-card');
        profileCards.forEach((card, index) => {
            card.style.opacity = '0';
            card.style.transform = 'translateY(30px)';
            
            setTimeout(() => {
                card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
                card.style.opacity = '1';
                card.style.transform = 'translateY(0)';
            }, 100 + index * 150);
        });
        
        // Animer les barres de progression
        setTimeout(() => {
            document.querySelectorAll('.genre-progress, .progress-fill').forEach(bar => {
                const width = bar.style.width;
                bar.style.width = '0';
                setTimeout(() => {
                    bar.style.width = width;
                }, 500);
            });
        }, 1000);
    }
    
    // Animation des cartes de playlists
    function animatePlaylistCards() {
        const cards = document.querySelectorAll('.playlist-card');
        cards.forEach((card, index) => {
            card.style.opacity = '0';
            card.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                card.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
                card.style.opacity = '1';
                card.style.transform = 'translateY(0)';
            }, index * 100);
        });
    }
    
    // Animation des graphiques
    function animateCharts() {
        const chartBars = document.querySelectorAll('.chart-bar');
        chartBars.forEach((bar, index) => {
            const height = bar.style.height;
            bar.style.height = '0';
            
            setTimeout(() => {
                bar.style.transition = 'height 0.8s ease';
                bar.style.height = height;
            }, index * 200);
        });
    }
    
    // Chargement de l'activité récente
    function loadRecentActivity() {
        const activityItems = document.querySelectorAll('.activity-item');
        activityItems.forEach((item, index) => {
            item.style.opacity = '0';
            item.style.transform = 'translateX(-20px)';
            
            setTimeout(() => {
                item.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
                item.style.opacity = '1';
                item.style.transform = 'translateX(0)';
            }, index * 150);
        });
    }
    
    // Chargement des favoris
    function loadFavorites() {
        showNotification('🎵 Chargement de vos favoris...', 'info');
    }
    
    // Mises à jour en temps réel
    function startRealTimeUpdates() {
        // Mettre à jour les statistiques périodiquement
        setInterval(() => {
            updateStats();
        }, 30000);
        
        // Simuler des nouveaux badges
        setTimeout(() => {
            checkNewBadges();
        }, 10000);
    }
    
    // Mise à jour des statistiques
    function updateStats() {
        const statValues = document.querySelectorAll('.stat-value');
        statValues.forEach(stat => {
            if (Math.random() < 0.3) { // 30% de chance de mise à jour
                const currentValue = parseInt(stat.textContent);
                const newValue = currentValue + Math.floor(Math.random() * 5) + 1;
                
                // Animation de mise à jour
                stat.style.transform = 'scale(1.2)';
                setTimeout(() => {
                    stat.textContent = newValue.toLocaleString();
                    stat.style.transform = 'scale(1)';
                }, 200);
            }
        });
    }
    
    // Vérification de nouveaux badges
    function checkNewBadges() {
        const lockedBadges = document.querySelectorAll('.badge-item.locked');
        if (lockedBadges.length > 0 && Math.random() < 0.5) {
            const badge = lockedBadges[0];
            const badgeName = badge.querySelector('h5').textContent;
            
            badge.classList.remove('locked');
            badge.classList.add('earned');
            
            // Animation de déverrouillage
            badge.style.transform = 'scale(1.1)';
            setTimeout(() => {
                badge.style.transform = 'scale(1)';
            }, 300);
            
            showNotification(`🏆 Nouveau badge débloqué: ${badgeName}!`, 'success');
        }
    }
    
    // Fonction de notification
    function showNotification(message, type = 'info') {
        // Utiliser la fonction globale si disponible
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
        } else {
            console.log(`Notification: ${message}`);
        }
    }
    
    // Raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // Échap pour fermer les modaux
        if (e.key === 'Escape' && isEditing) {
            closeEditModal();
        }
        
        // Ctrl + E pour éditer le profil
        if (e.ctrlKey && e.key === 'e') {
            e.preventDefault();
            openEditModal();
        }
        
        // Touches 1-5 pour naviguer entre les onglets
        if (e.key >= '1' && e.key <= '5') {
            const tabIndex = parseInt(e.key) - 1;
            const tabItem = tabItems[tabIndex];
            if (tabItem) {
                const tabId = tabItem.getAttribute('data-tab');
                showTab(tabId);
            }
        }
    });
    
    // ===== NOUVELLE FONCTION POUR GERER LES UPLOADS =====
    function handleImageUpload(event, type) {
        console.log(`🚀 ===== DEBUT handleImageUpload - Type: ${type} =====`);
        
        const file = event.target.files[0];
        if (!file) {
            console.log('❌ Aucun fichier sélectionné');
            return;
        }
        
        console.log('📁 Fichier:', file.name, 'Taille:', file.size, 'Type:', file.type);
        
        // Validation du fichier
        if (!validateImageFile(file)) {
            console.log('❌ Fichier invalide');
            return;
        }
        
        console.log('✅ Fichier valide, début upload...');
        
        // Upload du fichier
        uploadProfileImageFromModal(file, type);
    }
    
    // ===== NOUVELLE FONCTION UPLOAD DEPUIS LE MODAL =====
    function uploadProfileImageFromModal(file, type) {
        console.log(`🚀 ===== DEBUT uploadProfileImageFromModal - Type: ${type} =====`);
        
        const formData = new FormData();
        formData.append('image', file);
        formData.append('type', 'profile');
        
        console.log('📦 FormData créé avec image et type=profile');
        
        // Afficher un indicateur de chargement
        showNotification('📤 Téléchargement en cours...', 'info');
        
        console.log('📤 ENVOI requête vers /upload/image');
        
        fetch('/upload/image', {
            method: 'POST',
            body: formData,
            credentials: 'same-origin'
        })
        .then(response => {
            console.log('📥 Réponse upload reçue - Status:', response.status, 'OK:', response.ok);
            if (!response.ok) {
                throw new Error('Erreur upload: ' + response.status);
            }
            return response.json();
        })
        .then(data => {
            console.log('📥 DONNEES upload reçues:', JSON.stringify(data, null, 2));
            if (data.success && data.imageUrl) {
                console.log('✅ Upload réussi! URL:', data.imageUrl);
                
                // Mettre à jour le modal avec la nouvelle image
                updateModalImagePreview(type, data.imageUrl);
                
                // Mettre à jour le champ hidden
                updateHiddenImageField(type, data.imageUrl);
                
                showNotification('✅ Image téléchargée avec succès !', 'success');
                
                console.log('✅ uploadProfileImageFromModal TERMINE avec succès');
            } else {
                console.log('❌ Upload échoué - Data:', data);
                throw new Error(data.message || 'Erreur inconnue');
            }
        })
        .catch(error => {
            console.error('❌ ERREUR uploadProfileImageFromModal:', error);
            showNotification('Erreur lors du téléchargement de l\'image: ' + error.message, 'error');
        });
    }
    
    // ===== FONCTION POUR METTRE A JOUR L'APERCU DANS LE MODAL =====
    function updateModalImagePreview(type, imageUrl) {
        console.log(`🖼️ ===== updateModalImagePreview - Type: ${type}, URL: ${imageUrl} =====`);
        
        const previewId = type === 'avatar' ? 'avatar-preview' : 'banner-preview';
        const preview = document.getElementById(previewId);
        
        if (preview) {
            console.log('✅ Preview element trouvé:', previewId);
            preview.innerHTML = `<img src="${imageUrl}" alt="${type} preview" style="width: 100%; height: 100%; object-fit: cover;">`;
            console.log('✅ Preview mis à jour');
        } else {
            console.log('❌ Preview element NOT FOUND:', previewId);
        }
    }
    
    // ===== FONCTION POUR METTRE A JOUR LE CHAMP HIDDEN =====
    function updateHiddenImageField(type, imageUrl) {
        console.log(`📝 ===== updateHiddenImageField - Type: ${type}, URL: ${imageUrl} =====`);
        
        const fieldId = type === 'avatar' ? 'avatar-url' : 'banner-url';
        const hiddenField = document.getElementById(fieldId);
        
        if (hiddenField) {
            console.log('✅ Hidden field trouvé:', fieldId);
            hiddenField.value = imageUrl;
            console.log('✅ Hidden field mis à jour avec:', imageUrl);
        } else {
            console.log('❌ Hidden field NOT FOUND:', fieldId);
        }
    }
    
    console.log('👤 Page Profil Rythm\'it initialisée avec succès !');
    console.log('🎯 Fonctionnalités: Onglets dynamiques, Édition de profil, Statistiques en temps réel');
});