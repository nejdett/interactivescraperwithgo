package models

import "time"

// ContentItem represents a content item collected from Dark Web sources
type ContentItem struct {
	ID               string     `json:"id" db:"id"`
	Title            string     `json:"title" db:"title"`
	SourceName       string     `json:"source_name" db:"source_name"`
	SourceURL        string     `json:"source_url" db:"source_url"`
	Content          string     `json:"content" db:"content"`
	PublishedAt      *time.Time `json:"published_at" db:"published_at"`
	CriticalityScore int        `json:"criticality_score" db:"criticality_score"`
	CollectedAt      time.Time  `json:"collected_at" db:"collected_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	Categories       []Category `json:"categories,omitempty"`
}

// Category represents a content category
type Category struct {
	ID                 string    `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	DefaultCriticality int       `json:"default_criticality" db:"default_criticality"`
	Color              string    `json:"color" db:"color"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// ContentListResponse represents the response for content list endpoint
type ContentListResponse struct {
	Items      []ContentItem `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// StatsResponse represents the response for statistics endpoint
type StatsResponse struct {
	CategoryDistribution    map[string]int `json:"category_distribution"`
	CriticalityDistribution map[string]int `json:"criticality_distribution"`
	TotalItems              int            `json:"total_items"`
	LastUpdated             string         `json:"last_updated"`
}
