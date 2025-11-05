package handlers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
)

// ExpenseWithCategory is a helper struct for joining expenses with categories
type ExpenseWithCategory struct {
	models.Expense
	CategoryName  string `gorm:"column:category_name"`
	CategoryColor string `gorm:"column:category_color"`
	IsDefault     bool   `gorm:"column:category_is_default"`
}

// GetExpenses retrieves expenses with filters and pagination
func GetExpenses(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	// Parse filter parameters
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	accountIDStr := c.Query("accountId")
	categoryIDStr := c.Query("categoryId")
	tagsParam := c.Query("tags")
	period := c.Query("period")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	sortOption := c.DefaultQuery("sort", "date")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build base query with JOIN on categories table
	query := db.DB.Table("expenses").
		Select(`expenses.*,
			categories.name as category_name,
			categories.color as category_color,
			categories.is_default as category_is_default`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ?", userID)

	// Handle date filtering with priority: period > custom range > month/year
	if period != "" {
		now := time.Now()
		var startDate time.Time
		switch period {
		case "today":
			startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			query = query.Where("expenses.date >= ?", startDate)
		case "week":
			// Start of week (Monday)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday
			}
			startDate = now.AddDate(0, 0, -(weekday - 1))
			startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
			query = query.Where("expenses.date >= ?", startDate)
		case "month":
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			query = query.Where("expenses.date >= ?", startDate)
		case "year":
			startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			query = query.Where("expenses.date >= ?", startDate)
		}
	} else if startDateStr != "" || endDateStr != "" {
		// Custom date range
		if startDateStr != "" {
			if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
				query = query.Where("expenses.date >= ?", startDate)
			}
		}
		if endDateStr != "" {
			if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
				// Add one day to include the end date
				endDate = endDate.AddDate(0, 0, 1)
				query = query.Where("expenses.date < ?", endDate)
			}
		}
	} else if year > 0 {
		// Legacy month/year filtering
		query = query.Where("EXTRACT(YEAR FROM expenses.date) = ?", year)
		if month > 0 && month <= 12 {
			query = query.Where("EXTRACT(MONTH FROM expenses.date) = ?", month)
		}
	}

	// Account filter
	if accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			query = query.Where("expenses.account_id = ?", accountID)
		}
	}

	// Category filter
	if categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			query = query.Where("expenses.category_id = ?", categoryID)
		}
	}

	// Tags filter (supports comma-separated tags)
	if tagsParam != "" {
		tags := []string{}
		for _, tag := range parseCommaSeparated(tagsParam) {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		if len(tags) > 0 {
			// Use PostgreSQL array overlap operator with pq.Array for proper array conversion
			query = query.Where("expenses.tags && ?", pq.Array(tags))
		}
	}

	// Get total count for pagination
	var totalCount int64
	if err := query.Session(&gorm.Session{}).Count(&totalCount).Error; err != nil {
		logger.Log.Errorw("Failed to count expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve expenses"))
		return
	}

	// Calculate total amount for filtered results
	var totalAmount float64
	err = query.Session(&gorm.Session{}).Select("COALESCE(SUM(expenses.amount), 0)").Row().Scan(&totalAmount)
	if err != nil {
		logger.Log.Errorw("Failed to calculate total amount", "error", err)
		// Continue without failing the request
		totalAmount = 0
	}

	// Execute query with pagination
	var expensesWithCategories []ExpenseWithCategory
	orderClause := "expenses.date DESC"
	if sortOption == "updated" {
		orderClause = "expenses.updated_at DESC"
	}

	err = query.Order(orderClause).
		Limit(limit).
		Offset(offset).
		Scan(&expensesWithCategories).Error
	if err != nil {
		logger.Log.Errorw("Failed to get expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve expenses"))
		return
	}

	// Convert to ExpenseResponse with embedded category
	expenseResponses := make([]models.ExpenseResponse, len(expensesWithCategories))
	for i, exp := range expensesWithCategories {
		expenseResponses[i] = models.ExpenseResponse{
			ID:     exp.ID.String(),
			UserID: exp.UserID.String(),
			Amount: exp.Amount,
			Category: models.CategoryBasic{
				CategoryID:   exp.CategoryID.String(),
				CategoryName: exp.CategoryName,
				Color:        exp.CategoryColor,
				IsDefault:    exp.IsDefault,
			},
			AccountID:   exp.AccountID.String(),
			Date:        exp.Date,
			Description: exp.Description,
			Tags:        exp.Tags,
			Verified:    exp.Verified,
			CreatedAt:   exp.CreatedAt,
			UpdatedAt:   exp.UpdatedAt,
		}

		if exp.MerchantID != nil {
			merchantIDStr := exp.MerchantID.String()
			expenseResponses[i].MerchantID = &merchantIDStr
		}
		if exp.LocationID != nil {
			locationIDStr := exp.LocationID.String()
			expenseResponses[i].LocationID = &locationIDStr
		}
		if exp.RawData != "" {
			expenseResponses[i].RawData = exp.RawData
		}
	}

	// Calculate pagination metadata
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}

	response := models.ExpenseListResponse{
		Success: true,
		Data:    expenseResponses,
		Pagination: &models.PaginationMetadata{
			Page:        page,
			Limit:       limit,
			TotalCount:  int(totalCount),
			TotalPages:  totalPages,
			TotalAmount: totalAmount,
		},
	}

	c.JSON(http.StatusOK, response)
}

// parseCommaSeparated splits a comma-separated string
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	for _, part := range splitByComma(s) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitByComma(s string) []string {
	var result []string
	var current string
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
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

	// Set default tags if not provided
	tags := models.StringArray(req.Tags)
	if len(tags) == 0 {
		tags = models.StringArray{"misc"}
	}

	// Start transaction
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create expense
		expense := models.Expense{
			UserID:      userID,
			Amount:      req.Amount,
			CategoryID:  req.CategoryID,
			AccountID:   req.AccountID,
			Date:        req.Date,
			Description: req.Description,
			Tags:        tags,
			Verified:    true,
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

	// Retrieve the created expense with category details
	var expenseWithCategory ExpenseWithCategory
	err = db.DB.Table("expenses").
		Select(`expenses.*,
			categories.name as category_name,
			categories.color as category_color,
			categories.is_default as category_is_default`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.user_id = ? AND expenses.account_id = ?", userID, req.AccountID).
		Order("expenses.created_at DESC").
		First(&expenseWithCategory).Error
	if err != nil {
		logger.Log.Errorw("Failed to retrieve created expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create expense"))
		return
	}

	// Convert to ExpenseResponse
	expenseResponse := models.ExpenseResponse{
		ID:     expenseWithCategory.ID.String(),
		UserID: expenseWithCategory.UserID.String(),
		Amount: expenseWithCategory.Amount,
		Category: models.CategoryBasic{
			CategoryID:   expenseWithCategory.CategoryID.String(),
			CategoryName: expenseWithCategory.CategoryName,
			Color:        expenseWithCategory.CategoryColor,
			IsDefault:    expenseWithCategory.IsDefault,
		},
		AccountID:   expenseWithCategory.AccountID.String(),
		Date:        expenseWithCategory.Date,
		Description: expenseWithCategory.Description,
		Tags:        expenseWithCategory.Tags,
		Verified:    expenseWithCategory.Verified,
		CreatedAt:   expenseWithCategory.CreatedAt,
		UpdatedAt:   expenseWithCategory.UpdatedAt,
	}

	if expenseWithCategory.MerchantID != nil {
		merchantIDStr := expenseWithCategory.MerchantID.String()
		expenseResponse.MerchantID = &merchantIDStr
	}
	if expenseWithCategory.LocationID != nil {
		locationIDStr := expenseWithCategory.LocationID.String()
		expenseResponse.LocationID = &locationIDStr
	}
	if expenseWithCategory.RawData != "" {
		expenseResponse.RawData = expenseWithCategory.RawData
	}

	logger.Log.Infow("Expense created", "expenseId", expenseWithCategory.ID, "userId", userID)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(expenseResponse))
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
	if req.AccountID != nil {
		updates["account_id"] = *req.AccountID
	}
	if req.Date != nil {
		updates["date"] = *req.Date
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Tags != nil {
		if len(req.Tags) == 0 {
			updates["tags"] = models.StringArray{"misc"}
		} else {
			updates["tags"] = models.StringArray(req.Tags)
		}
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

	// Get updated expense with category details
	var expenseWithCategory ExpenseWithCategory
	err = db.DB.Table("expenses").
		Select(`expenses.*,
			categories.name as category_name,
			categories.color as category_color,
			categories.is_default as category_is_default`).
		Joins("LEFT JOIN categories ON expenses.category_id = categories.id").
		Where("expenses.id = ?", expenseID).
		First(&expenseWithCategory).Error
	if err != nil {
		logger.Log.Errorw("Failed to get updated expense", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update expense"))
		return
	}

	// Convert to ExpenseResponse
	expenseResponse := models.ExpenseResponse{
		ID:     expenseWithCategory.ID.String(),
		UserID: expenseWithCategory.UserID.String(),
		Amount: expenseWithCategory.Amount,
		Category: models.CategoryBasic{
			CategoryID:   expenseWithCategory.CategoryID.String(),
			CategoryName: expenseWithCategory.CategoryName,
			Color:        expenseWithCategory.CategoryColor,
			IsDefault:    expenseWithCategory.IsDefault,
		},
		AccountID:   expenseWithCategory.AccountID.String(),
		Date:        expenseWithCategory.Date,
		Description: expenseWithCategory.Description,
		Tags:        expenseWithCategory.Tags,
		Verified:    expenseWithCategory.Verified,
		CreatedAt:   expenseWithCategory.CreatedAt,
		UpdatedAt:   expenseWithCategory.UpdatedAt,
	}

	if expenseWithCategory.MerchantID != nil {
		merchantIDStr := expenseWithCategory.MerchantID.String()
		expenseResponse.MerchantID = &merchantIDStr
	}
	if expenseWithCategory.LocationID != nil {
		locationIDStr := expenseWithCategory.LocationID.String()
		expenseResponse.LocationID = &locationIDStr
	}
	if expenseWithCategory.RawData != "" {
		expenseResponse.RawData = expenseWithCategory.RawData
	}

	logger.Log.Infow("Expense updated", "expenseId", expenseID, "userId", userID)
	c.JSON(http.StatusOK, models.NewSuccessResponse(expenseResponse))
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

// GetExpenseTags retrieves all unique tags for the user's expenses
func GetExpenseTags(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	// Query to get all unique tags from user's expenses
	var expenses []models.Expense
	err = db.DB.Select("tags").Where("user_id = ?", userID).Find(&expenses).Error
	if err != nil {
		logger.Log.Errorw("Failed to get expense tags", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve tags"))
		return
	}

	// Collect all unique tags
	tagSet := make(map[string]bool)
	for _, expense := range expenses {
		for _, tag := range expense.Tags {
			if tag != "" {
				tagSet[tag] = true
			}
		}
	}

	// Convert to slice
	uniqueTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		uniqueTags = append(uniqueTags, tag)
	}

	// Sort tags alphabetically for consistent output
	sort.Strings(uniqueTags)

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"tags":  uniqueTags,
		"count": len(uniqueTags),
	}))
}
