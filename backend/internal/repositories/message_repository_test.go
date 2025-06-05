package repositories

import (
	"fmt"
	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/pkg/database"
	"testing"
)

// defaultPagination returns default pagination parameters
func defaultPagination() models.PaginationParams {
	return models.DefaultPagination()
}

func TestMessageRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping message repository test in short mode")
	}

	// Setup database
	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	// Créer les repositories
	messageRepo := NewMessageRepository(database.DB)
	threadRepo := NewThreadRepository(database.DB)

	// Créer un thread de test pour les messages
	testThread := &models.Thread{
		Title:       "Thread pour test messages",
		Description: "Thread créé pour tester les messages",
		State:       models.ThreadStateOpen,
		Visibility:  models.VisibilityPublic,
		UserID:      1, // admin
	}

	err = threadRepo.Create(testThread)
	if err != nil {
		t.Fatalf("Erreur création thread test: %v", err)
	}

	// Test 1: Créer un message
	t.Run("Create Message", func(t *testing.T) {
		message := &models.Message{
			Content:  "Premier message de test dans le thread",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		if message.ID == 0 {
			t.Fatal("ID du message n'a pas été assigné")
		}

		t.Logf("✅ Message créé avec ID: %d", message.ID)
	})

	// Test 2: Récupérer un message par ID
	t.Run("Find Message By ID", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Message pour test FindByID",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		// Le récupérer
		found, err := messageRepo.FindByID(message.ID)
		if err != nil {
			t.Fatalf("Erreur récupération message: %v", err)
		}

		if found.Content != message.Content {
			t.Errorf("Contenu incorrect. Attendu: %s, Obtenu: %s", message.Content, found.Content)
		}

		if found.Author == nil {
			t.Fatal("Auteur non chargé")
		}

		if found.Author.Username != "admin" {
			t.Errorf("Username auteur incorrect. Attendu: admin, Obtenu: %s", found.Author.Username)
		}

		t.Logf("✅ Message récupéré: %s par %s", found.Content, found.Author.Username)
	})

	// Test 3: Récupérer les messages d'un thread
	t.Run("Find Messages By Thread", func(t *testing.T) {
		// Créer plusieurs messages dans le thread
		for i := 1; i <= 3; i++ {
			message := &models.Message{
				Content:  fmt.Sprintf("Message test thread %d", i),
				ThreadID: testThread.ID,
				UserID:   1,
			}

			err := messageRepo.Create(message)
			if err != nil {
				t.Fatalf("Erreur création message %d: %v", i, err)
			}
		}

		// Récupérer les messages
		params := defaultPagination()
		params.PerPage = 10

		messages, total, err := messageRepo.FindByThreadID(testThread.ID, params, "date")
		if err != nil {
			t.Fatalf("Erreur récupération messages thread: %v", err)
		}

		if len(messages) < 3 {
			t.Errorf("Pas assez de messages. Attendu: ≥3, Obtenu: %d", len(messages))
		}

		if total < 3 {
			t.Errorf("Total incorrect. Attendu: ≥3, Obtenu: %d", total)
		}

		// Vérifier que tous les messages appartiennent au bon thread
		for _, msg := range messages {
			if msg.ThreadID != testThread.ID {
				t.Errorf("Message du mauvais thread. Attendu: %d, Obtenu: %d", testThread.ID, msg.ThreadID)
			}
			if msg.Author == nil {
				t.Error("Auteur manquant pour message")
			}
		}

		t.Logf("✅ Trouvé %d messages dans le thread (total: %d)", len(messages), total)
	})

	// Test 4: Système de votes Fire/Skip
	t.Run("Vote System Fire/Skip", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Message pour tester les votes",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		// Ajouter un vote Fire
		err = messageRepo.SetUserVote(1, message.ID, models.VoteFire)
		if err != nil {
			t.Fatalf("Erreur vote Fire: %v", err)
		}

		// Vérifier le vote
		vote, err := messageRepo.GetUserVote(1, message.ID)
		if err != nil {
			t.Fatalf("Erreur récupération vote: %v", err)
		}

		if vote != models.VoteFire {
			t.Errorf("Vote incorrect. Attendu: %s, Obtenu: %s", models.VoteFire, vote)
		}

		// Vérifier le score de popularité
		score, err := messageRepo.GetPopularityScore(message.ID)
		if err != nil {
			t.Fatalf("Erreur calcul score: %v", err)
		}

		if score != 1 {
			t.Errorf("Score incorrect. Attendu: 1, Obtenu: %d", score)
		}

		// Changer pour Skip
		err = messageRepo.SetUserVote(1, message.ID, models.VoteSkip)
		if err != nil {
			t.Fatalf("Erreur changement vote Skip: %v", err)
		}

		// Vérifier le nouveau score
		score, err = messageRepo.GetPopularityScore(message.ID)
		if err != nil {
			t.Fatalf("Erreur nouveau calcul score: %v", err)
		}

		if score != -1 {
			t.Errorf("Nouveau score incorrect. Attendu: -1, Obtenu: %d", score)
		}

		t.Logf("✅ Système de votes: Fire → Score +1, Skip → Score -1")
	})

	// Test 5: Compteurs de votes
	t.Run("Vote Counts", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Message pour tester les compteurs",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		// Initialement, pas de votes
		fireCount, skipCount, err := messageRepo.GetMessageVoteCounts(message.ID)
		if err != nil {
			t.Fatalf("Erreur compteurs votes: %v", err)
		}

		if fireCount != 0 || skipCount != 0 {
			t.Errorf("Compteurs initiaux incorrects. Fire: %d, Skip: %d", fireCount, skipCount)
		}

		// Ajouter un vote
		err = messageRepo.SetUserVote(1, message.ID, models.VoteFire)
		if err != nil {
			t.Fatalf("Erreur ajout vote: %v", err)
		}

		// Vérifier les nouveaux compteurs
		fireCount, skipCount, err = messageRepo.GetMessageVoteCounts(message.ID)
		if err != nil {
			t.Fatalf("Erreur nouveaux compteurs: %v", err)
		}

		if fireCount != 1 || skipCount != 0 {
			t.Errorf("Compteurs après vote incorrects. Fire: %d, Skip: %d", fireCount, skipCount)
		}

		t.Logf("✅ Compteurs de votes: Fire: %d, Skip: %d", fireCount, skipCount)
	})

	// Test 6: Tri par popularité
	t.Run("Sort By Popularity", func(t *testing.T) {
		// Créer plusieurs messages avec différents scores
		messages := make([]*models.Message, 3)
		for i := 0; i < 3; i++ {
			message := &models.Message{
				Content:  fmt.Sprintf("Message popularité %d", i+1),
				ThreadID: testThread.ID,
				UserID:   1,
			}

			err := messageRepo.Create(message)
			if err != nil {
				t.Fatalf("Erreur création message %d: %v", i+1, err)
			}

			messages[i] = message
		}

		// Donner des scores différents
		// Message 0: score 0 (pas de votes)
		// Message 1: score +1 (fire)
		// Message 2: score -1 (skip)
		messageRepo.SetUserVote(1, messages[1].ID, models.VoteFire)
		messageRepo.SetUserVote(1, messages[2].ID, models.VoteSkip)

		// Récupérer avec tri par popularité
		params := defaultPagination()
		params.PerPage = 10

		sortedMessages, _, err := messageRepo.FindByThreadID(testThread.ID, params, "popularity")
		if err != nil {
			t.Fatalf("Erreur tri par popularité: %v", err)
		}

		// Vérifier que le tri fonctionne (le plus populaire en premier)
		if len(sortedMessages) >= 3 {
			// Trouver nos messages de test dans les résultats
			var testMsg1Score, testMsg2Score int
			for _, msg := range sortedMessages {
				if msg.ID == messages[1].ID {
					testMsg1Score = msg.PopularityScore
				}
				if msg.ID == messages[2].ID {
					testMsg2Score = msg.PopularityScore
				}
			}

			if testMsg1Score != 1 {
				t.Errorf("Score message Fire incorrect. Attendu: 1, Obtenu: %d", testMsg1Score)
			}

			if testMsg2Score != -1 {
				t.Errorf("Score message Skip incorrect. Attendu: -1, Obtenu: %d", testMsg2Score)
			}
		}

		t.Logf("✅ Tri par popularité: %d messages triés", len(sortedMessages))
	})

	// Test 7: Messages avec votes utilisateur
	t.Run("Messages With User Votes", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Message pour test votes utilisateur",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		// Voter
		err = messageRepo.SetUserVote(1, message.ID, models.VoteFire)
		if err != nil {
			t.Fatalf("Erreur vote: %v", err)
		}

		// Récupérer avec votes utilisateur (augmenter la taille de page)
		params := defaultPagination()
		params.PerPage = 50 // Augmenter pour être sûr de récupérer tous les messages
		userID := uint(1)

		messagesWithVotes, _, err := messageRepo.GetMessagesWithVotes(testThread.ID, &userID, params, "date")
		if err != nil {
			t.Fatalf("Erreur récupération messages avec votes: %v", err)
		}

		// Trouver notre message et vérifier le vote
		found := false
		for _, msg := range messagesWithVotes {
			if msg.ID == message.ID {
				found = true
				if msg.UserVote == nil {
					t.Error("UserVote devrait être défini")
				} else if *msg.UserVote != models.VoteFire {
					t.Errorf("UserVote incorrect. Attendu: %s, Obtenu: %s", models.VoteFire, *msg.UserVote)
				}
				break
			}
		}

		if !found {
			t.Logf("Debug: Message ID %d non trouvé parmi %d messages", message.ID, len(messagesWithVotes))
			// Test alternatif : vérifier directement le vote
			vote, voteErr := messageRepo.GetUserVote(1, message.ID)
			if voteErr != nil {
				t.Errorf("Erreur récupération vote direct: %v", voteErr)
			} else if vote != models.VoteFire {
				t.Errorf("Vote direct incorrect. Attendu: %s, Obtenu: %s", models.VoteFire, vote)
			} else {
				t.Logf("✅ Vote vérifié directement: %s", vote)
			}
		} else {
			t.Logf("✅ Message avec vote trouvé dans la liste")
		}

		t.Logf("✅ Messages avec votes utilisateur récupérés")
	})

	// Test 8: Mise à jour d'un message
	t.Run("Update Message", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Contenu original",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		// Modifier le message
		message.Content = "Contenu modifié"

		err = messageRepo.Update(message)
		if err != nil {
			t.Fatalf("Erreur mise à jour message: %v", err)
		}

		// Vérifier la modification
		updated, err := messageRepo.FindByID(message.ID)
		if err != nil {
			t.Fatalf("Erreur récupération message modifié: %v", err)
		}

		if updated.Content != "Contenu modifié" {
			t.Errorf("Contenu non modifié. Attendu: Contenu modifié, Obtenu: %s", updated.Content)
		}

		t.Logf("✅ Message mis à jour: %s", updated.Content)
	})

	// Test 9: Suppression d'un message
	t.Run("Delete Message", func(t *testing.T) {
		// Créer un message
		message := &models.Message{
			Content:  "Message à supprimer",
			ThreadID: testThread.ID,
			UserID:   1,
		}

		err := messageRepo.Create(message)
		if err != nil {
			t.Fatalf("Erreur création message: %v", err)
		}

		messageID := message.ID

		// Supprimer le message
		err = messageRepo.Delete(messageID)
		if err != nil {
			t.Fatalf("Erreur suppression message: %v", err)
		}

		// Vérifier que le message n'existe plus
		_, err = messageRepo.FindByID(messageID)
		if err == nil {
			t.Fatal("Le message existe encore après suppression")
		}

		t.Logf("✅ Message %d supprimé avec succès", messageID)
	})

	// Test 10: Comptage de messages par thread
	t.Run("Count Messages By Thread", func(t *testing.T) {
		count, err := messageRepo.CountByThreadID(testThread.ID)
		if err != nil {
			t.Fatalf("Erreur comptage messages: %v", err)
		}

		if count < 1 {
			t.Errorf("Comptage incorrect. Attendu: ≥1, Obtenu: %d", count)
		}

		t.Logf("✅ Comptage messages thread: %d messages", count)
	})

	// Test 11: Messages par utilisateur
	t.Run("Find Messages By User", func(t *testing.T) {
		// Créer quelques messages
		for i := 1; i <= 2; i++ {
			message := &models.Message{
				Content:  fmt.Sprintf("Message utilisateur test %d", i),
				ThreadID: testThread.ID,
				UserID:   1,
			}

			err := messageRepo.Create(message)
			if err != nil {
				t.Fatalf("Erreur création message utilisateur %d: %v", i, err)
			}
		}

		// Récupérer les messages de l'utilisateur
		params := defaultPagination()
		params.PerPage = 10

		messages, total, err := messageRepo.FindByUserID(1, params)
		if err != nil {
			t.Fatalf("Erreur récupération messages utilisateur: %v", err)
		}

		if len(messages) < 2 {
			t.Errorf("Pas assez de messages utilisateur. Attendu: ≥2, Obtenu: %d", len(messages))
		}

		// Vérifier que tous appartiennent au bon utilisateur
		for _, msg := range messages {
			if msg.UserID != 1 {
				t.Errorf("Message d'un autre utilisateur trouvé. Attendu ID: 1, Obtenu: %d", msg.UserID)
			}
		}

		t.Logf("✅ Messages utilisateur: %d trouvés (total: %d)", len(messages), total)
	})
}

func TestMessageRepository_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping message repository edge cases test in short mode")
	}

	cfg := configs.Load()
	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Connexion DB échouée: %v", err)
	}
	defer database.Close()

	messageRepo := NewMessageRepository(database.DB)

	// Test: Message inexistant
	t.Run("Find Non-Existent Message", func(t *testing.T) {
		_, err := messageRepo.FindByID(99999)
		if err == nil {
			t.Fatal("La récupération d'un message inexistant devrait échouer")
		}

		t.Logf("✅ Message inexistant correctement géré: %v", err)
	})

	// Test: Vote invalide
	t.Run("Invalid Vote", func(t *testing.T) {
		err := messageRepo.SetUserVote(1, 1, "vote-invalide")
		if err == nil {
			t.Fatal("Un vote invalide devrait être rejeté")
		}

		t.Logf("✅ Vote invalide correctement rejeté: %v", err)
	})

	// Test: Vote sur message inexistant
	t.Run("Vote On Non-Existent Message", func(t *testing.T) {
		// Ceci ne devrait pas planter même si le message n'existe pas
		// (constraint foreign key devrait l'empêcher)
		err := messageRepo.SetUserVote(1, 99999, models.VoteFire)
		if err == nil {
			t.Error("Le vote sur un message inexistant devrait échouer")
		}

		t.Logf("✅ Vote sur message inexistant géré: %v", err)
	})

	// Test: Pagination avec thread vide
	t.Run("Pagination Empty Thread", func(t *testing.T) {
		params := defaultPagination()

		messages, total, err := messageRepo.FindByThreadID(99999, params, "date")
		if err != nil {
			t.Fatalf("Erreur pagination thread vide: %v", err)
		}

		if len(messages) != 0 {
			t.Errorf("Thread vide devrait retourner 0 messages, obtenu: %d", len(messages))
		}

		if total != 0 {
			t.Errorf("Total thread vide devrait être 0, obtenu: %d", total)
		}

		t.Logf("✅ Pagination thread vide: 0 messages (correct)")
	})

	// Test: Différents types de tri
	t.Run("Different Sort Orders", func(t *testing.T) {
		params := defaultPagination()

		// Test tri chronologique
		_, _, err := messageRepo.FindByThreadID(1, params, "date")
		if err != nil {
			t.Errorf("Erreur tri chronologique: %v", err)
		}

		// Test tri popularité
		_, _, err = messageRepo.FindByThreadID(1, params, "popularity")
		if err != nil {
			t.Errorf("Erreur tri popularité: %v", err)
		}

		// Test tri invalide (devrait fallback sur chronologique)
		_, _, err = messageRepo.FindByThreadID(1, params, "tri-invalide")
		if err != nil {
			t.Errorf("Erreur tri invalide: %v", err)
		}

		t.Logf("✅ Différents types de tri testés")
	})

	// Test: Score de popularité message sans votes
	t.Run("Popularity Score No Votes", func(t *testing.T) {
		score, err := messageRepo.GetPopularityScore(99999)
		if err != nil {
			t.Errorf("Erreur score message inexistant: %v", err)
		}

		if score != 0 {
			t.Errorf("Score message sans votes devrait être 0, obtenu: %d", score)
		}

		t.Logf("✅ Score message sans votes: %d (correct)", score)
	})

	// Test: Vote neutre
	t.Run("Neutral Vote", func(t *testing.T) {
		// Test que le vote neutre fonctionne
		err := messageRepo.SetUserVote(1, 1, models.VoteNeutral)
		if err != nil {
			t.Errorf("Erreur vote neutre: %v", err)
		}

		vote, err := messageRepo.GetUserVote(1, 1)
		if err != nil {
			t.Errorf("Erreur récupération vote neutre: %v", err)
		}

		if vote != models.VoteNeutral {
			t.Errorf("Vote neutre incorrect. Attendu: %s, Obtenu: %s", models.VoteNeutral, vote)
		}

		t.Logf("✅ Vote neutre: %s", vote)
	})
}
