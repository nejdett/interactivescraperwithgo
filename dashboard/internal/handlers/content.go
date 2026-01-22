package handlers

import (
	"net/http"
	"strconv"

	"github.com/cti-dashboard/dashboard/internal/models"
	"github.com/cti-dashboard/dashboard/internal/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ContentHandler handles content-related requests
type ContentHandler struct {
	contentRepo *repository.ContentRepository
}

// NewContentHandler creates a new ContentHandler
func NewContentHandler(contentRepo *repository.ContentRepository) *ContentHandler {
	return &ContentHandler{
		contentRepo: contentRepo,
	}
}

func (h *ContentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	sortBy := c.DefaultQuery("sort_by", "published_at")
	order := c.DefaultQuery("order", "desc")
	category := c.Query("category")

	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 50
	}

	params := repository.ListParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		Order:    order,
		Category: category,
	}

	items, total, err := h.contentRepo.List(params)
	if err != nil {
		log.WithError(err).Error("Failed to list content items")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve content items",
		})
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := models.ContentListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetByID handles GET /api/contents/:id - get content item by ID
func (h *ContentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	item, err := h.contentRepo.GetByID(id)
	if err != nil {
		if err.Error() == "content not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Content item not found",
			})
			return
		}

		log.WithError(err).WithField("id", id).Error("Failed to get content item")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve content item",
		})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ContentHandler) GetStats(c *gin.Context) {
	categoryDist, err := h.contentRepo.GetCategoryDistribution()
	if err != nil {
		log.WithError(err).Error("Failed to get category distribution")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve statistics",
		})
		return
	}

	criticalityDist, err := h.contentRepo.GetCriticalityDistribution()
	if err != nil {
		log.WithError(err).Error("Failed to get criticality distribution")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve statistics",
		})
		return
	}

	totalItems, err := h.contentRepo.GetTotalCount()
	if err != nil {
		log.WithError(err).Error("Failed to get total count")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve statistics",
		})
		return
	}

	lastUpdated, err := h.contentRepo.GetLastUpdated()
	if err != nil {
		log.WithError(err).Error("Failed to get last updated")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve statistics",
		})
		return
	}

	response := models.StatsResponse{
		CategoryDistribution:    categoryDist,
		CriticalityDistribution: criticalityDist,
		TotalItems:              totalItems,
		LastUpdated:             lastUpdated,
	}

	c.JSON(http.StatusOK, response)
}
