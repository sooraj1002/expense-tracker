package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
)

// GetExpenses retrieves expenses with filters and pagination
func GetExpenses(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	accountIDStr := c.Query("accountId")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := db.DB.Where("user_id = ?", userID)

	if year > 0 {
		query = query.Where("EXTRACT(YEAR FROM date) = ?", year)
	}
	if month > 0 && month <= 12 {
		query = query.Where("EXTRACT(MONTH FROM date) = ?", month)
	}
	if accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			query = query.Where("account_id = ?", accountID)
		}
	}

	expenses := []models.Expense{}
	err = query.Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&expenses).Error
	if err != nil {
		logger.Log.Errorw("Failed to get expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve expenses"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(expenses))
}

// CreateExpense creates a new expense
func CreateExpense(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	var req models.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	// Start transaction
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create expense
		expense := models.Expense{
			UserID:       userID,
			Amount:       req.Amount,
			CategoryID:   req.CategoryID,
			AccountID:    req.AccountID,
			Date:         req.Date,
			Description:  req.Description,
			Source:       "manual",
			MerchantName: req.MerchantName,
			Verified:     true,
		}

		if err := tx.Create(&expense).Error; err != nil {
			return err
		}

		// Update account balance
		now := time.Now()
		err := tx.Model(&models.Account{}).
			Where("id = ? AND user_id = ?", req.AccountID, userID).
			Updates(map[string]interface{}{
				"current_balance": gorm.Expr("current_balance - ?", req.Amount),
				"total_spent":     gorm.Expr("total_spent + ?", req.Amount),
				"updated_at":      now,
			}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Log.Errorw("Failed to create expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}

	// Retrieve the created expense
	var expense models.Expense
	err = db.DB.Where("user_id = ? AND account_id = ?", userID, req.AccountID).
		Order("created_at DESC").
		First(&expense).Error
	if err != nil {
		logger.Log.Errorw("Failed to retrieve created expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}

	logger.Log.Infow("Expense created", "expenseId", expense.ID, "userId", userID)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(expense))
}

// UpdateExpense updates an existing expense
func UpdateExpense(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "Invalid expense ID"))
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	// Get existing expense
	var oldExpense models.Expense
	err = db.DB.Where("id = ?", expenseID).First(&oldExpense).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Expense not found"))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update expense"))
		return
	}

	if oldExpense.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	// Build updates map dynamically
	updates := make(map[string]interface{})

	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Verified != nil {
		updates["verified"] = *req.Verified
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "No fields to update"))
		return
	}

	updates["updated_at"] = time.Now()

	err = db.DB.Model(&models.Expense{}).
		Where("id = ?", expenseID).
		Updates(updates).Error
	if err != nil {
		logger.Log.Errorw("Failed to update expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update expense"))
		return
	}

	// Get updated expense
	var expense models.Expense
	err = db.DB.Where("id = ?", expenseID).First(&expense).Error
	if err != nil {
		logger.Log.Errorw("Failed to get updated expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update expense"))
		return
	}

	logger.Log.Infow("Expense updated", "expenseId", expenseID, "userId", userID)
	c.JSON(http.StatusOK, models.NewSuccessResponse(expense))
}

// DeleteExpense deletes an expense
func DeleteExpense(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "Invalid expense ID"))
		return
	}

	// Get expense details
	var expense models.Expense
	err = db.DB.Where("id = ?", expenseID).First(&expense).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Expense not found"))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}

	if expense.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	// Start transaction
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Delete expense
		if err := tx.Delete(&expense).Error; err != nil {
			return err
		}

		// Update account balance
		now := time.Now()
		err := tx.Model(&models.Account{}).
			Where("id = ? AND user_id = ?", expense.AccountID, userID).
			Updates(map[string]interface{}{
				"current_balance": gorm.Expr("current_balance + ?", expense.Amount),
				"total_spent":     gorm.Expr("total_spent - ?", expense.Amount),
				"updated_at":      now,
			}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Log.Errorw("Failed to delete expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}

	logger.Log.Infow("Expense deleted", "expenseId", expenseID, "userId", userID)
	c.Status(http.StatusNoContent)
}
