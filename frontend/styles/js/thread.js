// JavaScript pour la page Thread - thread.js

// Variables globales
let isLiked = true;
let likeCount = 34;
let commentCount = 8;
let shareCount = 12;
let isPlaying = false;
let currentReplyingTo = null;

// Fonction pour extraire l'ID du thread depuis l'URL
function getThreadIdFromURL() {
    const path = window.location.pathname;
    const match = path.match(/\/thread\/(\d+)/);
    return match ? match[1] : null;
}

// Gestion des likes - FONCTION GLOBALE
async function toggleLike(btn) {
    const threadId = getThreadIdFromURL();
    if (!threadId) {
        console.error('ID du thread non trouvé');
        return;
    }

    const countElement = btn.querySelector('.btn-count');
    const iconElement = btn.querySelector('.btn-icon');
    
    // Désactiver le bouton temporairement
    btn.disabled = true;
    
    try {
        console.log(`🎯 Tentative de like pour thread ${threadId}`);
        
        // Appel à l'API backend
        const response = await fetch(`/api/v1/threads/${threadId}/like`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include' // Important pour envoyer les cookies d'auth
        });

        if (!response.ok) {
            throw new Error(`Erreur HTTP: ${response.status}`);
        }

        const data = await response.json();
        console.log('📊 Réponse like API:', data);

        if (data.success) {
            // Mettre à jour l'UI avec les données du serveur
            isLiked = data.liked;
            likeCount = data.likes_count;
            
            if (data.liked) {
                btn.classList.add('liked');
                iconElement.textContent = '❤️';
                showNotification('Post liké ! ❤️', 'success');
                createHeartAnimation(btn);
            } else {
                btn.classList.remove('liked');
                iconElement.textContent = '🤍';
                showNotification('Like retiré', 'info');
            }
            
            countElement.textContent = likeCount;
            updateLikedBySection();
        } else {
            console.error('❌ Erreur serveur:', data.message);
            showNotification('Erreur lors du like', 'error');
        }

    } catch (error) {
        console.error('❌ Erreur réseau like:', error);
        showNotification('Erreur de connexion', 'error');
    } finally {
        // Réactiver le bouton
        btn.disabled = false;
    }
}

// Animation de coeur
function createHeartAnimation(btn) {
    const heart = document.createElement('div');
    heart.textContent = '❤️';
    heart.style.cssText = `
        position: absolute;
        font-size: 20px;
        pointer-events: none;
        z-index: 1000;
        animation: heartFloat 1s ease-out forwards;
    `;
    
    const rect = btn.getBoundingClientRect();
    heart.style.left = rect.left + rect.width / 2 + 'px';
    heart.style.top = rect.top + 'px';
    
    document.body.appendChild(heart);
    
    setTimeout(() => {
        heart.remove();
    }, 1000);
}

// Mettre à jour la section "Aimé par"
function updateLikedBySection() {
    const likedBy = document.querySelector('.liked-by');
    if (likedBy) {
        const moreLikes = likedBy.querySelector('.more-likes');
        if (moreLikes) {
            const otherLikesCount = Math.max(likeCount - 3, 0);
            if (otherLikesCount > 0) {
                moreLikes.textContent = `et ${otherLikesCount} autres`;
            } else {
                moreLikes.textContent = '';
            }
        }
    }
}

document.addEventListener('DOMContentLoaded', function() {
    // Éléments DOM
    const engagementBtns = document.querySelectorAll('.engagement-btn');
    const playControls = document.querySelectorAll('.play-control, .mini-play-btn');
    const commentInput = document.querySelector('.comment-input');
    const commentBtn = document.querySelector('.comment-btn');
    const commentsList = document.querySelector('.comments-list');
    const loadMoreBtn = document.querySelector('.load-more-btn');
    const filterSelect = document.querySelector('.filter-select');
    const musicShareModal = document.getElementById('musicShareModal');
    
    // Initialisation
    init();
    
    function init() {
        // Attacher les événements
        attachEventListeners();
        
        // Animer l'entrée
        animateOnLoad();
        
        // Charger les données du thread depuis l'URL
        loadThreadFromURL();
        
        // Démarrer les mises à jour en temps réel
        startRealTimeUpdates();
    }
    
    // Gestion des événements
    function attachEventListeners() {
        // Boutons d'engagement du thread principal
        engagementBtns.forEach(btn => {
            btn.addEventListener('click', handleEngagement);
        });
        
        // Contrôles de lecture
        playControls.forEach(control => {
            control.addEventListener('click', handlePlayControl);
        });
        
        // Commentaire
        if (commentBtn) {
            commentBtn.addEventListener('click', submitComment);
        }
        
        if (commentInput) {
            commentInput.addEventListener('input', handleCommentInput);
            commentInput.addEventListener('keypress', function(e) {
                if (e.key === 'Enter' && e.ctrlKey) {
                    submitComment();
                }
            });
        }
        
        // Actions des commentaires existants
        attachCommentEventListeners();
        
        // Charger plus de commentaires
        if (loadMoreBtn) {
            loadMoreBtn.addEventListener('click', loadMoreComments);
        }
        
        // Filtre des commentaires - NOUVEAU
        initializeCommentsFilter();
        
        // Filtre des commentaires (ancien)
        if (filterSelect) {
            filterSelect.addEventListener('change', filterComments);
        }
        
        // Actions du thread
        document.querySelectorAll('.action-btn').forEach(btn => {
            btn.addEventListener('click', handleThreadAction);
        });
        
        // Actions de la track
        document.querySelectorAll('.track-btn').forEach(btn => {
            btn.addEventListener('click', handleTrackAction);
        });
        
        // Boutons de la sidebar
        document.querySelectorAll('.follow-btn, .message-btn').forEach(btn => {
            btn.addEventListener('click', handleSidebarAction);
        });
        
        // Tags tendance
        document.querySelectorAll('.trend-tag').forEach(tag => {
            tag.addEventListener('click', handleTrendTagClick);
        });
        
        // Outils du compositeur
        document.querySelectorAll('.tool-btn').forEach(btn => {
            btn.addEventListener('click', handleComposerTool);
        });
        
        // Modal de partage de musique
        if (musicShareModal) {
            attachModalEventListeners();
        }
    }
    
    // Gestion de l'engagement (like, comment, share, etc.)
    function handleEngagement(e) {
        const btn = e.currentTarget;
        const action = btn.querySelector('.btn-label').textContent.toLowerCase();
        
        // Animation de feedback
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(action) {
            case 'j\'aime':
                toggleLike(btn);
                break;
            case 'commentaires':
                scrollToComments();
                break;
            case 'partages':
                shareThread();
                break;
            case 'sauvegarder':
                saveThread(btn);
                break;
        }
    }
    
    // Scroll vers les commentaires
    function scrollToComments() {
        const commentsSection = document.querySelector('.comments-section');
        if (commentsSection) {
            commentsSection.scrollIntoView({ 
                behavior: 'smooth',
                block: 'start'
            });
            
            // Highlight temporaire
            commentsSection.style.border = '2px solid rgba(102, 126, 234, 0.5)';
            setTimeout(() => {
                commentsSection.style.border = '';
            }, 2000);
        }
    }
    
    // Partage du thread
    function shareThread() {
        shareCount++;
        const shareBtn = document.querySelector('.engagement-btn .btn-label[textContent="Partages"]');
        if (shareBtn) {
            const countElement = shareBtn.parentElement.querySelector('.btn-count');
            countElement.textContent = shareCount;
        }
        
        // Simuler le partage
        if (navigator.share) {
            navigator.share({
                title: 'Thread Rythm\'it',
                text: 'Découvrez cette incroyable musique ambient !',
                url: window.location.href
            });
        } else {
            // Fallback: copier le lien
            navigator.clipboard.writeText(window.location.href).then(() => {
                showNotification('🔗 Lien copié dans le presse-papier !', 'success');
            });
        }
    }
    
    // Sauvegarder le thread
    function saveThread(btn) {
        const isSaved = btn.classList.contains('saved');
        
        if (isSaved) {
            btn.classList.remove('saved');
            btn.querySelector('.btn-icon').textContent = '🔖';
            showNotification('Thread retiré des favoris', 'info');
        } else {
            btn.classList.add('saved');
            btn.querySelector('.btn-icon').textContent = '✅';
            showNotification('Thread sauvegardé ! 🔖', 'success');
        }
    }
    
    // Contrôles de lecture
    function handlePlayControl(e) {
        const control = e.currentTarget;
        
        // Animation
        control.style.transform = 'scale(0.9)';
        setTimeout(() => {
            control.style.transform = 'scale(1)';
        }, 150);
        
        if (isPlaying) {
            // Pause
            control.textContent = '▶️';
            isPlaying = false;
            showNotification('⏸️ Lecture en pause', 'info');
            stopPlayingAnimation();
        } else {
            // Play
            control.textContent = '⏸️';
            isPlaying = true;
            showNotification('🎵 Lecture: "Ethereal Landscapes"', 'music');
            startPlayingAnimation();
        }
    }
    
    // Animation de lecture
    function startPlayingAnimation() {
        const musicCard = document.querySelector('.main-thread .music-card');
        if (musicCard) {
            musicCard.style.border = '2px solid rgba(102, 126, 234, 0.5)';
            musicCard.style.boxShadow = '0 0 30px rgba(102, 126, 234, 0.3)';
            musicCard.style.animation = 'gentle-pulse 2s ease-in-out infinite';
        }
    }

    function stopPlayingAnimation() {
        const musicCard = document.querySelector('.main-thread .music-card');
        if (musicCard) {
            musicCard.style.border = '';
            musicCard.style.boxShadow = '';
            musicCard.style.animation = '';
        }
    }
    
    // Gestion des commentaires
    function handleCommentInput() {
        const btn = document.querySelector('.comment-btn');
        if (commentInput.value.trim()) {
            btn.style.opacity = '1';
            btn.disabled = false;
        } else {
            btn.style.opacity = '0.5';
            btn.disabled = true;
        }
    }
    
    async function submitComment() {
        const content = commentInput.value.trim();
        if (!content) return;
        
        // Extraire l'ID du thread depuis l'URL
        const threadId = window.location.pathname.split('/thread/')[1];
        if (!threadId) {
            showNotification('❌ Erreur : ID du thread non trouvé', 'error');
            return;
        }
        
        // Désactiver le bouton pendant l'envoi
        const submitBtn = document.querySelector('.comment-btn');
        if (submitBtn) {
            submitBtn.disabled = true;
            submitBtn.textContent = 'Envoi...';
        }
        
        try {
            const response = await fetch(`/api/threads/${threadId}/messages`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'same-origin',
                body: JSON.stringify({ content: content })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.success) {
                // Créer le commentaire avec les données du serveur
                const newComment = createCommentElement({
                    id: data.data.id,
                    avatar: data.data.author.username.substring(0, 2).toUpperCase(),
                    name: data.data.author.username,
                    time: 'À l\'instant',
                    content: data.data.content,
                    isOP: false,
                    isOwn: true
                });
                
                // Ajouter au début de la liste
                commentsList.insertBefore(newComment, commentsList.firstChild);
                
                // Réinitialiser l'input
                commentInput.value = '';
                handleCommentInput();
                
                // Mettre à jour le compteur (qui va compter automatiquement)
                updateCommentCount();
                
                // Animation d'apparition
                newComment.style.opacity = '0';
                newComment.style.transform = 'translateY(-20px)';
                setTimeout(() => {
                    newComment.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                    newComment.style.opacity = '1';
                    newComment.style.transform = 'translateY(0)';
                }, 100);
                
                showNotification('💬 Commentaire publié !', 'success');
            } else {
                throw new Error(data.message || 'Erreur lors de l\'ajout du commentaire');
            }
            
        } catch (error) {
            console.error('Erreur ajout commentaire:', error);
            
            // En cas d'erreur, fallback vers POST classique
            console.log('Fallback vers POST classique...');
            const form = document.getElementById('comment-form');
            if (form) {
                form.submit();
            } else {
                showNotification('❌ Erreur lors de l\'envoi du commentaire', 'error');
            }
            
        } finally {
            // Réactiver le bouton
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Commenter';
            }
        }
    }
    
    // Créer un élément commentaire
    function createCommentElement(data) {
        const comment = document.createElement('div');
        comment.className = 'comment-item';
        
        // Ajouter l'ID du message s'il est fourni
        if (data.id) {
            comment.setAttribute('data-message-id', data.id);
        }
        
        const isOP = data.name === 'AudioSeeker';
        const opBadge = isOP ? '<span class="op-badge">OP</span>' : '';
        
        comment.innerHTML = `
                <div class="comment-avatar">
                <div class="user-pic">${data.avatar}</div>
                </div>
                <div class="comment-content">
                    <div class="comment-header">
                    <h4>${data.name}</h4>
                    <span class="comment-time">${data.time}</span>
                    ${opBadge}
                    </div>
                <div class="comment-text">${data.content}</div>
                    <div class="comment-actions">
                        <button class="comment-action">❤️ 0</button>
                    <button class="comment-action reply-btn">💬 Répondre</button>
                        <button class="comment-action">📤 Partager</button>
                    ${data.isOwn ? '<button class="comment-action delete-btn">🗑️ Supprimer</button>' : ''}
                </div>
            </div>
        `;
        
        // Attacher les événements
        attachCommentEventListeners(comment);
        
        return comment;
    }
    
    // Attacher les événements aux commentaires
    function attachCommentEventListeners(container = document) {
        // Actions de commentaires
        container.querySelectorAll('.comment-action').forEach(btn => {
            btn.addEventListener('click', handleCommentAction);
        });
        
        // Posts similaires
        container.querySelectorAll('.similar-post').forEach(post => {
            post.addEventListener('click', handleSimilarPostClick);
        });
    }
    
    // Actions des commentaires
    function handleCommentAction(e) {
        const btn = e.currentTarget;
        const action = btn.textContent.trim();
        const commentItem = btn.closest('.comment-item');
        
        // Animation
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        if (action.includes('❤️')) {
            toggleCommentLike(btn);
        } else if (action.includes('Répondre')) {
            openReplyBox(commentItem);
        } else if (action.includes('Partager')) {
            shareComment();
        } else if (action.includes('Supprimer')) {
            deleteComment(commentItem);
        }
    }
    
    // Like de commentaire
    async function toggleCommentLike(btn) {
        const commentItem = btn.closest('.comment-item');
        const commentId = commentItem.dataset.messageId;
        
        if (!commentId) {
            console.error('ID du commentaire non trouvé');
            showNotification('❌ Erreur : ID du commentaire non trouvé', 'error');
            return;
        }
        
        const isLiked = btn.classList.contains('liked');
        const currentCount = parseInt(btn.textContent.match(/\d+/)[0] || '0');
        
        // Désactiver le bouton pendant la requête
        btn.disabled = true;
        
        try {
            const response = await fetch(`/api/messages/${commentId}/like`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'same-origin'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.success) {
                // Mettre à jour l'affichage avec les vraies données du serveur
                if (data.data.is_liked) {
                    btn.classList.add('liked');
                    btn.style.color = '#ff6b6b';
                } else {
                    btn.classList.remove('liked');
                    btn.style.color = '';
                }
                
                btn.textContent = `❤️ ${data.data.likes_count}`;
                
                console.log(`Like ${data.data.is_liked ? 'ajouté' : 'supprimé'} sur commentaire ${commentId}`);
            } else {
                throw new Error(data.message || 'Erreur lors du like');
            }
            
        } catch (error) {
            console.error('Erreur like commentaire:', error);
            showNotification('❌ Erreur lors du like', 'error');
            
            // Restaurer l'état précédent en cas d'erreur
            // (ne pas changer l'affichage)
            
        } finally {
            // Réactiver le bouton
            btn.disabled = false;
        }
    }
    
    // Ouvrir la boîte de réponse
    function openReplyBox(commentItem) {
        const userName = commentItem.querySelector('.comment-header h4').textContent;
        
        // Fermer les autres boîtes de réponse
        document.querySelectorAll('.reply-box').forEach(box => box.remove());
        
        const replyBox = document.createElement('div');
        replyBox.className = 'reply-box';
        replyBox.innerHTML = `
            <div class="reply-content">
                <div class="user-pic small">MO</div>
                <div class="reply-input-area">
                    <textarea class="reply-input" placeholder="Répondre à ${userName}..."></textarea>
                    <div class="reply-actions">
                        <button class="cancel-reply">Annuler</button>
                        <button class="submit-reply">Répondre</button>
                    </div>
                </div>
            </div>
        `;
        
        // Styles inline pour la boîte de réponse
        replyBox.style.cssText = `
            margin-top: 12px;
            margin-left: 52px;
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            padding: 16px;
        `;
        
        commentItem.appendChild(replyBox);
        
        // Focus sur l'input
        const replyInput = replyBox.querySelector('.reply-input');
        replyInput.focus();
        
        // Événements
        replyBox.querySelector('.cancel-reply').addEventListener('click', () => {
            replyBox.remove();
        });
        
        replyBox.querySelector('.submit-reply').addEventListener('click', () => {
            submitReply(replyInput.value, commentItem, userName);
            replyBox.remove();
        });
        
        replyInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && e.ctrlKey) {
                submitReply(replyInput.value, commentItem, userName);
                replyBox.remove();
            }
        });
    }
    
    // Soumettre une réponse
    function submitReply(content, commentItem, userName) {
        if (!content.trim()) return;
        
        const reply = document.createElement('div');
        reply.className = 'reply-item';
        reply.innerHTML = `
            <div class="comment-avatar">
                <div class="user-pic small">MO</div>
            </div>
            <div class="comment-content">
                <div class="comment-header">
                    <h4>Moi</h4>
                    <span class="comment-time">À l'instant</span>
                </div>
                <div class="comment-text">@${userName} ${content}</div>
                <div class="comment-actions">
                    <button class="comment-action">❤️ 0</button>
                    <button class="comment-action reply-btn">💬 Répondre</button>
                </div>
            </div>
        `;
        
        commentItem.querySelector('.comment-content').appendChild(reply);
        
        // Attacher les événements
        attachCommentEventListeners(reply);
        
        showNotification(`💬 Réponse à ${userName} publiée !`, 'success');
    }
    
    // Supprimer un commentaire
    function deleteComment(commentItem) {
        if (confirm('Êtes-vous sûr de vouloir supprimer ce commentaire ?')) {
            commentItem.style.opacity = '0';
            commentItem.style.transform = 'translateY(-20px)';
            setTimeout(() => {
                commentItem.remove();
                commentCount--;
                updateCommentCount();
            }, 300);
            
            showNotification('🗑️ Commentaire supprimé', 'info');
        }
    }
    
    // Mettre à jour le compteur de commentaires
    function updateCommentCount() {
        // Compter les commentaires réels dans le DOM
        const actualCommentCount = document.querySelectorAll('.comment-item').length;
        
        const commentCountElement = document.querySelector('.engagement-btn .btn-label[textContent="Commentaires"]');
        if (commentCountElement) {
            const countElement = commentCountElement.parentElement.querySelector('.btn-count');
            countElement.textContent = actualCommentCount;
        }
        
        const commentsHeader = document.querySelector('.comments-header h3');
        if (commentsHeader) {
            commentsHeader.textContent = `Commentaires (${actualCommentCount})`;
        }
        
        // Mettre à jour la variable globale pour cohérence
        commentCount = actualCommentCount;
    }
    
    // Charger plus de commentaires
    function loadMoreComments() {
        const btn = loadMoreBtn;
        const originalText = btn.textContent;
        
        btn.textContent = 'Chargement...';
        btn.disabled = true;
        
        setTimeout(() => {
            // Simuler le chargement de 2 nouveaux commentaires
            const newComments = [
                {
                    avatar: 'VW',
                    name: 'VibeWave',
                    time: 'il y a 2h',
                    content: 'Cette track me rappelle mes premières explorations de l\'ambient. Nostalgie ! 🌊'
                },
                {
                    avatar: 'DS',
                    name: 'DeepSounds',
                    time: 'il y a 3h',
                    content: 'Album entier à écouter absolument. Chaque morceau est une pépite. 💎'
                }
            ];
            
            newComments.forEach(commentData => {
                const comment = createCommentElement(commentData);
                comment.style.opacity = '0';
                comment.style.transform = 'translateY(20px)';
                commentsList.appendChild(comment);
                
                setTimeout(() => {
                    comment.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                    comment.style.opacity = '1';
                    comment.style.transform = 'translateY(0)';
                }, 100);
            });
            
            btn.textContent = 'Charger plus de commentaires (2 restants)';
            btn.disabled = false;
            
            // Masquer après le prochain clic
            let clickCount = 0;
            const originalClickHandler = btn.onclick;
            btn.onclick = function() {
                clickCount++;
                if (clickCount >= 1) {
                    btn.style.display = 'none';
                }
                if (originalClickHandler) originalClickHandler();
            };
            
        }, 1500);
    }
    
    // Filtrer les commentaires
    function filterComments() {
        const filter = filterSelect.value;
        const comments = Array.from(commentsList.children);
        
        switch(filter) {
            case 'newest':
                // Tri par date (simulation)
                comments.sort((a, b) => {
                    const timeA = a.querySelector('.comment-time').textContent;
                    const timeB = b.querySelector('.comment-time').textContent;
                    return timeA.localeCompare(timeB);
                });
                break;
            case 'oldest':
                comments.reverse();
                break;
            case 'popular':
                // Tri par nombre de likes
                comments.sort((a, b) => {
                    const likesA = parseInt(a.querySelector('.comment-action').textContent.match(/\d+/)[0] || '0');
                    const likesB = parseInt(b.querySelector('.comment-action').textContent.match(/\d+/)[0] || '0');
                    return likesB - likesA;
                });
                break;
        }
        
        // Réorganiser les commentaires
        comments.forEach(comment => {
            commentsList.appendChild(comment);
        });
        
        showNotification(`Commentaires triés par ${filter === 'newest' ? 'plus récents' : filter === 'oldest' ? 'plus anciens' : 'popularité'}`, 'info');
    }
    
    // Actions du thread
    function handleThreadAction(e) {
        const btn = e.currentTarget;
        const title = btn.getAttribute('title');
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(title) {
            case 'Partager':
                shareThread();
                break;
            case 'Signaler':
                reportThread();
                break;
            case 'Plus':
                showThreadMenu(btn);
                break;
        }
    }
    
    // Actions de la track
    function handleTrackAction(e) {
        const btn = e.currentTarget;
        const title = btn.getAttribute('title');
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(title) {
            case 'Ajouter aux favoris':
                toggleTrackFavorite(btn);
                break;
            case 'Ajouter à une playlist':
                showPlaylistSelector();
                break;
            case 'Partager':
                shareTrack();
                break;
        }
    }
    
    // Toggle favori de track
    function toggleTrackFavorite(btn) {
        const isFavorite = btn.style.color === 'rgb(255, 107, 107)';
        
        if (isFavorite) {
            btn.style.color = '';
            showNotification('💔 Retiré des favoris', 'info');
        } else {
            btn.style.color = '#ff6b6b';
            showNotification('❤️ Ajouté aux favoris !', 'success');
        }
    }
    
    // Actions de la sidebar
    function handleSidebarAction(e) {
        const btn = e.currentTarget;
        const isFollowBtn = btn.classList.contains('follow-btn');
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        if (isFollowBtn) {
            const isFollowing = btn.textContent.includes('✓');
            if (isFollowing) {
                btn.textContent = '+ Suivre';
                btn.style.background = 'var(--primary-gradient)';
                showNotification('Vous ne suivez plus AudioSeeker', 'info');
            } else {
                btn.textContent = '✓ Suivi';
                btn.style.background = 'var(--accent-success)';
                showNotification('Vous suivez maintenant AudioSeeker !', 'success');
            }
        } else {
            // Message button
            showNotification('💬 Ouverture de la conversation...', 'info');
            setTimeout(() => {
                window.location.href = 'messages.html';
            }, 1000);
        }
    }
    
    // Clic sur les tags tendance
    function handleTrendTagClick(e) {
        e.preventDefault();
        const tag = e.currentTarget.textContent;
        showNotification(`🔍 Recherche pour ${tag}...`, 'info');
        
        setTimeout(() => {
            window.location.href = `discover.html?tag=${encodeURIComponent(tag)}`;
        }, 1000);
    }
    
    // Outils du compositeur
    function handleComposerTool(e) {
        const btn = e.currentTarget;
        const title = btn.getAttribute('title');
        
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(title) {
            case 'Ajouter une piste':
                openMusicShareModal();
                break;
            case 'Joindre une image':
                showNotification('📷 Sélection d\'image...', 'info');
                break;
            case 'Emoji':
                showEmojiPicker();
                break;
        }
    }
    
    // Modal de partage de musique
    function openMusicShareModal() {
        if (musicShareModal) {
            musicShareModal.classList.add('active');
            
            const searchInput = musicShareModal.querySelector('.music-search-input');
            if (searchInput) {
                searchInput.focus();
            }
        }
    }

    function attachModalEventListeners() {
        const closeBtn = musicShareModal.querySelector('.close-modal');
        if (closeBtn) {
            closeBtn.addEventListener('click', () => {
                musicShareModal.classList.remove('active');
            });
        }
        
        musicShareModal.addEventListener('click', (e) => {
            if (e.target === musicShareModal) {
                musicShareModal.classList.remove('active');
            }
        });
        
        const selectBtns = musicShareModal.querySelectorAll('.select-music-btn');
        selectBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                const musicTitle = btn.closest('.music-item').querySelector('h5').textContent;
                addMusicToComment(musicTitle);
                musicShareModal.classList.remove('active');
            });
        });
    }
    
    // Ajouter de la musique au commentaire
    function addMusicToComment(musicTitle) {
        const currentText = commentInput.value;
        const musicText = `\n🎵 Je recommande: "${musicTitle}"`;
        commentInput.value = currentText + musicText;
        handleCommentInput();
        showNotification(`🎵 "${musicTitle}" ajouté au commentaire !`, 'success');
    }
    
    // Simuler une réponse de l'auteur
    function simulateAuthorReply() {
        const replies = [
            "Merci pour ton commentaire ! 🙏",
            "Content que ça te plaise !",
            "Excellente question, je vais y réfléchir 🤔",
            "Tu as vraiment bon goût musical 👌",
            "N'hésite pas à partager tes propres découvertes !",
            "Merci pour cette belle interaction ! ✨"
        ];
        
        const randomReply = replies[Math.floor(Math.random() * replies.length)];
        const authorReply = createCommentElement({
            avatar: 'AS',
            name: 'AudioSeeker',
            time: 'À l\'instant',
            content: randomReply,
            isOP: true,
            isOwn: false
        });
        
        commentsList.insertBefore(authorReply, commentsList.children[1]);
        
        // Animation
        authorReply.style.opacity = '0';
        authorReply.style.transform = 'translateY(-20px)';
        setTimeout(() => {
            authorReply.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
            authorReply.style.opacity = '1';
            authorReply.style.transform = 'translateY(0)';
        }, 100);
        
        commentCount++;
        updateCommentCount();
        
        showNotification('💬 AudioSeeker a répondu !', 'info');
    }
    
    // Charger le thread depuis l'URL
    function loadThreadFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        const threadId = urlParams.get('id');
        
        if (threadId) {
            // Simuler le chargement de données spécifiques
            showNotification(`📖 Thread #${threadId} chargé`, 'info');
        }
    }
    
    // Mises à jour en temps réel
    function startRealTimeUpdates() {
        // Simuler l'arrivée de nouveaux likes
        setInterval(() => {
            if (Math.random() < 0.1) { // 10% de chance
                likeCount++;
                const likeBtn = document.querySelector('.engagement-btn.liked .btn-count');
                if (likeBtn) {
                    likeBtn.textContent = likeCount;
                    updateLikedBySection();
                }
            }
        }, 10000);
        
        // Simuler l'arrivée de nouveaux commentaires
        setInterval(() => {
            if (Math.random() < 0.05) { // 5% de chance
                simulateNewComment();
            }
        }, 30000);
    }
    
    // Simuler un nouveau commentaire
    function simulateNewComment() {
        const users = [
            { avatar: 'TR', name: 'TechnoRider' },
            { avatar: 'AM', name: 'AmbientMood' },
            { avatar: 'SC', name: 'SoundCrafter' }
        ];
        
        const comments = [
            "Découverte incroyable ! 🎧",
            "Ajouté à ma playlist immédiatement",
            "Parfait pour mes sessions de méditation 🧘",
            "Cet artiste est un génie",
            "Merci pour le partage !"
        ];
        
        const randomUser = users[Math.floor(Math.random() * users.length)];
        const randomComment = comments[Math.floor(Math.random() * comments.length)];
        
        const newComment = createCommentElement({
            avatar: randomUser.avatar,
            name: randomUser.name,
            time: 'À l\'instant',
            content: randomComment,
            isOP: false,
            isOwn: false
        });
        
        commentsList.appendChild(newComment);
        
        // Animation
        newComment.style.opacity = '0';
        newComment.style.transform = 'translateY(20px)';
        setTimeout(() => {
            newComment.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
            newComment.style.opacity = '1';
            newComment.style.transform = 'translateY(0)';
        }, 100);
        
        commentCount++;
        updateCommentCount();
        
        showNotification(`💬 Nouveau commentaire de ${randomUser.name}`, 'info');
    }
    
    // Animations d'entrée
    function animateOnLoad() {
        const elements = [
            '.main-thread',
            '.comment-composer',
            '.comments-section'
        ];
        
        elements.forEach((selector, index) => {
            const element = document.querySelector(selector);
            if (element) {
                element.style.opacity = '0';
                element.style.transform = 'translateY(30px)';
                
                setTimeout(() => {
                    element.style.transition = 'opacity 0.8s ease, transform 0.8s ease';
                    element.style.opacity = '1';
                    element.style.transform = 'translateY(0)';
                }, 200 + index * 300);
            }
        });
    }
    
    // Fonctions utilitaires
    function showNotification(message, type = 'info') {
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
        } else {
            console.log(`Notification: ${message}`);
        }
    }
    
    function reportThread() {
        showNotification('⚠️ Thread signalé aux modérateurs', 'warning');
    }
    
    function shareComment() {
        showNotification('📤 Commentaire partagé !', 'success');
    }
    
    function shareTrack() {
        showNotification('🎵 Track partagée !', 'success');
    }
    
    function showPlaylistSelector() {
        showNotification('📚 Sélection de playlist...', 'info');
    }
    
    function showEmojiPicker() {
        showNotification('😊 Sélecteur d\'emoji ouvert', 'info');
    }
    
    function showThreadMenu(btn) {
        showNotification('⋯ Menu des options ouvert', 'info');
    }
    
    function handleSimilarPostClick(e) {
        const postId = Math.floor(Math.random() * 1000);
        showNotification('📖 Chargement du post similaire...', 'info');
        setTimeout(() => {
            window.location.href = `thread.html?id=${postId}`;
        }, 1000);
    }
    
    // Raccourcis clavier
    document.addEventListener('keydown', function(e) {
        // L pour like
        if (e.key === 'l' && !e.ctrlKey && !isInputFocused()) {
            const likeBtn = document.querySelector('.engagement-btn:first-child');
            if (likeBtn) likeBtn.click();
        }
        
        // C pour focus commentaire
        if (e.key === 'c' && !e.ctrlKey && !isInputFocused()) {
            commentInput.focus();
        }
        
        // Escape pour fermer les modaux
        if (e.key === 'Escape') {
            document.querySelectorAll('.reply-box').forEach(box => box.remove());
            if (musicShareModal && musicShareModal.classList.contains('active')) {
                musicShareModal.classList.remove('active');
            }
        }
        
        // S pour partager
        if (e.key === 's' && !e.ctrlKey && !isInputFocused()) {
            shareThread();
        }
    });
    
    function isInputFocused() {
        const activeElement = document.activeElement;
        return activeElement.tagName === 'INPUT' || activeElement.tagName === 'TEXTAREA';
    }
    
    console.log('📖 Page Thread Rythm\'it initialisée avec succès !');
    console.log('💬 Fonctionnalités: Commentaires, Réponses, Likes, Partage, Lecture audio');
    console.log('⌨️ Raccourcis: L (like), C (commentaire), S (partager), Esc (fermer)');

    // Initialiser le dropdown de filtre des commentaires
    function initializeCommentsFilter() {
        const dropdown = document.getElementById('sortDropdown');
        const trigger = document.getElementById('sortTrigger');
        const menu = document.getElementById('sortMenu');
        const items = menu ? menu.querySelectorAll('.dropdown-item') : [];

        if (!dropdown || !trigger || !menu) return;

        // Ouvrir/fermer le dropdown
        trigger.addEventListener('click', function(e) {
            e.stopPropagation();
            const isActive = trigger.classList.contains('active');
            
            if (isActive) {
                closeDropdown();
            } else {
                openDropdown();
            }
        });

        // Sélectionner un élément
        items.forEach(item => {
            item.addEventListener('click', function(e) {
                e.stopPropagation();
                selectFilterOption(this);
            });
        });

        // Fermer en cliquant ailleurs
        document.addEventListener('click', function(e) {
            if (!dropdown.contains(e.target)) {
                closeDropdown();
            }
        });

        function openDropdown() {
            trigger.classList.add('active');
            menu.classList.add('active');
        }

        function closeDropdown() {
            trigger.classList.remove('active');
            menu.classList.remove('active');
        }

        function selectFilterOption(selectedItem) {
            // Retirer la classe active de tous les éléments
            items.forEach(item => item.classList.remove('active'));
            
            // Ajouter la classe active à l'élément sélectionné
            selectedItem.classList.add('active');
            
            // Mettre à jour le texte du trigger
            const selectedText = selectedItem.querySelector('.item-text').textContent;
            trigger.querySelector('.dropdown-text').textContent = selectedText;
            
            // Appliquer le filtre
            const filterValue = selectedItem.getAttribute('data-value');
            applyCommentsFilter(filterValue);
            
            // Fermer le dropdown
            closeDropdown();
            
            // Notification
            showNotification(`📋 Tri: ${selectedText}`, 'info');
        }
    }

    // Appliquer le filtre des commentaires
    function applyCommentsFilter(filterType) {
        const commentsList = document.querySelector('.comments-list');
        if (!commentsList) return;

        const comments = Array.from(commentsList.querySelectorAll('.comment-item'));
        
        console.log(`🔄 Application du filtre: ${filterType}`);

        // Trier les commentaires selon le filtre
        comments.sort((a, b) => {
            switch(filterType) {
                case 'newest':
                    // Trier par timestamp (plus récent d'abord)
                    const timestampA = parseInt(a.getAttribute('data-timestamp')) || 0;
                    const timestampB = parseInt(b.getAttribute('data-timestamp')) || 0;
                    return timestampB - timestampA;
                    
                case 'oldest':
                    // Trier par timestamp (plus ancien d'abord)
                    const timestampA2 = parseInt(a.getAttribute('data-timestamp')) || 0;
                    const timestampB2 = parseInt(b.getAttribute('data-timestamp')) || 0;
                    return timestampA2 - timestampB2;
                    
                case 'most_liked':
                    // Trier par nombre de likes (plus aimé d'abord)
                    const likesA = parseInt(a.getAttribute('data-likes')) || 0;
                    const likesB = parseInt(b.getAttribute('data-likes')) || 0;
                    return likesB - likesA;
                    
                default:
                    return 0;
            }
        });

        // Réorganiser les commentaires dans le DOM avec animation
        comments.forEach((comment, index) => {
            comment.style.opacity = '0.5';
            comment.style.transform = 'translateX(-10px)';
            
            setTimeout(() => {
                commentsList.appendChild(comment);
                comment.style.transition = 'all 0.3s ease';
                comment.style.opacity = '1';
                comment.style.transform = 'translateX(0)';
            }, index * 50);
        });

        console.log(`✅ ${comments.length} commentaires triés par ${filterType}`);
    }
});

// FONCTION GLOBALE: Afficher une notification
function showGlobalNotification(message, type = 'info') {
    // Essayer d'utiliser la fonction globale si elle existe
    if (typeof window.showNotification === 'function') {
        window.showNotification(message, type);
        return;
    }
    
    // Sinon, créer une notification simple
    console.log(`Notification [${type}]: ${message}`);
    
    // Créer un élément de notification temporaire
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: ${type === 'success' ? '#4CAF50' : type === 'error' ? '#f44336' : '#2196F3'};
        color: white;
        padding: 12px 20px;
        border-radius: 8px;
        z-index: 10000;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        font-size: 14px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        max-width: 300px;
        word-wrap: break-word;
    `;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // Supprimer après 3 secondes
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 3000);
}

// FONCTION GLOBALE: Supprimer un thread
async function deleteThread(threadId) {
    // Demander confirmation
    if (!confirm('Êtes-vous sûr de vouloir supprimer ce thread ? Cette action est irréversible.')) {
        return;
    }

    try {
        console.log(`🗑️ Tentative de suppression du thread ${threadId}`);
        
        // Afficher une notification de chargement
        showGlobalNotification('🗑️ Suppression en cours...', 'info');
        
        // Appel à l'API backend
        const response = await fetch(`/thread/${threadId}/delete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include' // Important pour envoyer les cookies d'auth
        });

        if (!response.ok) {
            throw new Error(`Erreur HTTP: ${response.status}`);
        }

        const data = await response.json();
        console.log('📊 Réponse suppression API:', data);

        if (data.success) {
            showGlobalNotification('✅ Thread supprimé avec succès !', 'success');
            
            // Rediriger vers la page d'accueil après 1.5 secondes
            setTimeout(() => {
                window.location.href = '/';
            }, 1500);
        } else {
            console.error('❌ Erreur serveur:', data.message);
            showGlobalNotification('❌ Erreur lors de la suppression: ' + (data.message || 'Erreur inconnue'), 'error');
        }

    } catch (error) {
        console.error('❌ Erreur réseau suppression:', error);
        showGlobalNotification('❌ Erreur de connexion lors de la suppression', 'error');
    }
}

// Styles CSS additionnels pour les animations
const threadAdditionalStyles = `
@keyframes heartFloat {
    0% {
        transform: translateY(0) scale(1);
        opacity: 1;
    }
    100% {
        transform: translateY(-50px) scale(1.5);
        opacity: 0;
    }
}

@keyframes gentle-pulse {
    0%, 100% {
        transform: scale(1);
        opacity: 1;
    }
    50% {
        transform: scale(1.02);
        opacity: 0.9;
    }
}

.reply-content {
    display: flex;
    gap: 12px;
    align-items: flex-start;
}

.reply-input-area {
    flex: 1;
}

.reply-input {
    width: 100%;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 12px;
    color: #f0f0f0;
    font-size: 13px;
    resize: vertical;
    min-height: 60px;
    font-family: inherit;
    margin-bottom: 8px;
}

.reply-input:focus {
    outline: none;
    border-color: rgba(102, 126, 234, 0.5);
    background: rgba(255, 255, 255, 0.08);
}

.reply-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
}

.cancel-reply, .submit-reply {
    padding: 6px 12px;
    border-radius: 6px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
}

.cancel-reply {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #f0f0f0;
}

.submit-reply {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;
}

.cancel-reply:hover, .submit-reply:hover {
    transform: translateY(-1px);
}

.music-share-modal {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(10px);
    z-index: 1000;
    align-items: center;
    justify-content: center;
}

.music-share-modal.active {
    display: flex;
}
`;

// Ajouter les styles
if (!document.getElementById('thread-additional-styles')) {
    const styleSheet = document.createElement('style');
    styleSheet.id = 'thread-additional-styles';
    styleSheet.textContent = threadAdditionalStyles;
    document.head.appendChild(styleSheet);
} 