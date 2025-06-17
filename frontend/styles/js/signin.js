document.addEventListener('DOMContentLoaded', function() {
    // Focus automatique sur le premier champ
    setTimeout(() => {
        const emailField = document.getElementById('email');
        if (emailField) {
            emailField.focus();
        }
    }, 500);
});

// Toggle du mot de passe
function togglePassword() {
    const passwordInput = document.getElementById('password');
    const toggleBtn = passwordInput.parentElement.querySelector('.password-toggle');
    
    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggleBtn.textContent = '🙈';
    } else {
        passwordInput.type = 'password';
        toggleBtn.textContent = '👁️';
    }
}

// Toggle pour confirmer le mot de passe (mode inscription)
function toggleConfirmPassword() {
    const confirmPasswordInput = document.getElementById('confirmPassword');
    if (confirmPasswordInput) {
        const toggleBtn = confirmPasswordInput.parentElement.querySelector('.password-toggle');
        
        if (confirmPasswordInput.type === 'password') {
            confirmPasswordInput.type = 'text';
            toggleBtn.textContent = '🙈';
        } else {
            confirmPasswordInput.type = 'password';
            toggleBtn.textContent = '👁️';
        }
    }
}

// Gestion du formulaire avec soumission réelle
document.getElementById('signinForm').addEventListener('submit', function(e) {
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const submitBtn = document.querySelector('.auth-button');
    
    // Validation basique côté client
    if (!email || !password) {
        e.preventDefault();
        showError('Veuillez remplir tous les champs');
        return;
    }
    
    if (!isValidEmail(email)) {
        e.preventDefault();
        showError('Veuillez entrer une adresse email valide');
        return;
    }

    // Afficher un indicateur de chargement
    const originalText = submitBtn.textContent;
    submitBtn.classList.add('loading');
    submitBtn.textContent = 'Connexion...';
    submitBtn.disabled = true;

    // Laisser la soumission normale du formulaire continuer
    // Le serveur va gérer la connexion et rediriger
});

// Connexion avec Google (placeholder)
function loginWithGoogle() {
    showInfo('Fonctionnalité bientôt disponible');
}

// Connexion avec Spotify (placeholder)
function loginWithSpotify() {
    showInfo('Fonctionnalité bientôt disponible');
}

// Utilitaires
function isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
}

function showError(message) {
    // Chercher les divs d'erreur existantes ou en créer une
    let errorDiv = document.querySelector('.error-message');
    
    if (!errorDiv) {
        errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        
        // Insérer avant le formulaire
        const form = document.getElementById('signinForm');
        form.parentNode.insertBefore(errorDiv, form);
    }
    
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
    
    // Masquer les messages de succès
    const successDiv = document.querySelector('.success-message');
    if (successDiv) {
        successDiv.style.display = 'none';
    }
    
    // Animation d'entrée
    errorDiv.style.opacity = '0';
    setTimeout(() => {
        errorDiv.style.transition = 'opacity 0.3s ease';
        errorDiv.style.opacity = '1';
    }, 10);
}

function showSuccess(message) {
    // Chercher les divs de succès existantes ou en créer une
    let successDiv = document.querySelector('.success-message');
    
    if (!successDiv) {
        successDiv = document.createElement('div');
        successDiv.className = 'success-message';
        
        // Insérer avant le formulaire
        const form = document.getElementById('signinForm');
        form.parentNode.insertBefore(successDiv, form);
    }
    
    successDiv.textContent = message;
    successDiv.style.display = 'block';
    
    // Masquer les messages d'erreur
    const errorDiv = document.querySelector('.error-message');
    if (errorDiv) {
        errorDiv.style.display = 'none';
    }
    
    // Animation d'entrée
    successDiv.style.opacity = '0';
    setTimeout(() => {
        successDiv.style.transition = 'opacity 0.3s ease';
        successDiv.style.opacity = '1';
    }, 10);
}

function showInfo(message) {
    // Messages informatifs
    let infoDiv = document.querySelector('.info-message');
    
    if (!infoDiv) {
        infoDiv = document.createElement('div');
        infoDiv.className = 'info-message';
        infoDiv.style.cssText = `
            background: linear-gradient(135deg, #3498db, #2980b9);
            color: white;
            padding: 12px 20px;
            border-radius: 8px;
            margin-bottom: 20px;
            text-align: center;
            box-shadow: 0 4px 12px rgba(52, 152, 219, 0.3);
        `;
        
        // Insérer avant le formulaire
        const form = document.getElementById('signinForm');
        form.parentNode.insertBefore(infoDiv, form);
    }
    
    infoDiv.textContent = message;
    infoDiv.style.display = 'block';
    
    // Masquer automatiquement après 3 secondes
    setTimeout(() => {
        infoDiv.style.display = 'none';
    }, 3000);
}

// Auto-remplissage si "se souvenir" était activé
window.addEventListener('load', function() {
    if (localStorage.getItem('rhythmit_remember') === 'true') {
        const savedEmail = localStorage.getItem('rhythmit_email');
        if (savedEmail) {
            const emailField = document.getElementById('email');
            const rememberField = document.getElementById('rememberMe');
            const passwordField = document.getElementById('password');
            
            if (emailField) emailField.value = savedEmail;
            if (rememberField) rememberField.checked = true;
            if (passwordField) passwordField.focus();
        }
    }
});

// Raccourcis clavier
document.addEventListener('keydown', function(e) {
    // Entrée pour soumettre le formulaire
    if (e.key === 'Enter' && e.target.tagName !== 'BUTTON' && e.target.type !== 'submit') {
        const form = document.getElementById('signinForm');
        if (form) {
            form.submit();
        }
    }
});

console.log('🎵 Page de connexion Rythm\'it chargée !');
console.log('💡 Utilisez admin@rythmit.com / admin123 pour tester la connexion');