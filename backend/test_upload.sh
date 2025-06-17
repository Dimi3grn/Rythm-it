#!/bin/bash

# Script de test pour vérifier l'upload d'images

echo "🧪 Test de la fonctionnalité d'upload d'images"

# Vérifier que les dossiers d'upload existent
echo "📁 Vérification des dossiers d'upload..."

if [ ! -d "uploads/threads" ]; then
    echo "❌ Dossier uploads/threads n'existe pas"
    mkdir -p uploads/threads
    echo "✅ Dossier uploads/threads créé"
else
    echo "✅ Dossier uploads/threads existe"
fi

if [ ! -d "uploads/profiles" ]; then
    echo "❌ Dossier uploads/profiles n'existe pas"
    mkdir -p uploads/profiles
    echo "✅ Dossier uploads/profiles créé"
else
    echo "✅ Dossier uploads/profiles existe"
fi

# Vérifier les permissions
echo "🔐 Vérification des permissions..."
chmod 755 uploads/threads
chmod 755 uploads/profiles
echo "✅ Permissions configurées"

# Créer une image de test si elle n'existe pas
if [ ! -f "test_image.png" ]; then
    echo "🖼️ Création d'une image de test..."
    # Créer une image de test simple avec ImageMagick (si disponible)
    if command -v convert &> /dev/null; then
        convert -size 200x200 xc:lightblue -pointsize 20 -fill black -gravity center -annotate +0+0 "Test Image" test_image.png
        echo "✅ Image de test créée: test_image.png"
    else
        echo "⚠️ ImageMagick non disponible, veuillez créer manuellement test_image.png"
    fi
fi

echo ""
echo "🎯 Pour tester l'upload:"
echo "1. Démarrez le serveur: go run cmd/main.go"
echo "2. Allez sur http://localhost:8080"
echo "3. Connectez-vous"
echo "4. Cliquez sur le bouton 📷 dans le formulaire"
echo "5. Sélectionnez une image"
echo "6. Vérifiez qu'elle apparaît en prévisualisation"
echo "7. Publiez le thread et vérifiez l'affichage"

echo ""
echo "✅ Préparation terminée !" 