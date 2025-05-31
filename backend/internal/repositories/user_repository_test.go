package repositories

import (
	"testing"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/pkg/database"
)

func setupTestDB(t *testing.T) UserRepository {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	cfg := configs.Load()

	err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Impossible de se connecter à la base de données: %v", err)
	}

	return NewUserRepository(database.DB)
}

func teardownTestDB() {
	database.Close()
}

func createTestUser() *models.User {
	return &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword123",
		IsAdmin:  false,
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()

	// Test création
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}

	// Vérifier que l'ID a été généré
	if user.ID == 0 {
		t.Error("L'ID utilisateur n'a pas été généré")
	}

	// Vérifier que les timestamps ont été définis
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt n'a pas été défini")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt n'a pas été défini")
	}

	// Nettoyer
	repo.Delete(user.ID)
}

func TestUserRepository_FindByID(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	// Créer un utilisateur de test
	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test FindByID
	foundUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Erreur recherche utilisateur: %v", err)
	}

	// Vérifications
	if foundUser.Username != user.Username {
		t.Errorf("Username attendu: %s, obtenu: %s", user.Username, foundUser.Username)
	}

	if foundUser.Email != user.Email {
		t.Errorf("Email attendu: %s, obtenu: %s", user.Email, foundUser.Email)
	}

	// Test utilisateur inexistant
	_, err = repo.FindByID(99999)
	if err == nil {
		t.Error("Devrait retourner une erreur pour un ID inexistant")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test FindByEmail
	foundUser, err := repo.FindByEmail(user.Email)
	if err != nil {
		t.Fatalf("Erreur recherche par email: %v", err)
	}

	if foundUser.ID != user.ID {
		t.Errorf("ID attendu: %d, obtenu: %d", user.ID, foundUser.ID)
	}

	// Test email inexistant
	_, err = repo.FindByEmail("inexistant@example.com")
	if err == nil {
		t.Error("Devrait retourner une erreur pour un email inexistant")
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test FindByUsername
	foundUser, err := repo.FindByUsername(user.Username)
	if err != nil {
		t.Fatalf("Erreur recherche par username: %v", err)
	}

	if foundUser.ID != user.ID {
		t.Errorf("ID attendu: %d, obtenu: %d", user.ID, foundUser.ID)
	}

	// Test username inexistant
	_, err = repo.FindByUsername("inexistant")
	if err == nil {
		t.Error("Devrait retourner une erreur pour un username inexistant")
	}
}

func TestUserRepository_Update(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Modifier l'utilisateur
	originalUpdatedAt := user.UpdatedAt
	time.Sleep(time.Millisecond * 10) // Pour s'assurer que updated_at change

	user.Username = "updateduser"
	user.Email = "updated@example.com"

	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Erreur mise à jour utilisateur: %v", err)
	}

	// Vérifier la mise à jour
	updatedUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Erreur récupération utilisateur mis à jour: %v", err)
	}

	if updatedUser.Username != "updateduser" {
		t.Errorf("Username non mis à jour: %s", updatedUser.Username)
	}

	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Email non mis à jour: %s", updatedUser.Email)
	}

	if !updatedUser.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt n'a pas été mis à jour")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}

	// Test suppression
	err = repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Erreur suppression utilisateur: %v", err)
	}

	// Vérifier que l'utilisateur n'existe plus
	_, err = repo.FindByID(user.ID)
	if err == nil {
		t.Error("L'utilisateur devrait être supprimé")
	}

	// Test suppression utilisateur inexistant
	err = repo.Delete(99999)
	if err == nil {
		t.Error("Devrait retourner une erreur pour un ID inexistant")
	}
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test email existant
	exists, err := repo.ExistsByEmail(user.Email)
	if err != nil {
		t.Fatalf("Erreur vérification existence email: %v", err)
	}

	if !exists {
		t.Error("L'email devrait exister")
	}

	// Test email inexistant
	exists, err = repo.ExistsByEmail("inexistant@example.com")
	if err != nil {
		t.Fatalf("Erreur vérification email inexistant: %v", err)
	}

	if exists {
		t.Error("L'email ne devrait pas exister")
	}
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test username existant
	exists, err := repo.ExistsByUsername(user.Username)
	if err != nil {
		t.Fatalf("Erreur vérification existence username: %v", err)
	}

	if !exists {
		t.Error("Le username devrait exister")
	}

	// Test username inexistant
	exists, err = repo.ExistsByUsername("inexistant")
	if err != nil {
		t.Fatalf("Erreur vérification username inexistant: %v", err)
	}

	if exists {
		t.Error("Le username ne devrait pas exister")
	}
}

func TestUserRepository_UpdateLastConnection(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Mettre à jour la dernière connexion
	err = repo.UpdateLastConnection(user.ID)
	if err != nil {
		t.Fatalf("Erreur mise à jour dernière connexion: %v", err)
	}

	// Vérifier la mise à jour
	updatedUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Erreur récupération utilisateur: %v", err)
	}

	if updatedUser.LastConnection == nil {
		t.Error("LastConnection devrait être définie")
	}

	// Test ID inexistant
	err = repo.UpdateLastConnection(99999)
	if err == nil {
		t.Error("Devrait retourner une erreur pour un ID inexistant")
	}
}

func TestUserRepository_IncrementCounters(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Erreur création utilisateur: %v", err)
	}
	defer repo.Delete(user.ID)

	// Test incrémentation compteur messages
	err = repo.IncrementMessageCount(user.ID)
	if err != nil {
		t.Fatalf("Erreur incrémentation messages: %v", err)
	}

	// Test incrémentation compteur threads
	err = repo.IncrementThreadCount(user.ID)
	if err != nil {
		t.Fatalf("Erreur incrémentation threads: %v", err)
	}

	// Vérifier les compteurs
	updatedUser, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("Erreur récupération utilisateur: %v", err)
	}

	if updatedUser.MessageCount != 1 {
		t.Errorf("MessageCount attendu: 1, obtenu: %d", updatedUser.MessageCount)
	}

	if updatedUser.ThreadCount != 1 {
		t.Errorf("ThreadCount attendu: 1, obtenu: %d", updatedUser.ThreadCount)
	}
}

func TestUserRepository_UniqueConstraints(t *testing.T) {
	repo := setupTestDB(t)
	defer teardownTestDB()

	user1 := createTestUser()
	err := repo.Create(user1)
	if err != nil {
		t.Fatalf("Erreur création premier utilisateur: %v", err)
	}
	defer repo.Delete(user1.ID)

	// Tenter de créer un utilisateur avec le même email
	user2 := &models.User{
		Username: "differentuser",
		Email:    user1.Email, // Même email
		Password: "hashedpassword123",
	}

	err = repo.Create(user2)
	if err == nil {
		defer repo.Delete(user2.ID)
		t.Error("Devrait échouer avec un email dupliqué")
	}

	// Tenter de créer un utilisateur avec le même username
	user3 := &models.User{
		Username: user1.Username, // Même username
		Email:    "different@example.com",
		Password: "hashedpassword123",
	}

	err = repo.Create(user3)
	if err == nil {
		defer repo.Delete(user3.ID)
		t.Error("Devrait échouer avec un username dupliqué")
	}
}
