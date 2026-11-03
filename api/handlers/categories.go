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
	"gorm.io/gorm/clause"
)

type categoryMonthlySpend struct {
	CategoryID uuid.UUID `gorm:"column:category_id"`
	Amount     float64   `gorm:"column:amount"`
}

func loadCategoryResponses(userID uuid.UUID, categories []models.Category) ([]models.CategoryResponse, error) {
	budgets := []models.CategoryBudget{}
	if err := db.DB.Where("user_id = ?", userID).Find(&budgets).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := monthStart.AddDate(0, 1, 0)
	spending := []categoryMonthlySpend{}
	if err := db.DB.Model(&models.Expense{}).
		Select("category_id, COALESCE(SUM(amount), 0) AS amount").
		Where("user_id = ? AND date >= ? AND date < ? AND currency = ?", userID, monthStart, nextMonth, "INR").
		Group("category_id").
		Scan(&spending).Error; err != nil {
		return nil, err
	}

	budgetByCategory := make(map[uuid.UUID]float64, len(budgets))
	for _, budget := range budgets {
		budgetByCategory[budget.CategoryID] = budget.Amount
	}
	spendByCategory := make(map[uuid.UUID]float64, len(spending))
	for _, spend := range spending {
		spendByCategory[spend.CategoryID] = spend.Amount
	}

	responses := make([]models.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		response := models.CategoryResponse{
			Category:     category,
			MonthlySpent: spendByCategory[category.ID],
		}
		if amount, ok := budgetByCategory[category.ID]; ok {
			budgetAmount := amount
			response.MonthlyBudget = &budgetAmount
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func getVisibleCategory(userID, categoryID uuid.UUID) (models.Category, error) {
	var category models.Category
	err := db.DB.Where("id = ? AND (user_id IS NULL OR user_id = ?)", categoryID, userID).
		First(&category).Error
	return category, err
}

func respondWithCategory(c *gin.Context, userID uuid.UUID, category models.Category) {
	responses, err := loadCategoryResponses(userID, []models.Category{category})
	if err != nil {
		logger.Log.Errorw("Failed to load category budget status", "error", err, "categoryId", category.ID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve category budget",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(responses[0]))
}

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

	responses, err := loadCategoryResponses(userID, categories)
	if err != nil {
		logger.Log.Errorw("Failed to load category budget status", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve categories",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(responses))
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

	responses, err := loadCategoryResponses(userID, []models.Category{category})
	if err != nil {
		logger.Log.Errorw("Failed to load created category budget status", "error", err, "categoryId", category.ID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Category created but could not be retrieved",
		))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse(responses[0]))
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

	respondWithCategory(c, userID, category)
}

// SetCategoryBudget creates or updates the authenticated user's monthly budget.
func SetCategoryBudget(c *gin.Context) {
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

	var req models.SetCategoryBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	category, err := getVisibleCategory(userID, categoryID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Category not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get category for budget", "error", err, "categoryId", categoryID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update category budget",
		))
		return
	}

	budget := models.CategoryBudget{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     req.Amount,
	}
	err = db.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "category_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"}),
	}).Create(&budget).Error
	if err != nil {
		logger.Log.Errorw("Failed to set category budget", "error", err, "categoryId", categoryID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update category budget",
		))
		return
	}

	logger.Log.Infow("Category budget updated", "categoryId", categoryID, "userId", userID, "amount", req.Amount)
	respondWithCategory(c, userID, category)
}

// ClearCategoryBudget removes the authenticated user's monthly category budget.
func ClearCategoryBudget(c *gin.Context) {
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

	category, err := getVisibleCategory(userID, categoryID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Category not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get category for budget", "error", err, "categoryId", categoryID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to clear category budget",
		))
		return
	}

	if err := db.DB.Where("user_id = ? AND category_id = ?", userID, categoryID).
		Delete(&models.CategoryBudget{}).Error; err != nil {
		logger.Log.Errorw("Failed to clear category budget", "error", err, "categoryId", categoryID, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to clear category budget",
		))
		return
	}

	logger.Log.Infow("Category budget cleared", "categoryId", categoryID, "userId", userID)
	respondWithCategory(c, userID, category)
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
