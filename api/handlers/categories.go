package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
)

// GetCategories retrieves all categories (system defaults + user custom)
func GetCategories(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	// Get all categories - system defaults (user_id IS NULL) + user custom categories
	rows, err := db.DB.Query(`
		SELECT id, user_id, name, color, is_default, created_at, updated_at
		FROM categories
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY is_default DESC, name ASC
	`, userID)
	if err != nil {
		logger.Log.Errorw("Failed to get categories", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve categories",
		))
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Color, &cat.IsDefault, &cat.CreatedAt, &cat.UpdatedAt)
		if err != nil {
			logger.Log.Errorw("Failed to scan category", "error", err)
			continue
		}
		categories = append(categories, cat)
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(categories))
}

// CreateCategory creates a new custom category for the user
func CreateCategory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Create category
	var category models.Category
	now := time.Now()
	err = db.DB.QueryRow(`
		INSERT INTO categories (user_id, name, color, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, color, is_default, created_at, updated_at
	`, userID, req.Name, req.Color, false, now, now).Scan(
		&category.ID, &category.UserID, &category.Name, &category.Color, &category.IsDefault, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		logger.Log.Errorw("Failed to create category", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to create category",
		))
		return
	}

	logger.Log.Infow("Category created", "categoryId", category.ID, "userId", userID, "name", req.Name)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(category))
}

// UpdateCategory updates an existing custom category
func UpdateCategory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Invalid category ID",
		))
		return
	}

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Check if category exists and belongs to user (not a default category)
	var isDefault bool
	var ownerID *uuid.UUID
	err = db.DB.QueryRow("SELECT is_default, user_id FROM categories WHERE id = $1", categoryID).Scan(&isDefault, &ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Category not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get category", "error", err, "categoryId", categoryID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update category",
		))
		return
	}

	// Can't update default categories
	if isDefault {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"Cannot update system default categories",
		))
		return
	}

	// Check if user owns this category
	if ownerID == nil || *ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to update this category",
		))
		return
	}

	// Update category
	var category models.Category
	err = db.DB.QueryRow(`
		UPDATE categories
		SET name = $1, color = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, user_id, name, color, is_default, created_at, updated_at
	`, req.Name, req.Color, time.Now(), categoryID).Scan(
		&category.ID, &category.UserID, &category.Name, &category.Color, &category.IsDefault, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		logger.Log.Errorw("Failed to update category", "error", err, "categoryId", categoryID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update category",
		))
		return
	}

	logger.Log.Infow("Category updated", "categoryId", categoryID, "userId", userID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(category))
}

// DeleteCategory deletes a custom category
func DeleteCategory(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Invalid category ID",
		))
		return
	}

	// Check if category exists and belongs to user (not a default category)
	var isDefault bool
	var ownerID *uuid.UUID
	err = db.DB.QueryRow("SELECT is_default, user_id FROM categories WHERE id = $1", categoryID).Scan(&isDefault, &ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Category not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get category", "error", err, "categoryId", categoryID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete category",
		))
		return
	}

	// Can't delete default categories
	if isDefault {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"Cannot delete system default categories",
		))
		return
	}

	// Check if user owns this category
	if ownerID == nil || *ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to delete this category",
		))
		return
	}

	// Check if category has expenses
	var expenseCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM expenses WHERE category_id = $1", categoryID).Scan(&expenseCount)
	if err != nil {
		logger.Log.Errorw("Failed to check expense count", "error", err, "categoryId", categoryID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete category",
		))
		return
	}

	if expenseCount > 0 {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"Cannot delete category with existing expenses",
		))
		return
	}

	// Delete category
	_, err = db.DB.Exec("DELETE FROM categories WHERE id = $1", categoryID)
	if err != nil {
		logger.Log.Errorw("Failed to delete category", "error", err, "categoryId", categoryID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete category",
		))
		return
	}

	logger.Log.Infow("Category deleted", "categoryId", categoryID, "userId", userID)

	c.Status(http.StatusNoContent)
}
