package repository

import (
	"database/sql"
	"fmt"

	"github.com/cti-dashboard/dashboard/internal/models"
)

// CategoryRepository handles database operations for categories
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// List retrieves all categories
func (r *CategoryRepository) List() ([]models.Category, error) {
	query := `
		SELECT id, name, description, default_criticality, color, created_at, updated_at
		FROM categories
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
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
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// Create creates a new category
func (r *CategoryRepository) Create(cat *models.Category) error {
	query := `
		INSERT INTO categories (name, description, default_criticality, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		cat.Name,
		cat.Description,
		cat.DefaultCriticality,
		cat.Color,
	).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// Update updates an existing category
func (r *CategoryRepository) Update(cat *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, default_criticality = $3, color = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		cat.Name,
		cat.Description,
		cat.DefaultCriticality,
		cat.Color,
		cat.ID,
	).Scan(&cat.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("category not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// Delete deletes a category by ID
func (r *CategoryRepository) Delete(id string) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(id string) (*models.Category, error) {
	query := `
		SELECT id, name, description, default_criticality, color, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	var cat models.Category
	err := r.db.QueryRow(query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Description,
		&cat.DefaultCriticality,
		&cat.Color,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &cat, nil
}
