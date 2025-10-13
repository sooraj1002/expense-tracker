package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
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

	rows, err := db.DB.Query(`
		SELECT id, user_id, name, initial_balance, current_balance, total_spent, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		logger.Log.Errorw("Failed to get accounts", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to retrieve accounts",
		))
		return
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.UserID, &acc.Name, &acc.InitialBalance, &acc.CurrentBalance, &acc.TotalSpent, &acc.CreatedAt, &acc.UpdatedAt)
		if err != nil {
			logger.Log.Errorw("Failed to scan account", "error", err)
			continue
		}
		accounts = append(accounts, acc)
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

	var account models.Account
	now := time.Now()
	err = db.DB.QueryRow(`
		INSERT INTO accounts (user_id, name, initial_balance, current_balance, total_spent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, initial_balance, current_balance, total_spent, created_at, updated_at
	`, userID, req.Name, req.InitialBalance, req.InitialBalance, 0, now, now).Scan(
		&account.ID, &account.UserID, &account.Name, &account.InitialBalance, &account.CurrentBalance, &account.TotalSpent, &account.CreatedAt, &account.UpdatedAt,
	)
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
	var ownerID uuid.UUID
	var currentBalance, currentInitialBalance float64
	err = db.DB.QueryRow("SELECT user_id, current_balance, initial_balance FROM accounts WHERE id = $1", accountID).Scan(&ownerID, &currentBalance, &currentInitialBalance)
	if err == sql.ErrNoRows {
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

	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to update this account",
		))
		return
	}

	// Recalculate current balance if initial balance changed
	newCurrentBalance := currentBalance
	if req.InitialBalance != currentInitialBalance {
		diff := req.InitialBalance - currentInitialBalance
		newCurrentBalance = currentBalance + diff
	}

	var account models.Account
	err = db.DB.QueryRow(`
		UPDATE accounts
		SET name = $1, initial_balance = $2, current_balance = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, user_id, name, initial_balance, current_balance, total_spent, created_at, updated_at
	`, req.Name, req.InitialBalance, newCurrentBalance, time.Now(), accountID).Scan(
		&account.ID, &account.UserID, &account.Name, &account.InitialBalance, &account.CurrentBalance, &account.TotalSpent, &account.CreatedAt, &account.UpdatedAt,
	)
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
	var ownerID uuid.UUID
	err = db.DB.QueryRow("SELECT user_id FROM accounts WHERE id = $1", accountID).Scan(&ownerID)
	if err == sql.ErrNoRows {
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

	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(
			models.ErrCodeForbidden,
			"You don't have permission to delete this account",
		))
		return
	}

	// Check if account has expenses
	var expenseCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM expenses WHERE account_id = $1", accountID).Scan(&expenseCount)
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

	_, err = db.DB.Exec("DELETE FROM accounts WHERE id = $1", accountID)
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
	err = db.DB.QueryRow(`
		SELECT
			COALESCE(SUM(initial_balance), 0),
			COALESCE(SUM(current_balance), 0),
			COALESCE(SUM(total_spent), 0),
			COUNT(*)
		FROM accounts
		WHERE user_id = $1
	`, userID).Scan(&summary.TotalInitialBalance, &summary.TotalCurrentBalance, &summary.TotalSpent, &summary.AccountCount)
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
	var ownerID uuid.UUID
	err = db.DB.QueryRow("SELECT user_id FROM accounts WHERE id = $1", accountID).Scan(&ownerID)
	if err == sql.ErrNoRows {
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

	if ownerID != userID {
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

	// Build query
	query := `SELECT id, user_id, amount, category_id, account_id, date, description, source, merchant_id, merchant_name, location_id, raw_data, verified, created_at, updated_at
		FROM expenses WHERE account_id = $1`
	args := []interface{}{accountID}
	argCount := 1

	if year > 0 {
		argCount++
		query += " AND EXTRACT(YEAR FROM date) = $" + strconv.Itoa(argCount)
		args = append(args, year)
	}
	if month > 0 && month <= 12 {
		argCount++
		query += " AND EXTRACT(MONTH FROM date) = $" + strconv.Itoa(argCount)
		args = append(args, month)
	}

	query += " ORDER BY date DESC"
	argCount++
	query += " LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)
	argCount++
	query += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		logger.Log.Errorw("Failed to get account expenses", "error", err, "accountId", accountID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get account expenses",
		))
		return
	}
	defer rows.Close()

	expenses := []models.Expense{}
	totalSpent := 0.0
	for rows.Next() {
		var exp models.Expense
		err := rows.Scan(&exp.ID, &exp.UserID, &exp.Amount, &exp.CategoryID, &exp.AccountID, &exp.Date, &exp.Description, &exp.Source, &exp.MerchantID, &exp.MerchantName, &exp.LocationID, &exp.RawData, &exp.Verified, &exp.CreatedAt, &exp.UpdatedAt)
		if err != nil {
			logger.Log.Errorw("Failed to scan expense", "error", err)
			continue
		}
		expenses = append(expenses, exp)
		totalSpent += exp.Amount
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM expenses WHERE account_id = $1"
	countArgs := []interface{}{accountID}
	if year > 0 {
		countQuery += " AND EXTRACT(YEAR FROM date) = $2"
		countArgs = append(countArgs, year)
	}
	if month > 0 && month <= 12 {
		idx := len(countArgs) + 1
		countQuery += " AND EXTRACT(MONTH FROM date) = $" + strconv.Itoa(idx)
		countArgs = append(countArgs, month)
	}

	var totalCount int
	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		logger.Log.Errorw("Failed to count expenses", "error", err)
	}

	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"expenses":             expenses,
		"totalPages":           totalPages,
		"currentPage":          page,
		"totalSpentFromAccount": totalSpent,
	}))
}
