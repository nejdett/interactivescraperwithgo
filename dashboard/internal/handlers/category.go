package handlers

import (
	"net/http"

	"github.com/cti-dashboard/dashboard/internal/models"
	"github.com/cti-dashboard/dashboard/internal/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CategoryHandler handles category-related requests
type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
	}
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name               string `json:"name" binding:"required"`
	Description        string `json:"description"`
	DefaultCriticality int    `json:"default_criticality" binding:"required,min=1,max=10"`
	Color              string `json:"color" binding:"required"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Name               string `json:"name" binding:"required"`
	Description        string `json:"description"`
	DefaultCriticality int    `json:"default_criticality" binding:"required,min=1,max=10"`
	Color              string `json:"color" binding:"required"`
}

// List handles GET /api/categories - list all categories
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.categoryRepo.List()
	if err != nil {
		log.WithError(err).Error("Failed to list categories")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// Create handles POST /api/categories - create a new category
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	category := &models.Category{
		Name:               req.Name,
		Description:        req.Description,
		DefaultCriticality: req.DefaultCriticality,
		Color:              req.Color,
	}

	if err := h.categoryRepo.Create(category); err != nil {
		log.WithError(err).Error("Failed to create category")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create category",
		})
		return
	}

	log.WithField("category", category.Name).Info("Category created")
	c.JSON(http.StatusCreated, category)
}

// Update handles PUT /api/categories/:id - update a category
func (h *CategoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	category := &models.Category{
		ID:                 id,
		Name:               req.Name,
		Description:        req.Description,
		DefaultCriticality: req.DefaultCriticality,
		Color:              req.Color,
	}

	if err := h.categoryRepo.Update(category); err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Category not found",
			})
			return
		}

		log.WithError(err).Error("Failed to update category")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to update category",
		})
		return
	}

	log.WithField("category", category.Name).Info("Category updated")
	c.JSON(http.StatusOK, category)
}

// Delete handles DELETE /api/categories/:id - delete a category
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.categoryRepo.Delete(id); err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Category not found",
			})
			return
		}

		log.WithError(err).Error("Failed to delete category")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to delete category",
		})
		return
	}

	log.WithField("id", id).Info("Category deleted")
	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
