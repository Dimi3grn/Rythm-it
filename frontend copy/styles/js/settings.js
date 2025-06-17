// JavaScript pour la page Paramètres - settings.js

document.addEventListener('DOMContentLoaded', function() {
    
    // Variables globales
    let currentSection = 'account';
    let hasUnsavedChanges = false;
    let originalSettings = {};
    
    // Éléments DOM
    const navItems = document.querySelectorAll('.settings-nav-item');
    const sections = document.querySelectorAll('.settings-section');
    const toggleSwitches = document.querySelectorAll('.toggle-switch input');
    const settingsInputs = document.querySelectorAll('.settings-input, .settings-textarea, .settings-select');
    const eqSliders = document.querySelectorAll('.eq-slider');
    const themeRadios = document.querySelectorAll('input[name="theme"]');
    const accentRadios = document.querySelectorAll('input[name="accent"]');
    
    // Initialisation
    init();
    
    function init() {
        // Charger les paramètres sauvegardés
        loadSettings();
        
        // Attacher les événements
        attachEventListeners();
        
        // Animer l'entrée
        animateOnLoad();
        
        // Afficher la section active
        showSection(currentSection);
        
        // Surveiller les changements
        startChangeTracking();
    }
    
    // Gestion des événements
    function attachEventListeners() {
        // Navigation entre sections
        navItems.forEach(item => {
            item.addEventListener('click', function(e) {
                e.preventDefault();
                const section = this.getAttribute('data-section');
                if (hasUnsavedChanges) {
                    confirmSectionChange(section);
                } else {
                    showSection(section);
                }
            });
        });
        
        // Changements dans les inputs
        settingsInputs.forEach(input => {
            input.addEventListener('input', trackChanges);
            input.addEventListener('change', trackChanges);
        });
        
        // Toggle switches
        toggleSwitches.forEach(toggle => {
            toggle.addEventListener('change', function() {
                trackChanges();
                handleToggleChange(this);
            });
        });
        
        // Sliders d'égaliseur
        eqSliders.forEach(slider => {
            slider.addEventListener('input', function() {
                updateEqualizer();
                trackChanges();
            });
        });
        
        // Changements de thème
        themeRadios.forEach(radio => {
            radio.addEventListener('change', function() {
                if (this.checked) {
                    applyTheme(this.value);
                    trackChanges();
                }
            });
        });
        
        // Changements de couleur d'accent
        accentRadios.forEach(radio => {
            radio.addEventListener('change', function() {
                if (this.checked) {
                    applyAccentColor(this.value);
                    trackChanges();
                }
            });
        });
        
        // Boutons d'action
        document.querySelectorAll('.save-btn').forEach(btn => {
            btn.addEventListener('click', saveSettings);
        });
        
        document.querySelectorAll('.action-btn').forEach(btn => {
            btn.addEventListener('click', handleActionButton);
        });
        
        document.querySelectorAll('.danger-btn').forEach(btn => {
            btn.addEventListener('click', handleDangerAction);
        });
        
        // Support buttons
        document.querySelectorAll('.support-btn').forEach(btn => {
            btn.addEventListener('click', handleSupportAction);
        });
        
        // Gestion des raccourcis clavier
        document.addEventListener('keydown', handleKeyboardShortcuts);
        
        // Prévenir la fermeture avec des changements non sauvegardés
        window.addEventListener('beforeunload', function(e) {
            if (hasUnsavedChanges) {
                e.preventDefault();
                e.returnValue = '';
            }
        });
    }
    
    // Affichage des sections
    function showSection(sectionId) {
        // Mettre à jour la navigation
        navItems.forEach(item => {
            item.classList.remove('active');
            if (item.getAttribute('data-section') === sectionId) {
                item.classList.add('active');
            }
        });
        
        // Mettre à jour les sections
        sections.forEach(section => {
            section.classList.remove('active');
            if (section.id === sectionId) {
                section.classList.add('active');
                
                // Animation d'entrée
                section.style.opacity = '0';
                section.style.transform = 'translateY(20px)';
                setTimeout(() => {
                    section.style.transition = 'opacity 0.5s ease, transform 0.5s ease';
                    section.style.opacity = '1';
                    section.style.transform = 'translateY(0)';
                }, 50);
            }
        });
        
        currentSection = sectionId;
        
        // Charger le contenu spécifique à la section
        loadSectionContent(sectionId);
    }
    
    // Confirmation de changement de section
    function confirmSectionChange(newSection) {
        const modal = createConfirmModal(
            'Changements non sauvegardés',
            'Vous avez des modifications non sauvegardées. Que souhaitez-vous faire ?',
            [
                {
                    text: 'Sauvegarder',
                    action: () => {
                        saveSettings();
                        showSection(newSection);
                    },
                    primary: true
                },
                {
                    text: 'Ignorer',
                    action: () => {
                        discardChanges();
                        showSection(newSection);
                    },
                    danger: true
                },
                {
                    text: 'Annuler',
                    action: () => {}
                }
            ]
        );
        
        document.body.appendChild(modal);
    }
    
    // Chargement du contenu spécifique aux sections
    function loadSectionContent(sectionId) {
        switch(sectionId) {
            case 'account':
                loadAccountSettings();
                break;
            case 'notifications':
                updateNotificationCount();
                break;
            case 'audio':
                initializeEqualizer();
                break;
            case 'storage':
                updateStorageInfo();
                break;
            case 'about':
                checkForUpdates();
                break;
        }
    }
    
    // Gestion des toggles
    function handleToggleChange(toggle) {
        const settingName = getSettingName(toggle);
        
        // Animation de feedback
        const slider = toggle.nextElementSibling;
        slider.style.transform = 'scale(1.05)';
        setTimeout(() => {
            slider.style.transform = 'scale(1)';
        }, 150);
        
        // Actions spécifiques selon le réglage
        switch(settingName) {
            case 'two-factor':
                if (toggle.checked) {
                    showNotification('🔐 Authentification à deux facteurs activée', 'success');
                } else {
                    showNotification('⚠️ Authentification à deux facteurs désactivée', 'warning');
                }
                break;
            case 'public-profile':
                if (toggle.checked) {
                    showNotification('🌍 Profil rendu public', 'info');
                } else {
                    showNotification('🔒 Profil rendu privé', 'info');
                }
                break;
            case 'listening-activity':
                if (toggle.checked) {
                    showNotification('🎵 Activité d\'écoute visible', 'info');
                } else {
                    showNotification('👻 Activité d\'écoute masquée', 'info');
                }
                break;
        }
    }
    
    // Gestion des boutons d'action
    function handleActionButton(e) {
        const btn = e.currentTarget;
        const action = btn.textContent.trim();
        
        // Animation de feedback
        btn.style.transform = 'scale(0.95)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(action) {
            case 'Modifier':
                openPasswordChangeModal();
                break;
            case 'Voir les sessions':
                openActiveSessionsModal();
                break;
            case 'Vider le cache':
                clearCache();
                break;
            case 'Gérer les téléchargements':
                openDownloadsManager();
                break;
            case 'Gérer les utilisateurs bloqués':
                openBlockedUsersModal();
                break;
            case 'Rechercher des mises à jour':
                checkForUpdates();
                break;
            case 'Notes de version':
                openReleaseNotes();
                break;
            case 'Réinitialiser les paramètres':
                confirmResetSettings();
                break;
            case 'Exporter les logs':
                exportLogs();
                break;
        }
    }
    
    // Gestion des actions dangereuses
    function handleDangerAction(e) {
        const btn = e.currentTarget;
        const action = btn.textContent.trim();
        
        switch(action) {
            case 'Télécharger':
                downloadUserData();
                break;
            case 'Supprimer':
                confirmAccountDeletion();
                break;
            case 'Réinitialiser toutes les données':
                confirmDataReset();
                break;
        }
    }
    
    // Gestion du support
    function handleSupportAction(e) {
        const btn = e.currentTarget;
        const action = btn.querySelector('span:last-child').textContent;
        
        // Animation
        btn.style.transform = 'scale(0.98)';
        setTimeout(() => {
            btn.style.transform = 'scale(1)';
        }, 150);
        
        switch(action) {
            case 'Centre d\'aide':
                showNotification('📚 Ouverture du centre d\'aide...', 'info');
                break;
            case 'Contacter le support':
                openSupportChat();
                break;
            case 'Signaler un bug':
                openBugReport();
                break;
            case 'Suggérer une fonctionnalité':
                openFeatureRequest();
                break;
        }
    }
    
    // Égaliseur
    function initializeEqualizer() {
        updateEqualizer();
    }
    
    function updateEqualizer() {
        const values = Array.from(eqSliders).map(slider => slider.value);
        
        // Visualisation des fréquences
        eqSliders.forEach((slider, index) => {
            const value = parseInt(slider.value);
            const color = value >= 0 ? 
                `rgba(102, 126, 234, ${Math.abs(value) / 12})` : 
                `rgba(255, 107, 107, ${Math.abs(value) / 12})`;
            
            slider.style.background = `linear-gradient(to top, ${color}, rgba(255, 255, 255, 0.1))`;
        });
        
        showNotification('🎚️ Égaliseur mis à jour', 'info');
    }
    
    // Thème
    function applyTheme(theme) {
        const root = document.documentElement;
        
        switch(theme) {
            case 'light':
                showNotification('☀️ Thème clair appliqué', 'info');
                // Ici on pourrait modifier les variables CSS
                break;
            case 'dark':
                showNotification('🌙 Thème sombre appliqué', 'info');
                break;
            case 'auto':
                showNotification('🌗 Thème automatique activé', 'info');
                break;
        }
    }
    
    // Couleur d'accent
    function applyAccentColor(color) {
        const root = document.documentElement;
        
        const colors = {
            purple: ['#667eea', '#764ba2'],
            blue: ['#3b82f6', '#1d4ed8'],
            green: ['#10b981', '#059669'],
            orange: ['#f59e0b', '#d97706'],
            red: ['#ef4444', '#dc2626'],
            pink: ['#ec4899', '#be185d']
        };
        
        if (colors[color]) {
            // Ici on pourrait mettre à jour les variables CSS
            showNotification(`🎨 Couleur d'accent ${color} appliquée`, 'success');
        }
    }
    
    // Sauvegarde des paramètres
    function saveSettings() {
        // Animation du bouton de sauvegarde
        const saveBtn = document.querySelector('.save-btn');
        if (saveBtn) {
            const originalText = saveBtn.textContent;
            saveBtn.textContent = '💾 Sauvegarde...';
            saveBtn.disabled = true;
            
            setTimeout(() => {
                saveBtn.textContent = '✅ Sauvegardé !';
                setTimeout(() => {
                    saveBtn.textContent = originalText;
                    saveBtn.disabled = false;
                }, 1000);
            }, 1000);
        }
        
        // Collecter tous les paramètres
        const settings = collectAllSettings();
        
        // Sauvegarder (simulation)
        try {
            localStorage.setItem('rhythmit_settings', JSON.stringify(settings));
            hasUnsavedChanges = false;
            originalSettings = { ...settings };
            showNotification('✅ Paramètres sauvegardés avec succès !', 'success');
        } catch (error) {
            showNotification('❌ Erreur lors de la sauvegarde', 'error');
        }
    }
    
    // Collecte de tous les paramètres
    function collectAllSettings() {
        const settings = {};
        
        // Inputs texte
        settingsInputs.forEach(input => {
            const name = getSettingName(input);
            settings[name] = input.value;
        });
        
        // Toggles
        toggleSwitches.forEach(toggle => {
            const name = getSettingName(toggle);
            settings[name] = toggle.checked;
        });
        
        // Égaliseur
        settings.equalizer = Array.from(eqSliders).map(slider => slider.value);
        
        // Thème
        const checkedTheme = document.querySelector('input[name="theme"]:checked');
        if (checkedTheme) {
            settings.theme = checkedTheme.value;
        }
        
        // Couleur d'accent
        const checkedAccent = document.querySelector('input[name="accent"]:checked');
        if (checkedAccent) {
            settings.accentColor = checkedAccent.value;
        }
        
        return settings;
    }
    
    // Chargement des paramètres
    function loadSettings() {
        try {
            const saved = localStorage.getItem('rhythmit_settings');
            if (saved) {
                const settings = JSON.parse(saved);
                applySettings(settings);
                originalSettings = { ...settings };
            }
        } catch (error) {
            console.error('Erreur lors du chargement des paramètres:', error);
        }
    }
    
    // Application des paramètres
    function applySettings(settings) {
        // Appliquer les valeurs aux inputs
        Object.keys(settings).forEach(key => {
            const element = document.querySelector(`[data-setting="${key}"]`);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = settings[key];
                } else {
                    element.value = settings[key];
                }
            }
        });
        
        // Appliquer l'égaliseur
        if (settings.equalizer) {
            settings.equalizer.forEach((value, index) => {
                if (eqSliders[index]) {
                    eqSliders[index].value = value;
                }
            });
            updateEqualizer();
        }
        
        // Appliquer le thème
        if (settings.theme) {
            const themeRadio = document.querySelector(`input[name="theme"][value="${settings.theme}"]`);
            if (themeRadio) {
                themeRadio.checked = true;
                applyTheme(settings.theme);
            }
        }
        
        // Appliquer la couleur d'accent
        if (settings.accentColor) {
            const accentRadio = document.querySelector(`input[name="accent"][value="${settings.accentColor}"]`);
            if (accentRadio) {
                accentRadio.checked = true;
                applyAccentColor(settings.accentColor);
            }
        }
    }
    
    // Suivi des changements
    function startChangeTracking() {
        const trackableElements = [
            ...settingsInputs,
            ...toggleSwitches,
            ...eqSliders,
            ...themeRadios,
            ...accentRadios
        ];
        
        trackableElements.forEach(element => {
            element.addEventListener('change', trackChanges);
            element.addEventListener('input', trackChanges);
        });
    }
    
    function trackChanges() {
        hasUnsavedChanges = true;
        
        // Afficher un indicateur visuel
        const indicator = document.querySelector('.unsaved-indicator');
        if (!indicator) {
            const badge = document.createElement('div');
            badge.className = 'unsaved-indicator';
            badge.textContent = '●';
            badge.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                background: #ff6b6b;
                color: white;
                width: 12px;
                height: 12px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 10000;
                animation: pulse 2s infinite;
            `;
            document.body.appendChild(badge);
        }
    }
    
    function discardChanges() {
        applySettings(originalSettings);
        hasUnsavedChanges = false;
        
        const indicator = document.querySelector('.unsaved-indicator');
        if (indicator) {
            indicator.remove();
        }
        
        showNotification('↶ Modifications annulées', 'info');
    }
    
    // Actions spécifiques
    function clearCache() {
        showNotification('🧹 Nettoyage du cache...', 'info');
        
        setTimeout(() => {
            // Simuler le nettoyage
            const cacheSize = '0.3 GB';
            showNotification(`✅ Cache vidé - ${cacheSize} libérés`, 'success');
            updateStorageInfo();
        }, 2000);
    }
    
    function updateStorageInfo() {
        // Simuler la mise à jour des informations de stockage
        const storageBar = document.querySelector('.storage-used');
        if (storageBar) {
            const currentWidth = parseInt(storageBar.style.width) || 65;
            const newWidth = Math.max(currentWidth - 5, 45);
            storageBar.style.width = newWidth + '%';
        }
    }
    
    function updateNotificationCount() {
        // Simuler la mise à jour du nombre de notifications
        const badge = document.querySelector('.notification-badge');
        if (badge) {
            const count = Math.floor(Math.random() * 5);
            badge.textContent = count;
            if (count === 0) {
                badge.style.display = 'none';
            }
        }
    }
    
    function checkForUpdates() {
        showNotification('🔍 Recherche de mises à jour...', 'info');
        
        setTimeout(() => {
            const hasUpdate = Math.random() < 0.3; // 30% de chance
            if (hasUpdate) {
                showNotification('🆕 Nouvelle version disponible !', 'success');
            } else {
                showNotification('✅ Vous avez la dernière version', 'success');
            }
        }, 2000);
    }
    
    function downloadUserData() {
        showNotification('📦 Préparation de vos données...', 'info');
        
        setTimeout(() => {
            // Simuler le téléchargement
            const blob = new Blob(['{"user": "data"}'], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'rhythmit-data.json';
            a.click();
            URL.revokeObjectURL(url);
            
            showNotification('📥 Données téléchargées !', 'success');
        }, 3000);
    }
    
    // Modaux de confirmation
    function confirmAccountDeletion() {
        const modal = createConfirmModal(
            'Supprimer le compte',
            'Cette action est irréversible. Toutes vos données seront définitivement supprimées.',
            [
                {
                    text: 'Supprimer définitivement',
                    action: () => {
                        showNotification('🗑️ Suppression du compte en cours...', 'warning');
                        // Simulation de suppression
                        setTimeout(() => {
                            showNotification('👋 Compte supprimé. Au revoir !', 'info');
                        }, 3000);
                    },
                    danger: true
                },
                {
                    text: 'Annuler',
                    action: () => {}
                }
            ]
        );
        
        document.body.appendChild(modal);
    }
    
    function confirmDataReset() {
        const modal = createConfirmModal(
            'Réinitialiser toutes les données',
            'Ceci supprimera toutes vos playlists, favoris et paramètres personnalisés.',
            [
                {
                    text: 'Réinitialiser',
                    action: () => {
                        localStorage.clear();
                        showNotification('🔄 Données réinitialisées', 'info');
                        setTimeout(() => {
                            location.reload();
                        }, 2000);
                    },
                    danger: true
                },
                {
                    text: 'Annuler',
                    action: () => {}
                }
            ]
        );
        
        document.body.appendChild(modal);
    }
    
    function confirmResetSettings() {
        const modal = createConfirmModal(
            'Réinitialiser les paramètres',
            'Tous vos paramètres reviendront aux valeurs par défaut.',
            [
                {
                    text: 'Réinitialiser',
                    action: () => {
                        localStorage.removeItem('rhythmit_settings');
                        showNotification('🔄 Paramètres réinitialisés', 'info');
                        setTimeout(() => {
                            location.reload();
                        }, 1500);
                    },
                    primary: true
                },
                {
                    text: 'Annuler',
                    action: () => {}
                }
            ]
        );
        
        document.body.appendChild(modal);
    }
    
    // Fonctions d'ouverture de modaux
    function openPasswordChangeModal() {
        showNotification('🔐 Ouverture du changement de mot de passe...', 'info');
    }
    
    function openActiveSessionsModal() {
        showNotification('📱 Chargement des sessions actives...', 'info');
    }
    
    function openDownloadsManager() {
        showNotification('📥 Ouverture du gestionnaire de téléchargements...', 'info');
    }
    
    function openBlockedUsersModal() {
        showNotification('🚫 Chargement des utilisateurs bloqués...', 'info');
    }
    
    function openSupportChat() {
        showNotification('💬 Connexion au support...', 'info');
    }
    
    function openBugReport() {
        showNotification('🐛 Ouverture du formulaire de rapport de bug...', 'info');
    }
    
    function openFeatureRequest() {
        showNotification('💡 Ouverture du formulaire de suggestion...', 'info');
    }
    
    function openReleaseNotes() {
        showNotification('📋 Chargement des notes de version...', 'info');
    }
    
    function exportLogs() {
        showNotification('📊 Export des logs en cours...', 'info');
        
        setTimeout(() => {
            const logs = 'Debug logs content...';
            const blob = new Blob([logs], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'rhythmit-logs.txt';
            a.click();
            URL.revokeObjectURL(url);
            
            showNotification('📤 Logs exportés !', 'success');
        }, 2000);
    }
    
    // Utilitaires
    function getSettingName(element) {
        return element.getAttribute('data-setting') || 
               element.closest('[data-setting]')?.getAttribute('data-setting') ||
               element.id || 
               element.name || 
               'unknown';
    }
    
    function createConfirmModal(title, message, buttons) {
        const modal = document.createElement('div');
        modal.className = 'confirm-modal';
        modal.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(10px);
            z-index: 10000;
            display: flex;
            align-items: center;
            justify-content: center;
        `;
        
        const content = document.createElement('div');
        content.style.cssText = `
            background: rgba(26, 26, 46, 0.95);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            padding: 30px;
            max-width: 500px;
            width: 90%;
            text-align: center;
        `;
        
        content.innerHTML = `
            <h3 style="color: #fff; margin-bottom: 15px; font-size: 20px;">${title}</h3>
            <p style="color: #888; margin-bottom: 25px; line-height: 1.5;">${message}</p>
            <div class="modal-buttons" style="display: flex; gap: 15px; justify-content: center;"></div>
        `;
        
        const buttonContainer = content.querySelector('.modal-buttons');
        
        buttons.forEach(btn => {
            const button = document.createElement('button');
            button.textContent = btn.text;
            button.style.cssText = `
                padding: 12px 20px;
                border-radius: 12px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.2s ease;
                border: none;
                ${btn.primary ? 
                    'background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white;' :
                    btn.danger ?
                    'background: linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%); color: white;' :
                    'background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); color: #f0f0f0;'
                }
            `;
            
            button.addEventListener('click', () => {
                btn.action();
                modal.remove();
            });
            
            buttonContainer.appendChild(button);
        });
        
        modal.appendChild(content);
        
        // Fermer en cliquant à l'extérieur
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.remove();
            }
        });
        
        return modal;
    }
    
    // Animations
    function animateOnLoad() {
        const cards = document.querySelectorAll('.settings-card');
        cards.forEach((card, index) => {
            card.style.opacity = '0';
            card.style.transform = 'translateY(20px)';
            
            setTimeout(() => {
                card.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
                card.style.opacity = '1';
                card.style.transform = 'translateY(0)';
            }, 100 + index * 100);
        });
    }
    
    // Raccourcis clavier
    function handleKeyboardShortcuts(e) {
        // Ctrl + S pour sauvegarder
        if (e.ctrlKey && e.key === 's') {
            e.preventDefault();
            saveSettings();
        }
        
        // Échap pour annuler les changements
        if (e.key === 'Escape' && hasUnsavedChanges) {
            if (confirm('Annuler les modifications non sauvegardées ?')) {
                discardChanges();
            }
        }
        
        // Ctrl + R pour réinitialiser
        if (e.ctrlKey && e.shiftKey && e.key === 'R') {
            e.preventDefault();
            confirmResetSettings();
        }
    }
    
    // Fonction de notification
    function showNotification(message, type = 'info') {
        if (typeof window.showNotification === 'function') {
            window.showNotification(message, type);
        } else {
            console.log(`Notification: ${message}`);
        }
    }
    
    // Gestion de la section account spécifique
    function loadAccountSettings() {
        // Simuler le chargement des informations de compte
        showNotification('👤 Chargement des informations du compte...', 'info');
    }
    
    console.log('⚙️ Page Paramètres Rythm\'it initialisée avec succès !');
    console.log('🎛️ Fonctionnalités: Gestion complète des paramètres, Thèmes, Égaliseur, Sécurité');
});