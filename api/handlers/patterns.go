package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
)

// GetMerchantPatterns retrieves all patterns for the user
func GetMerchantPatterns(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	isActive := c.Query("isActive")
	query := `SELECT id, user_id, merchant_name, category_id, match_type, is_active, use_count, last_used_at, created_at, updated_at
		FROM merchant_patterns WHERE user_id = $1`
	args := []interface{}{userID}

	if isActive == "true" {
		query += " AND is_active = true"
	} else if isActive == "false" {
		query += " AND is_active = false"
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		logger.Log.Errorw("Failed to get patterns", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve patterns"))
		return
	}
	defer rows.Close()

	patterns := []models.MerchantPattern{}
	for rows.Next() {
		var p models.MerchantPattern
		err := rows.Scan(&p.ID, &p.UserID, &p.MerchantName, &p.CategoryID, &p.MatchType, &p.IsActive, &p.UseCount, &p.LastUsedAt, &p.CreatedAt, &p.UpdatedAt)
		if err == nil {
			patterns = append(patterns, p)
		}
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(patterns))
}

// CreateMerchantPattern creates a new pattern
func CreateMerchantPattern(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	var req models.CreatePatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	// Check for existing pattern
	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM merchant_patterns WHERE user_id = $1 AND merchant_name = $2)", userID, req.MerchantName).Scan(&exists)
	if err == nil && exists {
		c.JSON(http.StatusConflict, models.NewErrorResponse(models.ErrCodeConflict, "Pattern already exists for this merchant"))
		return
	}

	var pattern models.MerchantPattern
	now := time.Now()
	err = db.DB.QueryRow(`
		INSERT INTO merchant_patterns (user_id, merchant_name, category_id, match_type, is_active, use_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, merchant_name, category_id, match_type, is_active, use_count, last_used_at, created_at, updated_at
	`, userID, req.MerchantName, req.CategoryID, req.MatchType, true, 0, now, now).Scan(
		&pattern.ID, &pattern.UserID, &pattern.MerchantName, &pattern.CategoryID, &pattern.MatchType, &pattern.IsActive, &pattern.UseCount, &pattern.LastUsedAt, &pattern.CreatedAt, &pattern.UpdatedAt,
	)
	if err != nil {
		logger.Log.Errorw("Failed to create pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to create pattern"))
		return
	}

	logger.Log.Infow("Pattern created", "patternId", pattern.ID, "userId", userID)
	c.JSON(http.StatusCreated, models.NewSuccessResponse(pattern))
}

// UpdateMerchantPattern updates a pattern
func UpdateMerchantPattern(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	patternID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "Invalid pattern ID"))
		return
	}

	var req models.UpdatePatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	// Check ownership
	var ownerID uuid.UUID
	err = db.DB.QueryRow("SELECT user_id FROM merchant_patterns WHERE id = $1", patternID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Pattern not found"))
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	// Build update
	updates := "updated_at = $1"
	args := []interface{}{time.Now()}
	argCount := 1

	if req.CategoryID != nil {
		argCount++
		updates += ", category_id = $" + string(rune(argCount+'0'))
		args = append(args, *req.CategoryID)
	}
	if req.MatchType != nil {
		argCount++
		updates += ", match_type = $" + string(rune(argCount+'0'))
		args = append(args, *req.MatchType)
	}
	if req.IsActive != nil {
		argCount++
		updates += ", is_active = $" + string(rune(argCount+'0'))
		args = append(args, *req.IsActive)
	}

	argCount++
	args = append(args, patternID)

	_, err = db.DB.Exec("UPDATE merchant_patterns SET "+updates+" WHERE id = $"+string(rune(argCount+'0')), args...)
	if err != nil {
		logger.Log.Errorw("Failed to update pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update pattern"))
		return
	}

	// Get updated pattern
	var pattern models.MerchantPattern
	err = db.DB.QueryRow("SELECT id, user_id, merchant_name, category_id, match_type, is_active, use_count, last_used_at, created_at, updated_at FROM merchant_patterns WHERE id = $1", patternID).Scan(
		&pattern.ID, &pattern.UserID, &pattern.MerchantName, &pattern.CategoryID, &pattern.MatchType, &pattern.IsActive, &pattern.UseCount, &pattern.LastUsedAt, &pattern.CreatedAt, &pattern.UpdatedAt,
	)
	if err != nil {
		logger.Log.Errorw("Failed to get updated pattern", "error", err)
	}

	logger.Log.Infow("Pattern updated", "patternId", patternID, "userId", userID)
	c.JSON(http.StatusOK, models.NewSuccessResponse(pattern))
}

// DeleteMerchantPattern deletes a pattern
func DeleteMerchantPattern(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	patternID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, "Invalid pattern ID"))
		return
	}

	// Check ownership
	var ownerID uuid.UUID
	err = db.DB.QueryRow("SELECT user_id FROM merchant_patterns WHERE id = $1", patternID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Pattern not found"))
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	_, err = db.DB.Exec("DELETE FROM merchant_patterns WHERE id = $1", patternID)
	if err != nil {
		logger.Log.Errorw("Failed to delete pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete pattern"))
		return
	}

	logger.Log.Infow("Pattern deleted", "patternId", patternID, "userId", userID)
	c.Status(http.StatusNoContent)
}

// MatchMerchantPattern tests if a merchant name matches any pattern
func MatchMerchantPattern(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	var req models.MatchPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(models.ErrCodeInvalidInput, err.Error()))
		return
	}

	// Get all active patterns
	rows, err := db.DB.Query(`
		SELECT id, user_id, merchant_name, category_id, match_type, is_active, use_count, last_used_at, created_at, updated_at
		FROM merchant_patterns
		WHERE user_id = $1 AND is_active = true
		ORDER BY match_type ASC
	`, userID)
	if err != nil {
		logger.Log.Errorw("Failed to get patterns", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to match pattern"))
		return
	}
	defer rows.Close()

	// Try to find a match
	merchantNameLower := strings.ToLower(req.MerchantName)
	for rows.Next() {
		var p models.MerchantPattern
		err := rows.Scan(&p.ID, &p.UserID, &p.MerchantName, &p.CategoryID, &p.MatchType, &p.IsActive, &p.UseCount, &p.LastUsedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}

		patternNameLower := strings.ToLower(p.MerchantName)
		matched := false

		if p.MatchType == "exact" {
			matched = merchantNameLower == patternNameLower
		} else if p.MatchType == "contains" {
			matched = strings.Contains(merchantNameLower, patternNameLower)
		}

		if matched {
			c.JSON(http.StatusOK, models.NewSuccessResponse(models.MatchPatternResponse{
				Matched: true,
				Pattern: &p,
			}))
			return
		}
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.MatchPatternResponse{
		Matched: false,
		Pattern: nil,
	}))
}
