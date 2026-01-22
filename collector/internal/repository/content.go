package repository

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type ContentItem struct {
	ID               string
	Title            string
	SourceName       string
	SourceURL        string
	Content          string
	PublishedAt      time.Time
	CriticalityScore int
	CollectedAt      time.Time
	Categories       []string
}

type ContentRepository struct {
	db *sql.DB
}

func NewContentRepository(db *sql.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func (cr *ContentRepository) Insert(item *ContentItem) error {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := cr.insertWithTransaction(item)
		if err == nil {
			return nil
		}

		lastErr = err
		log.WithFields(log.Fields{
			"attempt": attempt,
			"error":   err.Error(),
		}).Warn("Failed to insert, retrying...")

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
	}

	return fmt.Errorf("insert failed after %d attempts: %w", maxRetries, lastErr)
}

func (cr *ContentRepository) insertWithTransaction(item *ContentItem) error {
	tx, err := cr.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var contentID string
	query := `
		INSERT INTO content_items (title, source_name, source_url, content, published_at, criticality_score, collected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err = tx.QueryRow(
		query,
		item.Title,
		item.SourceName,
		item.SourceURL,
		item.Content,
		item.PublishedAt,
		item.CriticalityScore,
		time.Now(),
	).Scan(&contentID)

	if err != nil {
		return err
	}

	if len(item.Categories) > 0 {
		err = cr.assignCategories(tx, contentID, item.Categories)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"content_id":        contentID,
		"title":             item.Title,
		"source":            item.SourceName,
		"criticality_score": item.CriticalityScore,
		"categories":        item.Categories,
	}).Info("Content item inserted successfully")

	return nil
}

func (cr *ContentRepository) ExistsByURL(url string) (bool, error) {
	var exists bool
	err := cr.db.QueryRow("SELECT EXISTS(SELECT 1 FROM content_items WHERE source_url = $1)", url).Scan(&exists)
	return exists, err
}

func (cr *ContentRepository) assignCategories(tx *sql.Tx, contentID string, categoryNames []string) error {
	for _, name := range categoryNames {
		var catID string
		err := tx.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&catID)
		if err != nil {
			if err == sql.ErrNoRows {
				log.WithField("category", name).Warn("Category not found")
				continue
			}
			return err
		}

		_, err = tx.Exec(
			"INSERT INTO content_categories (content_id, category_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			contentID, catID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
