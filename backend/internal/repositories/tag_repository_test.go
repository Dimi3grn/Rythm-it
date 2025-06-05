package repositories_test

import (
	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/pkg/database"
	"testing"
)

func TestTagRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tag repository test in short mode")
	}

	// Setup database
	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	// Créer le repository
	repo := repositories.NewTagRepository(database.DB)

	// Test 1: Créer un tag
	t.Run("Create Tag", func(t *testing.T) {
		tag := &models.Tag{
			Name: "test-rap",
			Type: "genre",
		}

		err := repo.Create(tag)
		if err != nil {
			t.Fatalf("Erreur création tag: %v", err)
		}

		if tag.ID == 0 {
			t.Fatal("ID du tag n'a pas été assigné")
		}

		t.Logf("✅ Tag créé avec ID: %d", tag.ID)

		// Nettoyer après le test
		repo.Delete(tag.ID)
	})

	// Test 2: Trouver un tag par nom
	t.Run("Find Tag By Name", func(t *testing.T) {
		// Utiliser un tag existant de la base (rap)
		tag, err := repo.FindByName("rap")
		if err != nil {
			t.Fatalf("Erreur recherche tag: %v", err)
		}

		if tag.Name != "rap" {
			t.Errorf("Nom incorrect. Attendu: rap, Obtenu: %s", tag.Name)
		}

		if tag.Type != "genre" {
			t.Errorf("Type incorrect. Attendu: genre, Obtenu: %s", tag.Type)
		}

		t.Logf("✅ Tag trouvé: %s (type: %s)", tag.Name, tag.Type)
	})

	// Test 3: FindOrCreate - tag existant
	t.Run("FindOrCreate Existing Tag", func(t *testing.T) {
		tag, err := repo.FindOrCreate("rap", "genre")
		if err != nil {
			t.Fatalf("Erreur FindOrCreate: %v", err)
		}

		if tag.Name != "rap" {
			t.Errorf("Nom incorrect. Attendu: rap, Obtenu: %s", tag.Name)
		}

		t.Logf("✅ Tag existant récupéré: %s", tag.Name)
	})

	// Test 4: FindOrCreate - nouveau tag
	t.Run("FindOrCreate New Tag", func(t *testing.T) {
		tagName := "test-findorcreate"
		tag, err := repo.FindOrCreate(tagName, "artist")
		if err != nil {
			t.Fatalf("Erreur FindOrCreate nouveau tag: %v", err)
		}

		if tag.Name != tagName {
			t.Errorf("Nom incorrect. Attendu: %s, Obtenu: %s", tagName, tag.Name)
		}

		if tag.ID == 0 {
			t.Fatal("ID du nouveau tag n'a pas été assigné")
		}

		t.Logf("✅ Nouveau tag créé: %s (ID: %d)", tag.Name, tag.ID)

		// Nettoyer
		repo.Delete(tag.ID)
	})

	// Test 5: Récupérer par type
	t.Run("Find Tags By Type", func(t *testing.T) {
		tags, err := repo.FindByType("genre")
		if err != nil {
			t.Fatalf("Erreur récupération tags par type: %v", err)
		}

		if len(tags) == 0 {
			t.Fatal("Aucun tag de type 'genre' trouvé")
		}

		// Vérifier que tous sont du bon type
		for _, tag := range tags {
			if tag.Type != "genre" {
				t.Errorf("Tag avec mauvais type. Attendu: genre, Obtenu: %s", tag.Type)
			}
		}

		t.Logf("✅ Trouvé %d tags de type 'genre'", len(tags))
	})

	// Test 6: Récupérer tous les tags
	t.Run("Find All Tags", func(t *testing.T) {
		tags, err := repo.FindAll()
		if err != nil {
			t.Fatalf("Erreur récupération tous les tags: %v", err)
		}

		if len(tags) < 10 {
			t.Errorf("Attendu au moins 10 tags (base données), obtenu: %d", len(tags))
		}

		t.Logf("✅ Trouvé %d tags au total", len(tags))
	})

	// Test 7: Tags populaires
	t.Run("Get Popular Tags", func(t *testing.T) {
		tags, err := repo.GetPopularTags(5)
		if err != nil {
			t.Fatalf("Erreur récupération tags populaires: %v", err)
		}

		if len(tags) > 5 {
			t.Errorf("Trop de tags retournés. Attendu: ≤5, Obtenu: %d", len(tags))
		}

		t.Logf("✅ Tags populaires: %d tags", len(tags))
	})

	// Test 8: Recherche de tags
	t.Run("Search Tags", func(t *testing.T) {
		// Rechercher "ra" (devrait trouver "rap")
		tags, err := repo.SearchTags("ra", "", 10)
		if err != nil {
			t.Fatalf("Erreur recherche tags: %v", err)
		}

		found := false
		for _, tag := range tags {
			if tag.Name == "rap" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Tag 'rap' non trouvé dans la recherche 'ra'")
		}

		t.Logf("✅ Recherche 'ra': %d résultats", len(tags))
	})

	// Test 9: Recherche par type spécifique
	t.Run("Search Tags By Type", func(t *testing.T) {
		tags, err := repo.SearchTags("dr", "artist", 10)
		if err != nil {
			t.Fatalf("Erreur recherche tags par type: %v", err)
		}

		// Vérifier que tous sont du bon type
		for _, tag := range tags {
			if tag.Type != "artist" {
				t.Errorf("Tag avec mauvais type. Attendu: artist, Obtenu: %s", tag.Type)
			}
		}

		t.Logf("✅ Recherche 'dr' dans artistes: %d résultats", len(tags))
	})

	// Test 10: Mise à jour d'un tag
	t.Run("Update Tag", func(t *testing.T) {
		// Créer un tag temporaire
		tag := &models.Tag{
			Name: "test-update",
			Type: "genre",
		}

		err := repo.Create(tag)
		if err != nil {
			t.Fatalf("Erreur création tag: %v", err)
		}

		// Modifier le tag
		tag.Name = "test-updated"
		tag.Type = "artist"

		err = repo.Update(tag)
		if err != nil {
			t.Fatalf("Erreur mise à jour tag: %v", err)
		}

		// Vérifier la modification
		updated, err := repo.FindByID(tag.ID)
		if err != nil {
			t.Fatalf("Erreur récupération tag modifié: %v", err)
		}

		if updated.Name != "test-updated" {
			t.Errorf("Nom non modifié. Attendu: test-updated, Obtenu: %s", updated.Name)
		}

		if updated.Type != "artist" {
			t.Errorf("Type non modifié. Attendu: artist, Obtenu: %s", updated.Type)
		}

		t.Logf("✅ Tag mis à jour: %s", updated.Name)

		// Nettoyer
		repo.Delete(tag.ID)
	})

	// Test 11: Comptage d'usage
	t.Run("Get Tag Usage Count", func(t *testing.T) {
		// Utiliser un tag existant (rap)
		tag, err := repo.FindByName("rap")
		if err != nil {
			t.Fatalf("Erreur recherche tag rap: %v", err)
		}

		count, err := repo.GetTagUsageCount(tag.ID)
		if err != nil {
			t.Fatalf("Erreur comptage usage: %v", err)
		}

		t.Logf("✅ Tag 'rap' utilisé dans %d thread(s)", count)
	})

	// Test 12: Suppression (tag non utilisé)
	t.Run("Delete Unused Tag", func(t *testing.T) {
		// Créer un tag temporaire
		tag := &models.Tag{
			Name: "test-delete",
			Type: "genre",
		}

		err := repo.Create(tag)
		if err != nil {
			t.Fatalf("Erreur création tag: %v", err)
		}

		tagID := tag.ID

		// Supprimer le tag
		err = repo.Delete(tagID)
		if err != nil {
			t.Fatalf("Erreur suppression tag: %v", err)
		}

		// Vérifier que le tag n'existe plus
		_, err = repo.FindByID(tagID)
		if err == nil {
			t.Fatal("Le tag existe encore après suppression")
		}

		t.Logf("✅ Tag %d supprimé avec succès", tagID)
	})

	// Test 13: Normalisation des noms
	t.Run("Name Normalization", func(t *testing.T) {
		tag := &models.Tag{
			Name: "  Test-CASE-normalization  ",
			Type: "genre",
		}

		err := repo.Create(tag)
		if err != nil {
			t.Fatalf("Erreur création tag: %v", err)
		}

		// Vérifier que le nom a été normalisé
		if tag.Name != "test-case-normalization" {
			t.Errorf("Nom non normalisé. Attendu: test-case-normalization, Obtenu: %s", tag.Name)
		}

		t.Logf("✅ Nom normalisé: '%s'", tag.Name)

		// Nettoyer
		repo.Delete(tag.ID)
	})
}

func TestTagRepository_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tag repository edge cases test in short mode")
	}

	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	repo := repositories.NewTagRepository(database.DB)

	// Test: Créer un tag en double
	t.Run("Create Duplicate Tag", func(t *testing.T) {
		tag1 := &models.Tag{
			Name: "test-duplicate",
			Type: "genre",
		}

		err := repo.Create(tag1)
		if err != nil {
			t.Fatalf("Erreur création premier tag: %v", err)
		}

		// Essayer de créer le même tag
		tag2 := &models.Tag{
			Name: "test-duplicate",
			Type: "genre",
		}

		err = repo.Create(tag2)
		if err == nil {
			t.Fatal("La création d'un tag en double devrait échouer")
		}

		t.Logf("✅ Duplicata correctement rejeté: %v", err)

		// Nettoyer
		repo.Delete(tag1.ID)
	})

	// Test: Recherche avec chaîne vide
	t.Run("Search Empty String", func(t *testing.T) {
		tags, err := repo.SearchTags("", "", 10)
		if err != nil {
			t.Fatalf("Erreur recherche chaîne vide: %v", err)
		}

		t.Logf("✅ Recherche chaîne vide: %d résultats", len(tags))
	})

	// Test: Tag inexistant
	t.Run("Find Non-Existent Tag", func(t *testing.T) {
		_, err := repo.FindByName("tag-inexistant-999")
		if err == nil {
			t.Fatal("La recherche d'un tag inexistant devrait échouer")
		}

		t.Logf("✅ Tag inexistant correctement géré: %v", err)
	})
}
