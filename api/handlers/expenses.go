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

	query := `SELECT id, user_id, amount, category_id, account_id, date, description, source, merchant_id, merchant_name, location_id, raw_data, verified, created_at, updated_at
		FROM expenses WHERE user_id = $1`
	args := []interface{}{userID}
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
	if accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			argCount++
			query += " AND account_id = $" + strconv.Itoa(argCount)
			args = append(args, accountID)
		}
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
		logger.Log.Errorw("Failed to get expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve expenses"))
		return
	}
	defer rows.Close()

	expenses := []models.Expense{}
	for rows.Next() {
		var exp models.Expense
		err := rows.Scan(&exp.ID, &exp.UserID, &exp.Amount, &exp.CategoryID, &exp.AccountID, &exp.Date, &exp.Description, &exp.Source, &exp.MerchantID, &exp.MerchantName, &exp.LocationID, &exp.RawData, &exp.Verified, &exp.CreatedAt, &exp.UpdatedAt)
		if err == nil {
			expenses = append(expenses, exp)
		}
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
	tx, err := db.DB.Begin()
	if err != nil {
		logger.Log.Errorw("Failed to begin transaction", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}
	defer tx.Rollback()

	// Create expense
	var expense models.Expense
	now := time.Now()
	err = tx.QueryRow(`
		INSERT INTO expenses (user_id, amount, category_id, account_id, date, description, source, merchant_name, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, amount, category_id, account_id, date, description, source, merchant_id, merchant_name, location_id, raw_data, verified, created_at, updated_at
	`, userID, req.Amount, req.CategoryID, req.AccountID, req.Date, req.Description, "manual", req.MerchantName, true, now, now).Scan(
		&expense.ID, &expense.UserID, &expense.Amount, &expense.CategoryID, &expense.AccountID, &expense.Date, &expense.Description, &expense.Source, &expense.MerchantID, &expense.MerchantName, &expense.LocationID, &expense.RawData, &expense.Verified, &expense.CreatedAt, &expense.UpdatedAt,
	)
	if err != nil {
		logger.Log.Errorw("Failed to create expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}

	// Update account balance
	_, err = tx.Exec(`
		UPDATE accounts
		SET current_balance = current_balance - $1, total_spent = total_spent + $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`, req.Amount, now, req.AccountID, userID)
	if err != nil {
		logger.Log.Errorw("Failed to update account balance", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Errorw("Failed to commit transaction", "error", err)
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
	err = db.DB.QueryRow("SELECT user_id, amount, account_id FROM expenses WHERE id = $1", expenseID).Scan(&oldExpense.UserID, &oldExpense.Amount, &oldExpense.AccountID)
	if err == sql.ErrNoRows {
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

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argCount := 0

	if req.Amount != nil {
		argCount++
		updates = append(updates, "amount = $"+strconv.Itoa(argCount))
		args = append(args, *req.Amount)
	}
	if req.CategoryID != nil {
		argCount++
		updates = append(updates, "category_id = $"+strconv.Itoa(argCount))
		args = append(args, *req.CategoryID)
	}
	if req.Description != nil {
		argCount++
		updates = append(updates, "description = $"+strconv.Itoa(argCount))
		args = append(args, *req.Description)
	}
	if req.Verified != nil {
		argCount++
		updates = append(updates, "verified = $"+strconv.Itoa(argCount))
		args = append(args, *req.Verified)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "No fields to update"))
		return
	}

	argCount++
	updates = append(updates, "updated_at = $"+strconv.Itoa(argCount))
	args = append(args, time.Now())

	argCount++
	args = append(args, expenseID)

	query := "UPDATE expenses SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = $" + strconv.Itoa(argCount)

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		logger.Log.Errorw("Failed to update expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update expense"))
		return
	}

	// Get updated expense
	var expense models.Expense
	err = db.DB.QueryRow("SELECT id, user_id, amount, category_id, account_id, date, description, source, merchant_id, merchant_name, location_id, raw_data, verified, created_at, updated_at FROM expenses WHERE id = $1", expenseID).Scan(
		&expense.ID, &expense.UserID, &expense.Amount, &expense.CategoryID, &expense.AccountID, &expense.Date, &expense.Description, &expense.Source, &expense.MerchantID, &expense.MerchantName, &expense.LocationID, &expense.RawData, &expense.Verified, &expense.CreatedAt, &expense.UpdatedAt,
	)
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
	err = db.DB.QueryRow("SELECT user_id, amount, account_id FROM expenses WHERE id = $1", expenseID).Scan(&expense.UserID, &expense.Amount, &expense.AccountID)
	if err == sql.ErrNoRows {
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
	tx, err := db.DB.Begin()
	if err != nil {
		logger.Log.Errorw("Failed to begin transaction", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}
	defer tx.Rollback()

	// Delete expense
	_, err = tx.Exec("DELETE FROM expenses WHERE id = $1", expenseID)
	if err != nil {
		logger.Log.Errorw("Failed to delete expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}

	// Update account balance
	_, err = tx.Exec(`
		UPDATE accounts
		SET current_balance = current_balance + $1, total_spent = total_spent - $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`, expense.Amount, time.Now(), expense.AccountID, userID)
	if err != nil {
		logger.Log.Errorw("Failed to update account balance", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}

	if err = tx.Commit(); err != nil {
		logger.Log.Errorw("Failed to commit transaction", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete expense"))
		return
	}

	logger.Log.Infow("Expense deleted", "expenseId", expenseID, "userId", userID)
	c.Status(http.StatusNoContent)
}
