package handlers

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExpenseWithCategory is a helper struct for joining expenses with categories
type ExpenseWithCategory struct {
	models.Expense
	CategoryName  string `gorm:"column:category_name"`
	CategoryColor string `gorm:"column:category_color"`
	IsDefault     bool   `gorm:"column:category_is_default"`
}

type expenseQueryOptions struct {
	Page        int
	Limit       int
	OrderClause string
}

func buildExpenseQuery(c *gin.Context, userID uuid.UUID) (*gorm.DB, expenseQueryOptions, error) {
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

	opts := expenseQueryOptions{
		Page:        page,
		Limit:       limit,
		OrderClause: "expenses.date DESC",
	}
	if sortOption == "updated" {
		opts.OrderClause = "expenses.updated_at DESC"
	}

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
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
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
		if startDateStr != "" {
			if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
				query = query.Where("expenses.date >= ?", startDate)
			}
		}
		if endDateStr != "" {
			if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
				endDate = endDate.AddDate(0, 0, 1)
				query = query.Where("expenses.date < ?", endDate)
			}
		}
	} else if year > 0 {
		query = query.Where("EXTRACT(YEAR FROM expenses.date) = ?", year)
		if month > 0 && month <= 12 {
			query = query.Where("EXTRACT(MONTH FROM expenses.date) = ?", month)
		}
	}

	if accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			query = query.Where("expenses.account_id = ?", accountID)
		} else {
			return nil, opts, err
		}
	}

	if categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			query = query.Where("expenses.category_id = ?", categoryID)
		} else {
			return nil, opts, err
		}
	}

	if tagsParam != "" {
		tags := []string{}
		for _, tag := range parseCommaSeparated(tagsParam) {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		if len(tags) > 0 {
			query = query.Where("expenses.tags && ?", pq.Array(tags))
		}
	}

	return query, opts, nil
}

// GetExpenses retrieves expenses with filters and pagination
func GetExpenses(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	query, opts, err := buildExpenseQuery(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
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
	offset := (opts.Page - 1) * opts.Limit

	err = query.Order(opts.OrderClause).
		Limit(opts.Limit).
		Offset(offset).
		Scan(&expensesWithCategories).Error
	if err != nil {
		logger.Log.Errorw("Failed to get expenses", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve expenses"))
		return
	}

	expenseIDs := make([]uuid.UUID, 0, len(expensesWithCategories))
	for _, exp := range expensesWithCategories {
		expenseIDs = append(expenseIDs, exp.ID)
	}
	tagMap, _ := loadTagsForExpenses(expenseIDs)

	// Convert to ExpenseResponse with embedded category
	expenseResponses := make([]models.ExpenseResponse, len(expensesWithCategories))
	for i, exp := range expensesWithCategories {
		tags := exp.Tags
		if mappedTags, ok := tagMap[exp.ID]; ok {
			tags = models.StringArray(mappedTags)
		}

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
			Tags:        tags,
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
	totalPages := int(totalCount) / opts.Limit
	if int(totalCount)%opts.Limit > 0 {
		totalPages++
	}

	response := models.ExpenseListResponse{
		Success: true,
		Data:    expenseResponses,
		Pagination: &models.PaginationMetadata{
			Page:        opts.Page,
			Limit:       opts.Limit,
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
	tagValues := normalizeTags(req.Tags)
	if len(tagValues) == 0 {
		tagValues = []string{"misc"}
	}
	tags := models.StringArray(tagValues)

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

		if err := syncExpenseTags(tx, userID, expense.ID, tagValues); err != nil {
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
	var normalizedTags []string
	tagsProvided := false

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
		tagsProvided = true
		normalizedTags = normalizeTags(req.Tags)
		if len(normalizedTags) == 0 {
			normalizedTags = []string{"misc"}
		}
		updates["tags"] = models.StringArray(normalizedTags)
	}
	if req.Verified != nil {
		updates["verified"] = *req.Verified
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "No fields to update"))
		return
	}

	updates["updated_at"] = time.Now()

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Expense{}).
			Where("id = ?", expenseID).
			Updates(updates).Error; err != nil {
			return err
		}

		if tagsProvided {
			if err := tx.Where("expense_id = ?", expenseID).Delete(&models.ExpenseTag{}).Error; err != nil {
				return err
			}
			if err := syncExpenseTags(tx, userID, expenseID, normalizedTags); err != nil {
				return err
			}
		}
		return nil
	})
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
		if err := tx.Where("expense_id = ?", expenseID).Delete(&models.ExpenseTag{}).Error; err != nil {
			return err
		}
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

// ExportExpensesCSV streams filtered expenses as CSV in batches to avoid blocking.
func ExportExpensesCSV(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	query, opts, err := buildExpenseQuery(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=expenses.csv")
	c.Status(http.StatusOK)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"Date", "Description", "Category", "Account", "Amount", "Tags", "Verified", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(headers); err != nil {
		logger.Log.Errorw("Failed to write CSV header", "error", err)
		return
	}

	batchSize := 500
	offset := 0

	for {
		var batch []ExpenseWithCategory
		err := query.Order(opts.OrderClause).
			Limit(batchSize).
			Offset(offset).
			Scan(&batch).Error
		if err != nil {
			logger.Log.Errorw("Failed to stream expenses", "error", err)
			return
		}
		if len(batch) == 0 {
			break
		}

		expenseIDs := make([]uuid.UUID, 0, len(batch))
		for _, exp := range batch {
			expenseIDs = append(expenseIDs, exp.ID)
		}
		tagMap, _ := loadTagsForExpenses(expenseIDs)

		for _, exp := range batch {
			tags := exp.Tags
			if mappedTags, ok := tagMap[exp.ID]; ok {
				tags = models.StringArray(mappedTags)
			}

			record := []string{
				exp.Date.Format("2006-01-02"),
				exp.Description,
				exp.CategoryName,
				exp.AccountID.String(),
				strconv.FormatFloat(exp.Amount, 'f', 2, 64),
				strings.Join(tags, ";"),
				strconv.FormatBool(exp.Verified),
				exp.CreatedAt.Format(time.RFC3339),
				exp.UpdatedAt.Format(time.RFC3339),
			}

			if err := writer.Write(record); err != nil {
				logger.Log.Errorw("Failed to write CSV record", "error", err)
				return
			}
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			logger.Log.Errorw("Error flushing CSV writer", "error", err)
			return
		}

		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}

		offset += batchSize
	}
}

// GetExpenseTags retrieves all unique tags for the user's expenses
func GetExpenseTags(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	var tags []string
	err = db.DB.Model(&models.Tag{}).
		Where("user_id = ?", userID).
		Order("name ASC").
		Pluck("name", &tags).Error
	if err != nil {
		logger.Log.Errorw("Failed to get expense tags", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve tags"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"tags":  tags,
		"count": len(tags),
	}))
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		lowered := strings.ToLower(trimmed)
		if seen[lowered] {
			continue
		}
		seen[lowered] = true
		normalized = append(normalized, lowered)
	}
	return normalized
}

func syncExpenseTags(tx *gorm.DB, userID uuid.UUID, expenseID uuid.UUID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	for _, tagName := range tags {
		var tag models.Tag
		err := tx.Where("user_id = ? AND name = ?", userID, tagName).First(&tag).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = models.Tag{
					UserID: userID,
					Name:   tagName,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		expenseTag := models.ExpenseTag{
			ExpenseID: expenseID,
			TagID:     tag.ID,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "expense_id"}, {Name: "tag_id"}},
			DoNothing: true,
		}).Create(&expenseTag).Error; err != nil {
			return err
		}
	}
	return nil
}

func loadTagsForExpenses(expenseIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	result := make(map[uuid.UUID][]string)
	if len(expenseIDs) == 0 {
		return result, nil
	}

	var rows []struct {
		ExpenseID uuid.UUID
		Name      string
	}

	err := db.DB.Table("expense_tags").
		Select("expense_tags.expense_id, tags.name").
		Joins("JOIN tags ON tags.id = expense_tags.tag_id").
		Where("expense_tags.expense_id IN ?", expenseIDs).
		Order("tags.name ASC").
		Scan(&rows).Error
	if err != nil {
		return result, err
	}

	for _, row := range rows {
		result[row.ExpenseID] = append(result[row.ExpenseID], row.Name)
	}

	return result, nil
}
