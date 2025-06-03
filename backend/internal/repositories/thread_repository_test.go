package repositories_test

import (
	"fmt"
	"testing"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/pkg/database"
)

func TestThreadRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping thread repository test in short mode")
	}

	// Setup database
	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	// Créer le repository
	repo := repositories.NewThreadRepository(database.DB)

	// Test 1: Créer un thread
	t.Run("Create Thread", func(t *testing.T) {
		thread := &models.Thread{
			Title:       "Test Drake vs Kendrick",
			Description: "Qui est le meilleur rappeur selon vous ?",
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1, // Utilise l'admin créé par défaut
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		if thread.ID == 0 {
			t.Fatal("ID du thread n'a pas été assigné")
		}

		t.Logf("✅ Thread créé avec ID: %d", thread.ID)
	})

	// Test 2: Récupérer un thread par ID
	t.Run("Find Thread By ID", func(t *testing.T) {
		// D'abord créer un thread
		thread := &models.Thread{
			Title:       "Test Find By ID",
			Description: "Thread pour tester FindByID",
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Maintenant le récupérer
		found, err := repo.FindByID(thread.ID)
		if err != nil {
			t.Fatalf("Erreur récupération thread: %v", err)
		}

		if found.Title != thread.Title {
			t.Errorf("Titre incorrect. Attendu: %s, Obtenu: %s", thread.Title, found.Title)
		}

		if found.Author == nil {
			t.Fatal("Auteur non chargé")
		}

		if found.Author.Username != "admin" {
			t.Errorf("Username auteur incorrect. Attendu: admin, Obtenu: %s", found.Author.Username)
		}

		t.Logf("✅ Thread récupéré: %s par %s", found.Title, found.Author.Username)
	})

	// Test 3: Lister les threads publics
	t.Run("Find Public Threads", func(t *testing.T) {
		params := models.PaginationParams{
			Page:    1,
			PerPage: 5,
			Sort:    "id",
			Order:   "DESC",
		}

		threads, total, err := repo.FindPublicThreads(params)
		if err != nil {
			t.Fatalf("Erreur récupération threads publics: %v", err)
		}

		t.Logf("✅ Trouvé %d threads publics (total: %d)", len(threads), total)

		// Vérifier que tous les threads sont publics et non archivés
		for _, thread := range threads {
			if thread.Visibility != models.VisibilityPublic {
				t.Errorf("Thread non public trouvé: %s", thread.Title)
			}
			if thread.State == models.ThreadStateArchived {
				t.Errorf("Thread archivé trouvé: %s", thread.Title)
			}
			if thread.Author == nil {
				t.Errorf("Auteur manquant pour thread: %s", thread.Title)
			}
		}
	})

	// Test 4: Mise à jour d'un thread
	t.Run("Update Thread", func(t *testing.T) {
		// Créer un thread
		thread := &models.Thread{
			Title:       "Thread à modifier",
			Description: "Description originale",
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Modifier le thread
		thread.Title = "Thread modifié"
		thread.Description = "Nouvelle description"
		thread.State = models.ThreadStateClosed

		err = repo.Update(thread)
		if err != nil {
			t.Fatalf("Erreur mise à jour thread: %v", err)
		}

		// Vérifier la modification
		updated, err := repo.FindByID(thread.ID)
		if err != nil {
			t.Fatalf("Erreur récupération thread modifié: %v", err)
		}

		if updated.Title != "Thread modifié" {
			t.Errorf("Titre non modifié. Attendu: Thread modifié, Obtenu: %s", updated.Title)
		}

		if updated.State != models.ThreadStateClosed {
			t.Errorf("État non modifié. Attendu: %s, Obtenu: %s", models.ThreadStateClosed, updated.State)
		}

		t.Logf("✅ Thread mis à jour: %s", updated.Title)
	})

	// Test 5: Changer l'état d'un thread
	t.Run("Update Thread State", func(t *testing.T) {
		// Créer un thread
		thread := &models.Thread{
			Title:       "Thread pour test état",
			Description: "Test changement d'état",
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		// Changer l'état
		err = repo.UpdateState(thread.ID, models.ThreadStateArchived)
		if err != nil {
			t.Fatalf("Erreur changement état: %v", err)
		}

		// Vérifier le changement
		updated, err := repo.FindByID(thread.ID)
		if err != nil {
			t.Fatalf("Erreur récupération thread: %v", err)
		}

		if updated.State != models.ThreadStateArchived {
			t.Errorf("État non changé. Attendu: %s, Obtenu: %s", models.ThreadStateArchived, updated.State)
		}

		t.Logf("✅ État du thread changé en: %s", updated.State)
	})

	// Test 6: Suppression d'un thread
	t.Run("Delete Thread", func(t *testing.T) {
		// Créer un thread
		thread := &models.Thread{
			Title:       "Thread à supprimer",
			Description: "Ce thread sera supprimé",
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread: %v", err)
		}

		threadID := thread.ID

		// Supprimer le thread
		err = repo.Delete(threadID)
		if err != nil {
			t.Fatalf("Erreur suppression thread: %v", err)
		}

		// Vérifier que le thread n'existe plus
		_, err = repo.FindByID(threadID)
		if err == nil {
			t.Fatal("Le thread existe encore après suppression")
		}

		t.Logf("✅ Thread %d supprimé avec succès", threadID)
	})
}

func TestThreadRepositoryPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pagination test in short mode")
	}

	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	repo := repositories.NewThreadRepository(database.DB)

	// Créer plusieurs threads pour tester la pagination
	for i := 1; i <= 15; i++ {
		thread := &models.Thread{
			Title:       fmt.Sprintf("Thread pagination %d", i),
			Description: fmt.Sprintf("Description du thread %d", i),
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread %d: %v", i, err)
		}
	}

	// Test pagination
	params := models.PaginationParams{
		Page:    1,
		PerPage: 10,
		Sort:    "id",
		Order:   "DESC",
	}

	threads, total, err := repo.FindPublicThreads(params)
	if err != nil {
		t.Fatalf("Erreur récupération threads paginés: %v", err)
	}

	if len(threads) > 10 {
		t.Errorf("Trop de threads retournés. Attendu: ≤10, Obtenu: %d", len(threads))
	}

	if total < 15 {
		t.Errorf("Total incorrect. Attendu: ≥15, Obtenu: %d", total)
	}

	t.Logf("✅ Pagination: %d threads sur %d total", len(threads), total)

	// Test page 2
	params.Page = 2
	threadsPage2, _, err := repo.FindPublicThreads(params)
	if err != nil {
		t.Fatalf("Erreur récupération page 2: %v", err)
	}

	// Vérifier que les threads de la page 2 sont différents
	if len(threadsPage2) > 0 && len(threads) > 0 {
		if threadsPage2[0].ID == threads[0].ID {
			t.Error("Page 2 retourne les mêmes threads que page 1")
		}
	}

	t.Logf("✅ Page 2: %d threads", len(threadsPage2))
}

func TestFindPublicThreads(t *testing.T) {
	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	repo := repositories.NewThreadRepository(database.DB)

	// Créer quelques threads de test
	for i := 1; i <= 3; i++ {
		thread := &models.Thread{
			Title:       fmt.Sprintf("Thread public %d", i),
			Description: fmt.Sprintf("Description du thread public %d", i),
			State:       models.ThreadStateOpen,
			Visibility:  models.VisibilityPublic,
			UserID:      1,
		}

		err := repo.Create(thread)
		if err != nil {
			t.Fatalf("Erreur création thread %d: %v", i, err)
		}
	}

	// Créer un thread privé
	privateThread := &models.Thread{
		Title:       "Thread privé",
		Description: "Ce thread ne devrait pas apparaître",
		State:       models.ThreadStateOpen,
		Visibility:  models.VisibilityPrivate,
		UserID:      1,
	}

	err = repo.Create(privateThread)
	if err != nil {
		t.Fatalf("Erreur création thread privé: %v", err)
	}

	// Tester la récupération des threads publics
	params := models.DefaultPagination()
	params.PerPage = 5

	threads, total, err := repo.FindPublicThreads(params)
	if err != nil {
		t.Fatalf("Erreur récupération threads publics: %v", err)
	}

	if total < 3 {
		t.Errorf("Total incorrect. Attendu: ≥3, Obtenu: %d", total)
	}

	// Vérifier que tous les threads sont publics
	for _, thread := range threads {
		if thread.Visibility != models.VisibilityPublic {
			t.Errorf("Thread non public trouvé: %s", thread.Title)
		}
		if thread.State == models.ThreadStateArchived {
			t.Errorf("Thread archivé trouvé: %s", thread.Title)
		}
	}

	t.Logf("✅ Trouvé %d threads publics (total: %d)", len(threads), total)
}
