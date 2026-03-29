document.addEventListener('DOMContentLoaded', function() {
    console.log('🎵 Page de connexion Rythm\'it chargée !');
    console.log('🔍 Vérification des éléments DOM...');
    
    // Debug: vérifier les éléments présents
    const passwordField = document.getElementById('password');
    const confirmPasswordField = document.getElementById('confirmPassword');
    const passwordValidation = document.getElementById('passwordValidation');
    const passwordMatch = document.getElementById('passwordMatch');
    
    console.log('🔍 Éléments trouvés:');
    console.log('  - passwordField:', passwordField);
    console.log('  - confirmPasswordField:', confirmPasswordField);
    console.log('  - passwordValidation:', passwordValidation);
    console.log('  - passwordMatch:', passwordMatch);
    
    // Focus automatique sur le premier champ
    setTimeout(() => {
        const emailField = document.getElementById('email');
        if (emailField) {
            emailField.focus();
        }
    }, 500);
    
    // Initialiser la validation des mots de passe si on est en mode inscription
    initPasswordValidation();
});

// Initialisation de la validation des mots de passe
function initPasswordValidation() {
    console.log('🔧 Initialisation de la validation des mots de passe...');
    
    const passwordField = document.getElementById('password');
    const confirmPasswordField = document.getElementById('confirmPassword');
    const passwordValidation = document.getElementById('passwordValidation');
    const passwordMatch = document.getElementById('passwordMatch');
    
    console.log('🔍 Éléments dans initPasswordValidation:');
    console.log('  - passwordField:', passwordField);
    console.log('  - confirmPasswordField:', confirmPasswordField);
    console.log('  - passwordValidation:', passwordValidation);
    console.log('  - passwordMatch:', passwordMatch);
    
    if (!passwordField) {
        console.log('❌ passwordField introuvable, arrêt de l\'initialisation');
        return;
    }
    
    if (!passwordValidation) {
        console.log('❌ passwordValidation introuvable, arrêt de l\'initialisation');
        return;
    }
    
    console.log('✅ Éléments trouvés, installation des event listeners...');
    
    // Validation en temps réel du mot de passe
    passwordField.addEventListener('input', function() {
        const password = this.value;
        console.log('⌨️ Saisie mot de passe, longueur:', password.length);
        validatePassword(password);
        
        // Valider aussi la confirmation si elle existe
        if (confirmPasswordField && confirmPasswordField.value) {
            validatePasswordMatch(password, confirmPasswordField.value);
        }
    });
    
    // Validation de la confirmation du mot de passe
    if (confirmPasswordField) {
        confirmPasswordField.addEventListener('input', function() {
            const password = passwordField.value;
            const confirmPassword = this.value;
            console.log('⌨️ Saisie confirmation mot de passe');
            validatePasswordMatch(password, confirmPassword);
        });
    }
    
    console.log('✅ Validation des mots de passe initialisée');
}

// Validation du mot de passe
function validatePassword(password) {
    console.log('🔍 Validation du mot de passe:', password);
    
    const passwordValidation = document.getElementById('passwordValidation');
    const passwordMatch = document.getElementById('passwordMatch');
    
    // Si le mot de passe est vide, cacher les divs de validation
    if (password.length === 0) {
        if (passwordValidation) {
            passwordValidation.classList.remove('show');
        }
        if (passwordMatch) {
            passwordMatch.classList.remove('show');
        }
        resetAllRules();
        updatePasswordStrength(0);
        return;
    }
    
    // Afficher les divs de validation si le mot de passe n'est pas vide
    if (passwordValidation) {
        passwordValidation.classList.add('show');
    }
    if (passwordMatch) {
        passwordMatch.classList.add('show');
    }
    
    const rules = {
        length: password.length >= 8,
        uppercase: /[A-Z]/.test(password),
        lowercase: /[a-z]/.test(password),
        number: /[0-9]/.test(password),
        special: /[!@#$%^&*(),.?":{}|<>]/.test(password)
    };
    
    console.log('📋 Règles validées:', rules);
    
    // Mettre à jour les règles visuellement
    updateRule('lengthRule', rules.length);
    updateRule('uppercaseRule', rules.uppercase);
    updateRule('lowercaseRule', rules.lowercase);
    updateRule('numberRule', rules.number);
    updateRule('specialRule', rules.special);
    
    // Calculer la force du mot de passe
    const strength = calculatePasswordStrength(password, rules);
    console.log('💪 Force du mot de passe:', strength);
    updatePasswordStrength(strength);
    
    // Mettre à jour l'apparence du champ
    const passwordField = document.getElementById('password');
    passwordField.classList.remove('password-weak', 'password-strong');
    
    if (password.length > 0) {
        if (strength >= 4) {
            passwordField.classList.add('password-strong');
        } else {
            passwordField.classList.add('password-weak');
        }
    }
}

// Reset toutes les règles à l'état neutre
function resetAllRules() {
    const ruleIds = ['lengthRule', 'uppercaseRule', 'lowercaseRule', 'numberRule', 'specialRule'];
    
    ruleIds.forEach(ruleId => {
        const ruleElement = document.getElementById(ruleId);
        if (ruleElement) {
            ruleElement.classList.remove('valid', 'invalid');
            const iconElement = ruleElement.querySelector('.rule-icon');
            if (iconElement) {
                iconElement.textContent = '⚪';
            }
        }
    });
}

// Validation de la correspondance des mots de passe
function validatePasswordMatch(password, confirmPassword) {
    console.log('🔍 Validation correspondance mots de passe');
    
    const matchIndicator = document.getElementById('matchIndicator');
    const confirmPasswordField = document.getElementById('confirmPassword');
    const passwordMatch = document.getElementById('passwordMatch');
    
    if (!matchIndicator || !confirmPasswordField || !passwordMatch) {
        console.log('❌ Éléments de validation de correspondance introuvables');
        return;
    }
    
    // Si la confirmation est vide, cacher la div
    if (confirmPassword.length === 0) {
        passwordMatch.classList.remove('show');
        return;
    }
    
    // Afficher la div si la confirmation n'est pas vide
    passwordMatch.classList.add('show');
    
    const isMatch = password === confirmPassword && confirmPassword.length > 0;
    console.log('✅ Mots de passe correspondent:', isMatch);
    
    // Mettre à jour l'indicateur
    matchIndicator.classList.remove('valid', 'invalid');
    confirmPasswordField.classList.remove('password-matched', 'password-mismatch');
    
    // Trouver l'icône dans l'élément
    const iconElement = matchIndicator.querySelector('.match-icon');
    
    if (isMatch) {
        matchIndicator.classList.add('valid');
        confirmPasswordField.classList.add('password-matched');
        if (iconElement) {
            iconElement.textContent = '✅';
        }
    } else {
        matchIndicator.classList.add('invalid');
        confirmPasswordField.classList.add('password-mismatch');
        if (iconElement) {
            iconElement.textContent = '❌';
        }
    }
}

// Mettre à jour une règle de validation
function updateRule(ruleId, isValid) {
    const ruleElement = document.getElementById(ruleId);
    if (!ruleElement) {
        console.log('❌ Élément de règle introuvable:', ruleId);
        return;
    }
    
    console.log('🔄 Mise à jour règle', ruleId, ':', isValid ? 'VALIDE' : 'INVALIDE');
    
    ruleElement.classList.remove('valid', 'invalid');
    
    // Trouver l'icône dans l'élément
    const iconElement = ruleElement.querySelector('.rule-icon');
    if (iconElement) {
        if (isValid) {
            ruleElement.classList.add('valid');
            iconElement.textContent = '✅';
        } else {
            ruleElement.classList.add('invalid');
            iconElement.textContent = '❌';
        }
    }
}

// Calculer la force du mot de passe
function calculatePasswordStrength(password, rules) {
    if (password.length === 0) return 0;
    
    let strength = 0;
    
    // Règles de base
    if (rules.length) strength++;
    if (rules.uppercase) strength++;
    if (rules.lowercase) strength++;
    if (rules.number) strength++;
    if (rules.special) strength++;
    
    // Bonus pour la longueur
    if (password.length >= 12) strength += 0.5;
    if (password.length >= 16) strength += 0.5;
    
    // Bonus pour la diversité des caractères
    const uniqueChars = new Set(password).size;
    if (uniqueChars >= 8) strength += 0.5;
    
    return Math.min(strength, 5);
}

// Mettre à jour la barre de force du mot de passe
function updatePasswordStrength(strength) {
    const strengthProgress = document.getElementById('strengthProgress');
    const strengthLabel = document.getElementById('strengthLabel');
    
    if (!strengthProgress || !strengthLabel) {
        console.log('❌ Éléments de force du mot de passe introuvables');
        return;
    }
    
    // Supprimer les classes existantes
    strengthProgress.classList.remove('weak', 'fair', 'good', 'strong', 'excellent');
    strengthLabel.classList.remove('weak', 'fair', 'good', 'strong', 'excellent');
    
    let strengthClass = '';
    let strengthText = '';
    
    if (strength === 0) {
        strengthText = 'Force du mot de passe';
    } else if (strength < 2) {
        strengthClass = 'weak';
        strengthText = 'Faible';
    } else if (strength < 3) {
        strengthClass = 'fair';
        strengthText = 'Acceptable';
    } else if (strength < 4) {
        strengthClass = 'good';
        strengthText = 'Bon';
    } else if (strength < 5) {
        strengthClass = 'strong';
        strengthText = 'Fort';
    } else {
        strengthClass = 'excellent';
        strengthText = 'Excellent';
    }
    
    if (strengthClass) {
        strengthProgress.classList.add(strengthClass);
        strengthLabel.classList.add(strengthClass);
    }
    
    strengthLabel.textContent = strengthText;
    console.log('💪 Force mise à jour:', strengthText);
}

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
    const confirmPassword = document.getElementById('confirmPassword');
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
    
    // Validation spécifique au mode inscription
    if (confirmPassword) {
        const confirmPasswordValue = confirmPassword.value;
        
        // Vérifier que la confirmation est remplie
        if (!confirmPasswordValue) {
            e.preventDefault();
            showError('Veuillez confirmer votre mot de passe');
            return;
        }
        
        // Vérifier que les mots de passe correspondent
        if (password !== confirmPasswordValue) {
            e.preventDefault();
            showError('Les mots de passe ne correspondent pas');
            return;
        }
        
        // Vérifier que le mot de passe respecte les contraintes
        if (!isPasswordValid(password)) {
            e.preventDefault();
            showError('Le mot de passe ne respecte pas toutes les contraintes de sécurité');
            return;
        }
    }

    // Afficher un indicateur de chargement
    const originalText = submitBtn.textContent;
    submitBtn.classList.add('loading');
    
    if (confirmPassword) {
        submitBtn.textContent = 'Création du compte...';
    } else {
        submitBtn.textContent = 'Connexion...';
    }
    
    submitBtn.disabled = true;

    // Laisser la soumission normale du formulaire continuer
    // Le serveur va gérer la connexion et rediriger
});

// Vérifier si le mot de passe est valide selon nos règles
function isPasswordValid(password) {
    const rules = {
        length: password.length >= 8,
        uppercase: /[A-Z]/.test(password),
        lowercase: /[a-z]/.test(password),
        number: /[0-9]/.test(password),
        special: /[!@#$%^&*(),.?":{}|<>]/.test(password)
    };
    
    return rules.length && rules.uppercase && rules.lowercase && rules.number && rules.special;
}

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