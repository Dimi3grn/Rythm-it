package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/utils"
	"strings"
	// Import the MySQL driver for the migrate tool. Not used directly in code,
	// but needed for the driver to be registered if using migrate as a library.
	// _ "github.com/go-sql-driver/mysql"
)

// BattleRepository interface pour les opérations CRUD sur les battles de musique
type BattleRepository interface {
	Create(battle *models.Battle) error
	FindByID(id uint) (*models.Battle, error)
	FindAll(params models.PaginationParams) ([]*models.Battle, int64, error)
	FindActive(limit int) ([]*models.Battle, error)
	Update(battle *models.Battle) error
	Delete(id uint) error
	AddVote(battleID uint, userID uint, optionID uint) error // Vote pour une option (musique)
	GetVoteCounts(battleID uint) (map[uint]int, error)       // Récupère le nombre de votes par option
	GetUserVote(battleID uint, userID uint) (uint, error)    // Récupère l'option votée par l'utilisateur
}

// battleRepository implémentation concrète
type battleRepository struct {
	*BaseRepository
}

// NewBattleRepository crée une nouvelle instance du repository
func NewBattleRepository(db *sql.DB) BattleRepository {
	return &battleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create crée une nouvelle battle (de musique)
func (r *battleRepository) Create(battle *models.Battle) error {
	query := `
		INSERT INTO battles (title, description, state, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	result, err := r.DB.Exec(query,
		battle.Title,
		battle.Description,
		battle.State,
	)
	if err != nil {
		return fmt.Errorf("erreur création battle: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID battle: %w", err)
	}

	battle.ID = uint(id)

	// TODO: Ajouter l'insertion des battle_options associées si elles sont passées dans le battle struct

	return nil
}

// FindByID trouve une battle de musique par son ID avec ses options et les votes
func (r *battleRepository) FindByID(id uint) (*models.Battle, error) {
	if id == 0 {
		return nil, fmt.Errorf("ID de battle invalide")
	}

	// Query pour sélectionner la battle principale
	query := `
		SELECT id, title, description, state, created_at, updated_at
		FROM battles
		WHERE id = ?
	`

	battle := &models.Battle{}
	err := r.DB.QueryRow(query, id).Scan(
		&battle.ID, &battle.Title, &battle.Description, &battle.State, &battle.CreatedAt, &battle.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrBattleNotFound
		}
		return nil, fmt.Errorf("erreur récupération battle: %w", err)
	}

	// Récupérer les options et les votes pour cette battle
	options, err := r.getBattleOptionsWithVotes(battle.ID)
	if err != nil {
		// Si on ne peut pas récupérer les options, on retourne une erreur
		// car une battle sans options n'est pas valide
		return nil, fmt.Errorf("erreur récupération options battle: %w", err)
	}
	battle.Options = options

	// Vérifier que la battle a au moins une option
	if len(options) == 0 {
		return nil, fmt.Errorf("battle invalide: aucune option trouvée")
	}

	return battle, nil
}

// FindActive récupère les battles de musique actives avec leurs options et les votes
func (r *battleRepository) FindActive(limit int) ([]*models.Battle, error) {
	// Query pour sélectionner les battles actives
	query := `
		SELECT id, title, description, state, created_at, updated_at
		FROM battles
		WHERE state = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.DB.Query(query, models.BattleStateActive, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération battles actives: %w", err)
	}
	defer rows.Close()

	var battles []*models.Battle
	for rows.Next() {
		battle := &models.Battle{}

		// Scanner les colonnes de la table 'battles'
		err := rows.Scan(
			&battle.ID,
			&battle.Title,
			&battle.Description,
			&battle.State,
			&battle.CreatedAt,
			&battle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan battle active: %w", err)
		}

		// Pour chaque battle, récupérer ses options et les votes
		options, err := r.getBattleOptionsWithVotes(battle.ID)
		if err != nil {
			// Log the error but don't fail the whole function, return battle without options
			fmt.Printf("WARNING: Erreur récupération options et votes pour battle %d: %v\n", battle.ID, err)
		} else {
			battle.Options = options
		}

		battles = append(battles, battle)
	}

	// Vérifier les erreurs après la boucle
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur après itération sur battles actives: %w", err)
	}

	return battles, nil
}

// FindAll récupère toutes les battles de musique avec pagination, options et votes
func (r *battleRepository) FindAll(params models.PaginationParams) ([]*models.Battle, int64, error) {
	// Valider les paramètres de pagination
	models.ValidatePagination(&params)

	// Compter le total
	var total int64
	err := r.DB.QueryRow("SELECT COUNT(*) FROM battles").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage battles: %w", err)
	}

	// Si aucune battle n'existe, retourner une liste vide
	if total == 0 {
		return []*models.Battle{}, 0, nil
	}

	// Récupérer les battles avec pagination
	offset := (params.Page - 1) * params.PerPage
	query := `
		SELECT id, title, description, state, created_at, updated_at
		FROM battles
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération battles: %w", err)
	}
	defer rows.Close()

	var battles []*models.Battle
	var errors []error

	for rows.Next() {
		battle := &models.Battle{}

		// Scanner les colonnes de la table 'battles'
		err := rows.Scan(
			&battle.ID,
			&battle.Title,
			&battle.Description,
			&battle.State,
			&battle.CreatedAt,
			&battle.UpdatedAt,
		)
		if err != nil {
			errors = append(errors, fmt.Errorf("erreur scan battle ID %d: %w", battle.ID, err))
			continue
		}

		// Pour chaque battle, récupérer ses options et les votes
		options, err := r.getBattleOptionsWithVotes(battle.ID)
		if err != nil {
			errors = append(errors, fmt.Errorf("erreur récupération options battle ID %d: %w", battle.ID, err))
			continue
		}

		// Vérifier que la battle a au moins une option
		if len(options) == 0 {
			errors = append(errors, fmt.Errorf("battle ID %d invalide: aucune option trouvée", battle.ID))
			continue
		}

		battle.Options = options
		battles = append(battles, battle)
	}

	// Vérifier les erreurs après la boucle
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("erreur après itération sur battles: %w", err)
	}

	// Si on a des erreurs individuelles, les logger mais continuer
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("WARNING: %v\n", err)
		}
	}

	// Si aucune battle valide n'a été trouvée à cause des erreurs
	if len(battles) == 0 && len(errors) > 0 {
		return nil, 0, fmt.Errorf("aucune battle valide trouvée: %v", errors[0])
	}

	return battles, total, nil
}

// Update met à jour une battle de musique
func (r *battleRepository) Update(battle *models.Battle) error {
	query := `
		UPDATE battles
		SET title = ?, description = ?, state = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.DB.Exec(query, battle.Title, battle.Description, battle.State, battle.ID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour battle: %w", err)
	}

	// TODO: Ajouter la logique pour mettre à jour les battle_options associées si nécessaire

	return nil
}

// Delete supprime une battle de musique
func (r *battleRepository) Delete(id uint) error {
	// Grâce aux contraintes ON DELETE CASCADE dans la DB, les options et votes associés seront aussi supprimés.
	query := "DELETE FROM battles WHERE id = ?"

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression battle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("battle ID %d non trouvée pour suppression", id)
	}

	return nil
}

// AddVote ajoute ou met à jour un vote d'utilisateur pour une option de battle
func (r *battleRepository) AddVote(battleID uint, userID uint, optionID uint) error {
	// Vérifier que l'option existe et fait partie de la battle
	var exists bool
	err := r.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM battle_options
			WHERE id = ? AND battle_id = ?
		)`, optionID, battleID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("erreur vérification option %d pour battle %d: %w", optionID, battleID, err)
	}
	if !exists {
		return fmt.Errorf("option ID %d ne fait pas partie de la battle %d", optionID, battleID)
	}

	// Supprimer l'ancien vote de l'utilisateur pour cette battle s'il existe
	_, err = r.DB.Exec("DELETE FROM battle_votes WHERE battle_id = ? AND user_id = ?", battleID, userID)
	if err != nil {
		return fmt.Errorf("erreur suppression ancien vote: %w", err)
	}

	// Ajouter le nouveau vote
	_, err = r.DB.Exec(`
		INSERT INTO battle_votes (battle_id, user_id, option_id, created_at)
		VALUES (?, ?, ?, NOW())`, battleID, userID, optionID)
	if err != nil {
		return fmt.Errorf("erreur ajout vote: %w", err)
	}

	return nil
}

// GetVoteCounts récupère le nombre de votes pour chaque option d'une battle
func (r *battleRepository) GetVoteCounts(battleID uint) (map[uint]int, error) {
	query := `
		SELECT option_id, COUNT(*) as vote_count
		FROM battle_votes
		WHERE battle_id = ?
		GROUP BY option_id
	`

	rows, err := r.DB.Query(query, battleID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération votes: %w", err)
	}
	defer rows.Close()

	votes := make(map[uint]int)

	for rows.Next() {
		var optionID uint
		var count int
		err := rows.Scan(&optionID, &count)
		if err != nil {
			return nil, fmt.Errorf("erreur scan vote count: %w", err)
		}
		votes[optionID] = count
	}

	// Vérifier les erreurs après la boucle
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur après itération sur vote counts: %w", err)
	}

	return votes, nil
}

// GetUserVote récupère l'option (musique) votée par un utilisateur pour une battle
func (r *battleRepository) GetUserVote(battleID uint, userID uint) (uint, error) {
	var optionID uint
	err := r.DB.QueryRow(`
		SELECT option_id
		FROM battle_votes
		WHERE battle_id = ? AND user_id = ?`, battleID, userID).Scan(&optionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Pas de vote
		}
		return 0, fmt.Errorf("erreur récupération vote utilisateur: %w", err)
	}

	return optionID, nil
}

// getBattleOptionsWithVotes récupère les options pour une battle donnée et calcule leurs votes
// Cette est une nouvelle méthode interne pour aider FindActive, FindByID, et FindAll
func (r *battleRepository) getBattleOptionsWithVotes(battleID uint) ([]*models.BattleOption, error) {
	// Query pour sélectionner les options pour une battle
	optionsQuery := `
		SELECT id, battle_id, title, artist, music_url, image_url
		FROM battle_options
		WHERE battle_id = ?
	`
	optionsRows, err := r.DB.Query(optionsQuery, battleID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération options pour battle %d: %w", battleID, err)
	}
	defer optionsRows.Close()

	var options []*models.BattleOption
	optionIDs := []uint{}
	optionsMap := make(map[uint]*models.BattleOption)

	for optionsRows.Next() {
		option := &models.BattleOption{}
		err := optionsRows.Scan(
			&option.ID,
			&option.BattleID,
			&option.Title,
			&option.Artist,
			&option.MusicURL,
			&option.ImageURL,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan option pour battle %d: %w", battleID, err)
		}
		options = append(options, option)
		optionIDs = append(optionIDs, option.ID)
		optionsMap[option.ID] = option
	}

	// Vérifier les erreurs après la boucle des options
	if err = optionsRows.Err(); err != nil {
		return nil, fmt.Errorf("erreur après itération sur options pour battle %d: %w", battleID, err)
	}

	// Si aucune option, retourner liste vide
	if len(options) == 0 {
		return options, nil
	}

	// Récupérer les comptes de votes pour ces options
	// Utiliser une requête groupée pour compter les votes par option_id
	// Utiliser une clause IN pour filtrer par optionIDs
	// Construire dynamiquement la clause IN pour gérer un nombre variable d'options
	places := strings.Repeat("?, ", len(optionIDs))
	places = places[:len(places)-2] // Remove trailing ", "
	voteCountsQuery := fmt.Sprintf(`
		SELECT option_id, COUNT(*) as vote_count
		FROM battle_votes
		WHERE battle_id = ? AND option_id IN (%s)
		GROUP BY option_id
	`, places)

	// Préparer les arguments pour la clause IN
	voteQueryArgs := []interface{}{battleID}
	for _, id := range optionIDs {
		voteQueryArgs = append(voteQueryArgs, id)
	}

	voteCountsRows, err := r.DB.Query(voteCountsQuery, voteQueryArgs...)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération votes pour options de battle %d: %w", battleID, err)
	}
	defer voteCountsRows.Close()

	// Mettre à jour le VoteCount pour chaque option dans optionsMap
	for voteCountsRows.Next() {
		var optionID uint
		var count int
		err := voteCountsRows.Scan(&optionID, &count)
		if err != nil {
			return nil, fmt.Errorf("erreur scan vote count pour battle %d: %w", battleID, err)
		}
		if option, ok := optionsMap[optionID]; ok {
			option.VoteCount = count
		}
	}

	// Vérifier les erreurs après la boucle des votes
	if err = voteCountsRows.Err(); err != nil {
		return nil, fmt.Errorf("erreur après itération sur vote counts pour battle %d: %w", battleID, err)
	}

	// Retourner les options avec les VoteCounts mis à jour
	return options, nil
}
