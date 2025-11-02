package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
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
	categories := []models.Category{}
	err = db.DB.Where("user_id IS NULL OR user_id = ?", userID).
		Order("is_default DESC, name ASC").
		Find(&categories).Error
	if err != nil {
		logger.Log.Errorw("Failed to get categories", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve categories",
		))
		return
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
	category := models.Category{
		UserID:    &userID,
		Name:      req.Name,
		Color:     req.Color,
		IsDefault: false,
	}

	err = db.DB.Create(&category).Error
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
	var category models.Category
	err = db.DB.Where("id = ?", categoryID).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if category.IsDefault {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"Cannot update system default categories",
		))
		return
	}

	// Check if user owns this category
	if category.UserID == nil || *category.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to update this category",
		))
		return
	}

	// Update category
	category.Name = req.Name
	category.Color = req.Color
	category.UpdatedAt = time.Now()

	err = db.DB.Save(&category).Error
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
	var category models.Category
	err = db.DB.Where("id = ?", categoryID).First(&category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if category.IsDefault {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"Cannot delete system default categories",
		))
		return
	}

	// Check if user owns this category
	if category.UserID == nil || *category.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to delete this category",
		))
		return
	}

	// Check if category has expenses
	var expenseCount int64
	err = db.DB.Model(&models.Expense{}).Where("category_id = ?", categoryID).Count(&expenseCount).Error
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
	err = db.DB.Delete(&category).Error
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
