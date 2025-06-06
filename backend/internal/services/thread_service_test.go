package services

import (
	"fmt"
	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/pkg/database"
	"testing"
)

// DefaultPagination returns default pagination parameters for testing
func DefaultPagination() models.PaginationParams {
	return models.PaginationParams{
		Page:    1,
		PerPage: 10,
		Sort:    "created_at",
		Order:   "DESC",
	}
}

func TestThreadService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping thread service test in short mode")
	}

	// Setup database
	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	// Créer les repositories
	threadRepo := repositories.NewThreadRepository(database.DB)
	tagRepo := repositories.NewTagRepository(database.DB)
	messageRepo := repositories.NewMessageRepository(database.DB)

	// Créer le service
	service := NewThreadService(threadRepo, tagRepo, messageRepo, database.DB)

	// Test 1: Créer un thread avec tags
	t.Run("Create Thread With Tags", func(t *testing.T) {
		dto := CreateThreadDTO{
			Title:       "Test Drake vs Kendrick Service",
			Description: "Un thread créé via le service pour tester la logique métier",
			Tags:        []string{"rap", "drake", "kendrick lamar", "hip-hop"},
			Visibility:  "public",
		}

		threadResponse, err := service.CreateThread(dto, 1) // userID = 1 (admin)
		if err != nil {
			t.Fatalf("Erreur création thread service: %v", err)
		}

		if threadResponse.ID == 0 {
			t.Fatal("ID du thread n'a pas été assigné")
		}

		if threadResponse.Title != dto.Title {
			t.Errorf("Titre incorrect. Attendu: %s, Obtenu: %s", dto.Title, threadResponse.Title)
		}

		if len(threadResponse.Tags) == 0 {
			t.Fatal("Aucun tag attaché au thread")
		}

		// Vérifier que les tags ont été créés/attachés
		foundRap := false
		foundDrake := false
		for _, tag := range threadResponse.Tags {
			if tag.Name == "rap" && tag.Type == "genre" {
				foundRap = true
			}
			if tag.Name == "drake" && tag.Type == "artist" {
				foundDrake = true
			}
		}

		if !foundRap {
			t.Error("Tag 'rap' non trouvé ou mauvais type")
		}
		if !foundDrake {
			t.Error("Tag 'drake' non trouvé ou mauvais type")
		}

		t.Logf("✅ Thread créé avec ID: %d et %d tags", threadResponse.ID, len(threadResponse.Tags))
	})

	// Test 2: Récupérer un thread
	t.Run("Get Thread", func(t *testing.T) {
		// D'abord créer un thread
		dto := CreateThreadDTO{
			Title:       "Thread pour test Get",
			Description: "Description du thread pour test Get",
			Tags:        []string{"test", "get"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Maintenant le récupérer
		userID := uint(1)
		retrieved, err := service.GetThread(created.ID, &userID)
		if err != nil {
			t.Fatalf("Erreur récupération thread: %v", err)
		}

		if retrieved.ID != created.ID {
			t.Errorf("ID incorrect. Attendu: %d, Obtenu: %d", created.ID, retrieved.ID)
		}

		if retrieved.Author.Username != "admin" {
			t.Errorf("Auteur incorrect. Attendu: admin, Obtenu: %s", retrieved.Author.Username)
		}

		t.Logf("✅ Thread récupéré: %s par %s", retrieved.Title, retrieved.Author.Username)
	})

	// Test 3: Lister les threads publics
	t.Run("Get Public Threads", func(t *testing.T) {
		params := DefaultPagination()
		params.PerPage = 5

		result, err := service.GetPublicThreads(params, ThreadFilters{})
		if err != nil {
			t.Fatalf("Erreur récupération threads publics: %v", err)
		}

		if len(result.Threads) > 5 {
			t.Errorf("Trop de threads retournés. Attendu: ≤5, Obtenu: %d", len(result.Threads))
		}

		if result.Pagination.Page != 1 {
			t.Errorf("Page incorrecte. Attendu: 1, Obtenu: %d", result.Pagination.Page)
		}

		// Vérifier que tous les threads sont publics
		for _, thread := range result.Threads {
			if thread.Visibility != "public" {
				t.Errorf("Thread non public trouvé: %s", thread.Title)
			}
			if thread.State == "archivé" {
				t.Errorf("Thread archivé trouvé: %s", thread.Title)
			}
		}

		t.Logf("✅ Trouvé %d threads publics (total: %d)", len(result.Threads), result.Pagination.Total)
	})

	// Test 4: Recherche de threads
	t.Run("Search Threads", func(t *testing.T) {
		// Créer un thread avec un titre spécifique
		dto := CreateThreadDTO{
			Title:       "Recherche Test Unique 12345",
			Description: "Thread pour tester la recherche",
			Tags:        []string{"search", "test"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread recherche: %v", err)
		}

		// Rechercher ce thread
		params := DefaultPagination()
		result, err := service.SearchThreads("Unique 12345", params)
		if err != nil {
			t.Fatalf("Erreur recherche threads: %v", err)
		}

		// Vérifier qu'on trouve le thread
		found := false
		for _, thread := range result.Threads {
			if thread.ID == created.ID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Thread créé non trouvé dans la recherche")
		}

		t.Logf("✅ Recherche: %d résultats", len(result.Threads))
	})

	// Test 5: Filtrage par tag
	t.Run("Get Threads By Tag", func(t *testing.T) {
		// Créer un thread avec un tag spécifique
		uniqueTag := "test-filter-unique"
		dto := CreateThreadDTO{
			Title:       "Thread pour test filtrage",
			Description: "Thread avec un tag unique",
			Tags:        []string{uniqueTag, "test"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread filtrage: %v", err)
		}

		// Filtrer par ce tag
		params := DefaultPagination()
		result, err := service.GetThreadsByTag(uniqueTag, params)
		if err != nil {
			t.Fatalf("Erreur filtrage par tag: %v", err)
		}

		// Vérifier qu'on trouve le thread
		found := false
		for _, thread := range result.Threads {
			if thread.ID == created.ID {
				found = true
				// Vérifier que le tag est présent
				tagFound := false
				for _, tag := range thread.Tags {
					if tag.Name == uniqueTag {
						tagFound = true
						break
					}
				}
				if !tagFound {
					t.Errorf("Tag '%s' non trouvé dans le thread", uniqueTag)
				}
				break
			}
		}

		if !found {
			t.Error("Thread avec tag spécifique non trouvé")
		}

		t.Logf("✅ Filtrage par tag '%s': %d résultats", uniqueTag, len(result.Threads))
	})

	// Test 6: Mise à jour d'un thread
	t.Run("Update Thread", func(t *testing.T) {
		// Créer un thread
		dto := CreateThreadDTO{
			Title:       "Thread à modifier",
			Description: "Description originale",
			Tags:        []string{"original"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Modifier le thread
		updateDTO := UpdateThreadDTO{
			Title:       "Thread modifié",
			Description: "Nouvelle description",
			State:       "fermé",
			Visibility:  "public",
		}

		err = service.UpdateThread(created.ID, updateDTO, 1, false)
		if err != nil {
			t.Fatalf("Erreur mise à jour thread: %v", err)
		}

		// Vérifier la modification
		userID := uint(1)
		updated, err := service.GetThread(created.ID, &userID)
		if err != nil {
			t.Fatalf("Erreur récupération thread modifié: %v", err)
		}

		if updated.Title != "Thread modifié" {
			t.Errorf("Titre non modifié. Attendu: Thread modifié, Obtenu: %s", updated.Title)
		}

		if updated.State != "fermé" {
			t.Errorf("État non modifié. Attendu: fermé, Obtenu: %s", updated.State)
		}

		t.Logf("✅ Thread mis à jour: %s", updated.Title)
	})

	// Test 7: Changement d'état
	t.Run("Change Thread State", func(t *testing.T) {
		// Créer un thread
		dto := CreateThreadDTO{
			Title:       "Thread pour test état",
			Description: "Thread pour tester le changement d'état",
			Tags:        []string{"test", "state"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Changer l'état
		err = service.ChangeThreadState(created.ID, "archivé", 1, true) // isAdmin = true
		if err != nil {
			t.Fatalf("Erreur changement état: %v", err)
		}

		// Vérifier le changement (en tant qu'admin)
		userID := uint(1)
		updated, err := service.GetThread(created.ID, &userID)
		if err != nil {
			t.Fatalf("Erreur récupération thread archivé: %v", err)
		}

		if updated.State != "archivé" {
			t.Errorf("État non changé. Attendu: archivé, Obtenu: %s", updated.State)
		}

		t.Logf("✅ État du thread changé en: %s", updated.State)
	})

	// Test 8: Validation des erreurs
	t.Run("Validation Errors", func(t *testing.T) {
		// DTO invalide (titre trop court)
		dto := CreateThreadDTO{
			Title:       "Test", // Trop court (< 5 chars)
			Description: "Description",
			Tags:        []string{"test"},
			Visibility:  "public",
		}

		_, err := service.CreateThread(dto, 1)
		if err == nil {
			t.Fatal("La validation devrait échouer pour un titre trop court")
		}

		t.Logf("✅ Validation correctement rejetée: %v", err)
	})

	// Test 9: Permissions
	t.Run("Permission Checks", func(t *testing.T) {
		// Créer un thread avec un utilisateur
		dto := CreateThreadDTO{
			Title:       "Thread privé pour permissions",
			Description: "Thread pour tester les permissions",
			Tags:        []string{"private", "test"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1) // userID = 1
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Essayer de modifier avec un autre utilisateur (non admin)
		updateDTO := UpdateThreadDTO{
			Title:       "Modification non autorisée",
			Description: "Ne devrait pas marcher",
			State:       "fermé",
			Visibility:  "public",
		}

		err = service.UpdateThread(created.ID, updateDTO, 999, false) // userID différent, pas admin
		if err == nil {
			t.Fatal("La modification devrait être refusée pour un utilisateur non autorisé")
		}

		t.Logf("✅ Permissions correctement vérifiées: %v", err)
	})

	// Test 10: Threads utilisateur
	t.Run("Get User Threads", func(t *testing.T) {
		// Créer plusieurs threads pour un utilisateur
		for i := 0; i < 3; i++ {
			dto := CreateThreadDTO{
				Title:       fmt.Sprintf("Thread utilisateur %d", i+1),
				Description: fmt.Sprintf("Description %d", i+1),
				Tags:        []string{"user-test", fmt.Sprintf("thread-%d", i+1)},
				Visibility:  "public",
			}

			_, err := service.CreateThread(dto, 1)
			if err != nil {
				t.Fatalf("Erreur création thread utilisateur %d: %v", i+1, err)
			}
		}

		// Récupérer les threads de l'utilisateur
		params := DefaultPagination()
		params.PerPage = 10

		result, err := service.GetUserThreads(1, params)
		if err != nil {
			t.Fatalf("Erreur récupération threads utilisateur: %v", err)
		}

		if len(result.Threads) < 3 {
			t.Errorf("Pas assez de threads utilisateur. Attendu: ≥3, Obtenu: %d", len(result.Threads))
		}

		// Vérifier que tous appartiennent au bon utilisateur
		for _, thread := range result.Threads {
			if thread.Author.ID != 1 {
				t.Errorf("Thread d'un autre utilisateur trouvé. Attendu ID: 1, Obtenu: %d", thread.Author.ID)
			}
		}

		t.Logf("✅ Threads utilisateur: %d trouvés", len(result.Threads))
	})
}

func TestThreadService_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping thread service edge cases test in short mode")
	}

	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	threadRepo := repositories.NewThreadRepository(database.DB)
	tagRepo := repositories.NewTagRepository(database.DB)
	messageRepo := repositories.NewMessageRepository(database.DB)
	service := NewThreadService(threadRepo, tagRepo, messageRepo, database.DB)

	// Test: Thread inexistant
	t.Run("Get Non-Existent Thread", func(t *testing.T) {
		userID := uint(1)
		_, err := service.GetThread(99999, &userID)
		if err == nil {
			t.Fatal("La récupération d'un thread inexistant devrait échouer")
		}

		t.Logf("✅ Thread inexistant correctement géré: %v", err)
	})

	// Test: Tags vides ou invalides
	t.Run("Create Thread With Empty Tags", func(t *testing.T) {
		dto := CreateThreadDTO{
			Title:       "Thread avec tags vides",
			Description: "Test des tags vides ou espaces",
			Tags:        []string{"valid-tag", "", "  ", "another-valid"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread avec tags vides: %v", err)
		}

		// Vérifier que seuls les tags valides ont été ajoutés
		if len(created.Tags) != 2 {
			t.Errorf("Nombre de tags incorrect. Attendu: 2, Obtenu: %d", len(created.Tags))
		}

		validTags := []string{"valid-tag", "another-valid"}
		for _, expectedTag := range validTags {
			found := false
			for _, tag := range created.Tags {
				if tag.Name == expectedTag {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Tag valide '%s' non trouvé", expectedTag)
			}
		}

		t.Logf("✅ Tags vides correctement filtrés: %d tags valides", len(created.Tags))
	})

	// Test: Recherche avec chaîne vide
	t.Run("Search With Empty Query", func(t *testing.T) {
		params := DefaultPagination()
		params.PerPage = 5

		result, err := service.SearchThreads("", params)
		if err != nil {
			t.Fatalf("Erreur recherche chaîne vide: %v", err)
		}

		// Devrait retourner les threads publics normaux
		if len(result.Threads) == 0 && result.Pagination.Total > 0 {
			t.Error("La recherche vide devrait retourner des threads")
		}

		t.Logf("✅ Recherche chaîne vide: %d résultats", len(result.Threads))
	})

	// Test: Tag inexistant pour filtrage
	t.Run("Filter By Non-Existent Tag", func(t *testing.T) {
		params := DefaultPagination()

		result, err := service.GetThreadsByTag("tag-inexistant-xyz", params)
		if err != nil {
			t.Fatalf("Erreur filtrage tag inexistant: %v", err)
		}

		// Devrait retourner une liste vide
		if len(result.Threads) != 0 {
			t.Errorf("Le filtrage par tag inexistant devrait retourner 0 résultats, obtenu: %d", len(result.Threads))
		}

		if result.Pagination.Total != 0 {
			t.Errorf("Le total devrait être 0, obtenu: %d", result.Pagination.Total)
		}

		t.Logf("✅ Filtrage tag inexistant: 0 résultats (correct)")
	})

	// Test: État invalide
	t.Run("Change Thread To Invalid State", func(t *testing.T) {
		// Créer un thread
		dto := CreateThreadDTO{
			Title:       "Thread pour état invalide",
			Description: "Test état invalide",
			Tags:        []string{"test"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Essayer de changer vers un état invalide
		err = service.ChangeThreadState(created.ID, "état-invalide", 1, true)
		if err == nil {
			t.Fatal("Le changement vers un état invalide devrait échouer")
		}

		t.Logf("✅ État invalide correctement rejeté: %v", err)
	})

	// Test: Suppression de thread
	t.Run("Delete Thread", func(t *testing.T) {
		// Créer un thread
		dto := CreateThreadDTO{
			Title:       "Thread à supprimer",
			Description: "Ce thread sera supprimé",
			Tags:        []string{"delete-test"},
			Visibility:  "public",
		}

		created, err := service.CreateThread(dto, 1)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		threadID := created.ID

		// Supprimer le thread
		err = service.DeleteThread(threadID, 1, false) // propriétaire, pas admin
		if err != nil {
			t.Fatalf("Erreur suppression thread: %v", err)
		}

		// Vérifier que le thread n'existe plus
		userID := uint(1)
		_, err = service.GetThread(threadID, &userID)
		if err == nil {
			t.Fatal("Le thread existe encore après suppression")
		}

		t.Logf("✅ Thread %d supprimé avec succès", threadID)
	})
}
