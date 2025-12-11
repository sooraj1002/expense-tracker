package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
)

// MigrateTags materializes tags into the tags/expense_tags tables from existing expense tags arrays.
func MigrateTags(c *gin.Context) {
	type expenseTagRow struct {
		ID     uuid.UUID
		UserID uuid.UUID
		Tags   models.StringArray
	}

	var rows []expenseTagRow
	if err := db.DB.Model(&models.Expense{}).
		Select("id, user_id, tags").
		Find(&rows).Error; err != nil {
		logger.Log.Errorw("Failed to load expenses for migration", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to load expenses for migration"))
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
			"migrated": 0,
			"message":  "No expenses found to migrate",
		}))
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			normalized := normalizeTags(row.Tags)
			if len(normalized) == 0 {
				normalized = []string{"misc"}
			}

			// Keep legacy column in sync for compatibility
			if err := tx.Model(&models.Expense{}).
				Where("id = ?", row.ID).
				Update("tags", models.StringArray(normalized)).Error; err != nil {
				return err
			}

			if err := syncExpenseTags(tx, row.UserID, row.ID, normalized); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		logger.Log.Errorw("Failed to migrate tags", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to migrate tags"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"migrated": len(rows),
		"message":  "Tags migrated to normalized tables",
	}))
}

// ExportDatabase returns a JSON snapshot of core tables for backup.
func ExportDatabase(c *gin.Context) {
	payload := struct {
		Users            []models.User            `json:"users"`
		Categories       []models.Category        `json:"categories"`
		Accounts         []models.Account         `json:"accounts"`
		Expenses         []models.Expense         `json:"expenses"`
		Tags             []models.Tag             `json:"tags"`
		ExpenseTags      []models.ExpenseTag      `json:"expenseTags"`
		MerchantInfos    []models.MerchantInfo    `json:"merchantInfos"`
		MerchantPatterns []models.MerchantPattern `json:"merchantPatterns"`
	}{
		Users:            []models.User{},
		Categories:       []models.Category{},
		Accounts:         []models.Account{},
		Expenses:         []models.Expense{},
		Tags:             []models.Tag{},
		ExpenseTags:      []models.ExpenseTag{},
		MerchantInfos:    []models.MerchantInfo{},
		MerchantPatterns: []models.MerchantPattern{},
	}

	if err := db.DB.Find(&payload.Users).Error; err != nil {
		logger.Log.Errorw("Export users failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export users"))
		return
	}
	if err := db.DB.Find(&payload.Categories).Error; err != nil {
		logger.Log.Errorw("Export categories failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export categories"))
		return
	}
	if err := db.DB.Find(&payload.Accounts).Error; err != nil {
		logger.Log.Errorw("Export accounts failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export accounts"))
		return
	}
	if err := db.DB.Find(&payload.Expenses).Error; err != nil {
		logger.Log.Errorw("Export expenses failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export expenses"))
		return
	}
	if err := db.DB.Find(&payload.Tags).Error; err != nil {
		logger.Log.Errorw("Export tags failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export tags"))
		return
	}
	if err := db.DB.Find(&payload.ExpenseTags).Error; err != nil {
		logger.Log.Errorw("Export expense tags failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export expense tags"))
		return
	}
	if err := db.DB.Find(&payload.MerchantInfos).Error; err != nil {
		logger.Log.Errorw("Export merchant info failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export merchant info"))
		return
	}
	if err := db.DB.Find(&payload.MerchantPatterns).Error; err != nil {
		logger.Log.Errorw("Export merchant patterns failed", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to export merchant patterns"))
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=db-export.json")
	encoder := json.NewEncoder(c.Writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		logger.Log.Errorw("Failed to encode export", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeInternalError, "Failed to encode export"))
		return
	}
}
