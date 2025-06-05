document.addEventListener('DOMContentLoaded', function() {
            // Focus automatique sur le premier champ
            setTimeout(() => {
                document.getElementById('email').focus();
            }, 500);
        });

        // Toggle du mot de passe
        function togglePassword() {
            const passwordInput = document.getElementById('password');
            const toggleBtn = document.querySelector('.password-toggle');
            
            if (passwordInput.type === 'password') {
                passwordInput.type = 'text';
                toggleBtn.textContent = '🙈';
            } else {
                passwordInput.type = 'password';
                toggleBtn.textContent = '👁️';
            }
        }

        // Gestion du formulaire
        document.getElementById('signinForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            const rememberMe = document.getElementById('rememberMe').checked;
            
            // Validation basique
            if (!email || !password) {
                showError('Veuillez remplir tous les champs');
                return;
            }
            
            if (!isValidEmail(email)) {
                showError('Veuillez entrer une adresse email valide');
                return;
            }
            
            // Simulation de connexion
            loginUser(email, password, rememberMe);
        });

        function loginUser(email, password, rememberMe) {
            const submitBtn = document.querySelector('.auth-button');
            const originalText = submitBtn.textContent;
            
            // Animation de chargement
            submitBtn.classList.add('loading');
            submitBtn.textContent = '';
            
            // Simulation d'appel API
            setTimeout(() => {
                // Simulation : succès si email contient "test"
                if (email.includes('test') || email === 'demo@rhythmit.com') {
                    showSuccess('Connexion réussie ! Redirection...');
                    
                    // Sauvegarder en localStorage si demandé
                    if (rememberMe) {
                        localStorage.setItem('rhythmit_remember', 'true');
                        localStorage.setItem('rhythmit_email', email);
                    }
                    
                    // Redirection vers l'accueil
                    setTimeout(() => {
                        window.location.href = 'index.html';
                    }, 1500);
                } else {
                    showError('Email ou mot de passe incorrect');
                    submitBtn.classList.remove('loading');
                    submitBtn.textContent = originalText;
                }
            }, 2000);
        }

        // Connexion avec Google
        function loginWithGoogle() {
            showSuccess('Redirection vers Google...');
            setTimeout(() => {
                // Simulation de connexion réussie
                window.location.href = 'index.html';
            }, 1500);
        }

        // Connexion avec Spotify
        function loginWithSpotify() {
            showSuccess('Redirection vers Spotify...');
            setTimeout(() => {
                // Simulation de connexion réussie
                window.location.href = 'index.html';
            }, 1500);
        }

        // Utilitaires
        function isValidEmail(email) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return emailRegex.test(email);
        }

        function showError(message) {
            const errorDiv = document.getElementById('errorMessage');
            const successDiv = document.getElementById('successMessage');
            
            successDiv.style.display = 'none';
            errorDiv.textContent = message;
            errorDiv.style.display = 'block';
            
            // Animation d'entrée
            errorDiv.style.opacity = '0';
            setTimeout(() => {
                errorDiv.style.transition = 'opacity 0.3s ease';
                errorDiv.style.opacity = '1';
            }, 10);
        }

        function showSuccess(message) {
            const errorDiv = document.getElementById('errorMessage');
            const successDiv = document.getElementById('successMessage');
            
            errorDiv.style.display = 'none';
            successDiv.textContent = message;
            successDiv.style.display = 'block';
            
            // Animation d'entrée
            successDiv.style.opacity = '0';
            setTimeout(() => {
                successDiv.style.transition = 'opacity 0.3s ease';
                successDiv.style.opacity = '1';
            }, 10);
        }

        // Auto-remplissage si "se souvenir" était activé
        window.addEventListener('load', function() {
            if (localStorage.getItem('rhythmit_remember') === 'true') {
                const savedEmail = localStorage.getItem('rhythmit_email');
                if (savedEmail) {
                    document.getElementById('email').value = savedEmail;
                    document.getElementById('rememberMe').checked = true;
                    document.getElementById('password').focus();
                }
            }
        });

        // Raccourcis clavier
        document.addEventListener('keydown', function(e) {
            // Entrée pour soumettre le formulaire
            if (e.key === 'Enter' && e.target.tagName !== 'BUTTON') {
                document.getElementById('signinForm').dispatchEvent(new Event('submit'));
            }
        });

        console.log('🎵 Page de connexion Rythm\'it chargée !');
        console.log('💡 Test: utilisez "demo@rhythmit.com" pour une connexion réussie');