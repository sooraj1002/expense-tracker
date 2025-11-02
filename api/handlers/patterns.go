package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"gorm.io/gorm"
)

// GetMerchantPatterns retrieves all patterns for the user
func GetMerchantPatterns(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(models.ErrCodeUnauthorized, "User not authenticated"))
		return
	}

	isActive := c.Query("isActive")
	query := db.DB.Where("user_id = ?", userID)

	if isActive == "true" {
		query = query.Where("is_active = ?", true)
	} else if isActive == "false" {
		query = query.Where("is_active = ?", false)
	}

	patterns := []models.MerchantPattern{}
	err = query.Order("created_at DESC").Find(&patterns).Error
	if err != nil {
		logger.Log.Errorw("Failed to get patterns", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to retrieve patterns"))
		return
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
	var existingCount int64
	err = db.DB.Model(&models.MerchantPattern{}).
		Where("user_id = ? AND merchant_name = ?", userID, req.MerchantName).
		Count(&existingCount).Error
	if err != nil {
		logger.Log.Errorw("Failed to check existing pattern", "error", err)
	} else if existingCount > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(models.ErrCodeConflict, "Pattern already exists for this merchant"))
		return
	}

	pattern := models.MerchantPattern{
		UserID:       userID,
		MerchantName: req.MerchantName,
		CategoryID:   req.CategoryID,
		MatchType:    req.MatchType,
		IsActive:     true,
		UseCount:     0,
	}

	err = db.DB.Create(&pattern).Error
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
	var pattern models.MerchantPattern
	err = db.DB.Where("id = ?", patternID).First(&pattern).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Pattern not found"))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update pattern"))
		return
	}
	if pattern.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}
	if req.MatchType != nil {
		updates["match_type"] = *req.MatchType
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	err = db.DB.Model(&pattern).Updates(updates).Error
	if err != nil {
		logger.Log.Errorw("Failed to update pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update pattern"))
		return
	}

	// Reload pattern to get updated values
	err = db.DB.Where("id = ?", patternID).First(&pattern).Error
	if err != nil {
		logger.Log.Errorw("Failed to get updated pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to update pattern"))
		return
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
	var pattern models.MerchantPattern
	err = db.DB.Where("id = ?", patternID).First(&pattern).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(models.ErrCodeNotFound, "Pattern not found"))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get pattern", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to delete pattern"))
		return
	}
	if pattern.UserID != userID {
		c.JSON(http.StatusForbidden, models.NewErrorResponse(models.ErrCodeForbidden, "Permission denied"))
		return
	}

	err = db.DB.Delete(&pattern).Error
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
	var patterns []models.MerchantPattern
	err = db.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Order("match_type ASC").
		Find(&patterns).Error
	if err != nil {
		logger.Log.Errorw("Failed to get patterns", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(models.ErrCodeDatabaseError, "Failed to match pattern"))
		return
	}

	// Try to find a match
	merchantNameLower := strings.ToLower(req.MerchantName)
	for _, p := range patterns {
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
