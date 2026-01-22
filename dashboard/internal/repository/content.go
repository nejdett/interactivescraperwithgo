package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cti-dashboard/dashboard/internal/models"
)

// ContentRepository handles database operations for content items
type ContentRepository struct {
	db *sql.DB
}

// NewContentRepository creates a new ContentRepository
func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

// ListParams holds parameters for listing content items
type ListParams struct {
	Page     int
	PageSize int
	SortBy   string
	Order    string
	Category string
}

func (r *ContentRepository) List(params ListParams) ([]models.ContentItem, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 50
	}
	if params.SortBy == "" {
		params.SortBy = "published_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}

	validSortFields := map[string]bool{
		"published_at":      true,
		"criticality_score": true,
		"created_at":        true,
	}
	if !validSortFields[params.SortBy] {
		params.SortBy = "published_at"
	}

	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	var whereClause string
	var args []interface{}
	argIndex := 1

	if params.Category != "" {
		whereClause = `
			WHERE ci.id IN (
				SELECT cc.content_id 
				FROM content_categories cc
				JOIN categories c ON cc.category_id = c.id
				WHERE c.name = $` + fmt.Sprintf("%d", argIndex) + `
			)
		`
		args = append(args, params.Category)
		argIndex++
	}

	countQuery := `SELECT COUNT(*) FROM content_items ci ` + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	query := `
		SELECT DISTINCT ci.id, ci.title, ci.source_name, ci.source_url, 
		       ci.content, ci.published_at, ci.criticality_score, 
		       ci.collected_at, ci.created_at
		FROM content_items ci
		` + whereClause + `
		ORDER BY ci.` + params.SortBy + ` ` + strings.ToUpper(params.Order) + `
		LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ContentItem
	for rows.Next() {
		var item models.ContentItem
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.SourceName,
			&item.SourceURL,
			&item.Content,
			&item.PublishedAt,
			&item.CriticalityScore,
			&item.CollectedAt,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		categories, err := r.getCategoriesForContent(item.ID)
		if err != nil {
			return nil, 0, err
		}
		item.Categories = categories

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// getCategoriesForContent retrieves all categories for a content item
func (r *ContentRepository) getCategoriesForContent(contentID string) ([]models.Category, error) {
	query := `
		SELECT c.id, c.name, c.description, c.default_criticality, c.color, c.created_at, c.updated_at
		FROM categories c
		JOIN content_categories cc ON c.id = cc.category_id
		WHERE cc.content_id = $1
		ORDER BY c.name
	`

	rows, err := r.db.Query(query, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Description,
			&cat.DefaultCriticality,
			&cat.Color,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

func (r *ContentRepository) GetByID(id string) (*models.ContentItem, error) {
	query := `
		SELECT id, title, source_name, source_url, content, 
		       published_at, criticality_score, collected_at, created_at
		FROM content_items
		WHERE id = $1
	`

	var item models.ContentItem
	err := r.db.QueryRow(query, id).Scan(
		&item.ID,
		&item.Title,
		&item.SourceName,
		&item.SourceURL,
		&item.Content,
		&item.PublishedAt,
		&item.CriticalityScore,
		&item.CollectedAt,
		&item.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("content not found")
	}
	if err != nil {
		return nil, err
	}

	categories, err := r.getCategoriesForContent(item.ID)
	if err != nil {
		return nil, err
	}
	item.Categories = categories

	return &item, nil
}

func (r *ContentRepository) GetCategoryDistribution() (map[string]int, error) {
	query := `
		SELECT c.name, COUNT(cc.content_id) as count
		FROM categories c
		LEFT JOIN content_categories cc ON c.id = cc.category_id
		GROUP BY c.name
		ORDER BY count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		dist[name] = count
	}

	return dist, rows.Err()
}

func (r *ContentRepository) GetCriticalityDistribution() (map[string]int, error) {
	query := `
		SELECT 
			CASE 
				WHEN criticality_score BETWEEN 1 AND 3 THEN '1-3'
				WHEN criticality_score BETWEEN 4 AND 6 THEN '4-6'
				WHEN criticality_score BETWEEN 7 AND 8 THEN '7-8'
				WHEN criticality_score BETWEEN 9 AND 10 THEN '9-10'
			END as range,
			COUNT(*) as count
		FROM content_items
		WHERE criticality_score IS NOT NULL
		GROUP BY range
		ORDER BY range
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var r string
		var count int
		if err := rows.Scan(&r, &count); err != nil {
			return nil, err
		}
		dist[r] = count
	}

	return dist, rows.Err()
}

func (r *ContentRepository) GetTotalCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM content_items").Scan(&count)
	return count, err
}

func (r *ContentRepository) GetLastUpdated() (string, error) {
	var lastUpdated sql.NullTime
	err := r.db.QueryRow("SELECT MAX(collected_at) FROM content_items").Scan(&lastUpdated)
	if err != nil {
		return "", err
	}
	
	if !lastUpdated.Valid {
		return "", nil
	}
	
	return lastUpdated.Time.Format("2006-01-02T15:04:05Z07:00"), nil
}
