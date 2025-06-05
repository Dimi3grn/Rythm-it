package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"strings"
)

// TagRepository interface pour les opérations CRUD sur les tags
type TagRepository interface {
	Create(tag *models.Tag) error
	FindByID(id uint) (*models.Tag, error)
	FindByName(name string) (*models.Tag, error)
	FindOrCreate(name string, tagType string) (*models.Tag, error)
	FindAll() ([]*models.Tag, error)
	FindByType(tagType string) ([]*models.Tag, error)
	GetPopularTags(limit int) ([]*models.Tag, error)
	SearchTags(query string, tagType string, limit int) ([]*models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id uint) error
	GetTagUsageCount(tagID uint) (int64, error)
}

// tagRepository implémentation concrète
type tagRepository struct {
	*BaseRepository
}

// NewTagRepository crée une nouvelle instance du repository
func NewTagRepository(db *sql.DB) TagRepository {
	return &tagRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create crée un nouveau tag
func (r *tagRepository) Create(tag *models.Tag) error {
	// Normaliser le nom (lowercase, trim)
	tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))

	// Vérifier que le tag n'existe pas déjà
	existing, err := r.FindByName(tag.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("tag '%s' existe déjà", tag.Name)
	}

	query := "INSERT INTO tags (name, type, created_at) VALUES (?, ?, NOW())"

	result, err := r.DB.Exec(query, tag.Name, tag.Type)
	if err != nil {
		return fmt.Errorf("erreur création tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID tag: %w", err)
	}

	tag.ID = uint(id)
	return nil
}

// FindByID trouve un tag par son ID
func (r *tagRepository) FindByID(id uint) (*models.Tag, error) {
	query := "SELECT id, name, type, created_at FROM tags WHERE id = ?"

	tag := &models.Tag{}
	var createdAt sql.NullTime

	err := r.DB.QueryRow(query, id).Scan(
		&tag.ID, &tag.Name, &tag.Type, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag ID %d non trouvé", id)
		}
		return nil, fmt.Errorf("erreur récupération tag: %w", err)
	}

	return tag, nil
}

// FindByName trouve un tag par son nom (case insensitive)
func (r *tagRepository) FindByName(name string) (*models.Tag, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	query := "SELECT id, name, type, created_at FROM tags WHERE LOWER(name) = ?"

	tag := &models.Tag{}
	var createdAt sql.NullTime

	err := r.DB.QueryRow(query, normalizedName).Scan(
		&tag.ID, &tag.Name, &tag.Type, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag '%s' non trouvé", name)
		}
		return nil, fmt.Errorf("erreur récupération tag par nom: %w", err)
	}

	return tag, nil
}

// FindOrCreate trouve un tag ou le crée s'il n'existe pas
func (r *tagRepository) FindOrCreate(name string, tagType string) (*models.Tag, error) {
	// D'abord essayer de le trouver
	tag, err := r.FindByName(name)
	if err == nil {
		return tag, nil
	}

	// S'il n'existe pas, le créer
	newTag := &models.Tag{
		Name: strings.ToLower(strings.TrimSpace(name)),
		Type: tagType,
	}

	err = r.Create(newTag)
	if err != nil {
		return nil, fmt.Errorf("erreur création tag: %w", err)
	}

	return newTag, nil
}

// FindAll récupère tous les tags
func (r *tagRepository) FindAll() ([]*models.Tag, error) {
	query := "SELECT id, name, type, created_at FROM tags ORDER BY name ASC"

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		var createdAt sql.NullTime

		err := rows.Scan(&tag.ID, &tag.Name, &tag.Type, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("erreur scan tag: %w", err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// FindByType récupère les tags par type (genre, artist, album)
func (r *tagRepository) FindByType(tagType string) ([]*models.Tag, error) {
	query := "SELECT id, name, type, created_at FROM tags WHERE type = ? ORDER BY name ASC"

	rows, err := r.DB.Query(query, tagType)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags par type: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		var createdAt sql.NullTime

		err := rows.Scan(&tag.ID, &tag.Name, &tag.Type, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("erreur scan tag: %w", err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// GetPopularTags récupère les tags les plus utilisés
func (r *tagRepository) GetPopularTags(limit int) ([]*models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.type, t.created_at, COUNT(tt.thread_id) as usage_count
		FROM tags t
		LEFT JOIN thread_tags tt ON t.id = tt.tag_id
		GROUP BY t.id, t.name, t.type, t.created_at
		ORDER BY usage_count DESC, t.name ASC
		LIMIT ?
	`

	rows, err := r.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags populaires: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		var createdAt sql.NullTime
		var usageCount int

		err := rows.Scan(&tag.ID, &tag.Name, &tag.Type, &createdAt, &usageCount)
		if err != nil {
			return nil, fmt.Errorf("erreur scan tag populaire: %w", err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// SearchTags recherche des tags par nom (pour auto-complétion)
func (r *tagRepository) SearchTags(query string, tagType string, limit int) ([]*models.Tag, error) {
	searchTerm := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"

	var sqlQuery string
	var args []interface{}

	if tagType != "" {
		sqlQuery = `
			SELECT id, name, type, created_at 
			FROM tags 
			WHERE LOWER(name) LIKE ? AND type = ?
			ORDER BY 
				CASE WHEN LOWER(name) = LOWER(?) THEN 0 ELSE 1 END,
				LENGTH(name),
				name ASC 
			LIMIT ?
		`
		args = []interface{}{searchTerm, tagType, strings.ToLower(strings.TrimSpace(query)), limit}
	} else {
		sqlQuery = `
			SELECT id, name, type, created_at 
			FROM tags 
			WHERE LOWER(name) LIKE ?
			ORDER BY 
				CASE WHEN LOWER(name) = LOWER(?) THEN 0 ELSE 1 END,
				LENGTH(name),
				name ASC 
			LIMIT ?
		`
		args = []interface{}{searchTerm, strings.ToLower(strings.TrimSpace(query)), limit}
	}

	rows, err := r.DB.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur recherche tags: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		var createdAt sql.NullTime

		err := rows.Scan(&tag.ID, &tag.Name, &tag.Type, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("erreur scan tag recherche: %w", err)
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// Update met à jour un tag
func (r *tagRepository) Update(tag *models.Tag) error {
	// Normaliser le nom
	tag.Name = strings.ToLower(strings.TrimSpace(tag.Name))

	query := "UPDATE tags SET name = ?, type = ? WHERE id = ?"

	result, err := r.DB.Exec(query, tag.Name, tag.Type, tag.ID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour tag: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("tag ID %d non trouvé pour mise à jour", tag.ID)
	}

	return nil
}

// Delete supprime un tag (seulement s'il n'est pas utilisé)
func (r *tagRepository) Delete(id uint) error {
	// Vérifier d'abord si le tag est utilisé
	usageCount, err := r.GetTagUsageCount(id)
	if err != nil {
		return fmt.Errorf("erreur vérification usage tag: %w", err)
	}

	if usageCount > 0 {
		return fmt.Errorf("impossible de supprimer le tag: utilisé dans %d thread(s)", usageCount)
	}

	query := "DELETE FROM tags WHERE id = ?"

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression tag: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("tag ID %d non trouvé pour suppression", id)
	}

	return nil
}

// GetTagUsageCount compte le nombre de threads utilisant ce tag
func (r *tagRepository) GetTagUsageCount(tagID uint) (int64, error) {
	query := "SELECT COUNT(*) FROM thread_tags WHERE tag_id = ?"

	var count int64
	err := r.DB.QueryRow(query, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage usage tag: %w", err)
	}

	return count, nil
}
