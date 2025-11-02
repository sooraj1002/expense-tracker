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

// GetAccounts retrieves all user accounts
func GetAccounts(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	accounts := []models.Account{}
	err = db.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&accounts).Error
	if err != nil {
		logger.Log.Errorw("Failed to get accounts", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve accounts",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(accounts))
}

// CreateAccount creates a new account
func CreateAccount(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	account := models.Account{
		UserID:         userID,
		Name:           req.Name,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		TotalSpent:     0,
	}

	err = db.DB.Create(&account).Error
	if err != nil {
		logger.Log.Errorw("Failed to create account", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to create account",
		))
		return
	}

	logger.Log.Infow("Account created", "accountId", account.ID, "userId", userID)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(account))
}

// UpdateAccount updates an existing account
func UpdateAccount(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Invalid account ID",
		))
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Check if account belongs to user
	var account models.Account
	err = db.DB.Where("id = ?", accountID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Account not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get account", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update account",
		))
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to update this account",
		))
		return
	}

	// Recalculate current balance if initial balance changed
	if req.InitialBalance != account.InitialBalance {
		diff := req.InitialBalance - account.InitialBalance
		account.CurrentBalance = account.CurrentBalance + diff
	}

	// Update account fields
	account.Name = req.Name
	account.InitialBalance = req.InitialBalance
	account.UpdatedAt = time.Now()

	err = db.DB.Save(&account).Error
	if err != nil {
		logger.Log.Errorw("Failed to update account", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to update account",
		))
		return
	}

	logger.Log.Infow("Account updated", "accountId", accountID, "userId", userID)

	c.JSON(http.StatusOK, models.NewSuccessResponse(account))
}

// DeleteAccount deletes an account
func DeleteAccount(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Invalid account ID",
		))
		return
	}

	// Check if account belongs to user
	var account models.Account
	err = db.DB.Where("id = ?", accountID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Account not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get account", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete account",
		))
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to delete this account",
		))
		return
	}

	// Check if account has expenses
	var expenseCount int64
	err = db.DB.Model(&models.Expense{}).Where("account_id = ?", accountID).Count(&expenseCount).Error
	if err != nil {
		logger.Log.Errorw("Failed to check expense count", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete account",
		))
		return
	}

	if expenseCount > 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Cannot delete account with existing expenses",
		))
		return
	}

	err = db.DB.Delete(&account).Error
	if err != nil {
		logger.Log.Errorw("Failed to delete account", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to delete account",
		))
		return
	}

	logger.Log.Infow("Account deleted", "accountId", accountID, "userId", userID)

	c.Status(http.StatusNoContent)
}

// GetAccountSummary returns summary of all accounts
func GetAccountSummary(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	var summary models.AccountSummary
	err = db.DB.Model(&models.Account{}).
		Select("COALESCE(SUM(initial_balance), 0) as total_initial_balance, COALESCE(SUM(current_balance), 0) as total_current_balance, COALESCE(SUM(total_spent), 0) as total_spent, COUNT(*) as account_count").
		Where("user_id = ?", userID).
		Scan(&summary).Error
	if err != nil {
		logger.Log.Errorw("Failed to get account summary", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get account summary",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(summary))
}

// GetAccountExpenses returns expenses for a specific account
func GetAccountExpenses(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			"Invalid account ID",
		))
		return
	}

	// Check if account belongs to user
	var account models.Account
	err = db.DB.Where("id = ?", accountID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Account not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get account", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get account expenses",
		))
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to view this account",
		))
		return
	}

	// Parse query parameters
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build query with GORM
	query := db.DB.Where("account_id = ?", accountID)

	if year > 0 {
		query = query.Where("EXTRACT(YEAR FROM date) = ?", year)
	}
	if month > 0 && month <= 12 {
		query = query.Where("EXTRACT(MONTH FROM date) = ?", month)
	}

	// Get total count
	var totalCount int64
	countQuery := query
	err = countQuery.Model(&models.Expense{}).Count(&totalCount).Error
	if err != nil {
		logger.Log.Errorw("Failed to count expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get account expenses",
		))
		return
	}

	// Get expenses
	expenses := []models.Expense{}
	err = query.Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&expenses).Error
	if err != nil {
		logger.Log.Errorw("Failed to get account expenses", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get account expenses",
		))
		return
	}

	totalSpent := 0.0
	for _, exp := range expenses {
		totalSpent += exp.Amount
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"expenses":             expenses,
		"totalPages":           totalPages,
		"currentPage":          page,
		"totalSpentFromAccount": totalSpent,
	}))
}
