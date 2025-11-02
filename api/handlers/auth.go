package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/config"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"github.com/sooraj1002/expense-tracker/utils"
	"gorm.io/gorm"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Check if user already exists
	var count int64
	if err := db.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		logger.Log.Errorw("Failed to check user existence", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to check user existence",
		))
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, models.NewErrorResponse(
			models.ErrCodeConflict,
			"User with this email already exists",
		))
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Log.Errorw("Failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create user",
		))
		return
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		logger.Log.Errorw("Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to create user",
		))
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, config.AppConfig.JWT.Secret, config.AppConfig.JWT.Expiry)
	if err != nil {
		logger.Log.Errorw("Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to generate authentication token",
		))
		return
	}

	logger.Log.Infow("User registered successfully", "userId", user.ID, "email", user.Email)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(models.LoginResponse{
		User:  user,
		Token: token,
	}))
}

// Login handles user login
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Get user from database
	var user models.User
	err := db.DB.Where("email = ?", req.Email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Invalid email or password",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get user", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to authenticate user",
		))
		return
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Invalid email or password",
		))
		return
	}

	// Update last login time
	now := time.Now()
	if err := db.DB.Model(&user).Update("last_login_at", now).Error; err != nil {
		logger.Log.Warnw("Failed to update last login time", "error", err, "userId", user.ID)
	}
	user.LastLoginAt = &now

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, config.AppConfig.JWT.Secret, config.AppConfig.JWT.Expiry)
	if err != nil {
		logger.Log.Errorw("Failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to generate authentication token",
		))
		return
	}

	logger.Log.Infow("User logged in successfully", "userId", user.ID, "email", user.Email)

	c.JSON(http.StatusOK, models.NewSuccessResponse(models.LoginResponse{
		User:  user,
		Token: token,
	}))
}

// RefreshToken handles token refresh
func RefreshToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Validate old token (even if expired, we want to extract user info)
	claims, err := utils.ValidateToken(req.Token, config.AppConfig.JWT.Secret)
	if err != nil && err != utils.ErrExpiredToken {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Invalid token",
		))
		return
	}

	// Generate new token
	newToken, err := utils.GenerateToken(claims.UserID, claims.Email, config.AppConfig.JWT.Secret, config.AppConfig.JWT.Expiry)
	if err != nil {
		logger.Log.Errorw("Failed to generate new token", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to refresh token",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"token":     newToken,
		"expiresAt": time.Now().Add(config.AppConfig.JWT.Expiry),
	}))
}

// GetMe returns the current authenticated user's profile
func GetMe(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	var user models.User
	err = db.DB.Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"User not found",
		))
		return
	}
	if err != nil {
		logger.Log.Errorw("Failed to get user", "error", err, "userId", userID)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to get user profile",
		))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(user))
}

// RegisterDevice registers a new device for the authenticated user
func RegisterDevice(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
		))
		return
	}

	var req models.RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeInvalidInput,
			err.Error(),
		))
		return
	}

	// Check if device already exists
	var device models.Device
	err = db.DB.Where("device_id = ?", req.DeviceID).First(&device).Error
	if err == nil {
		// Device exists, update it
		device.DeviceName = req.DeviceName
		device.UpdatedAt = time.Now()

		if err := db.DB.Save(&device).Error; err != nil {
			logger.Log.Errorw("Failed to update device", "error", err)
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeDatabaseError,
				"Failed to update device",
			))
			return
		}

		c.JSON(http.StatusOK, models.NewSuccessResponse(device))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Log.Errorw("Failed to check device", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to check device",
		))
		return
	}

	// Create new device
	now := time.Now()
	device = models.Device{
		UserID:       userID,
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		RegisteredAt: now,
	}

	if err := db.DB.Create(&device).Error; err != nil {
		logger.Log.Errorw("Failed to register device", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to register device",
		))
		return
	}

	logger.Log.Infow("Device registered successfully", "userId", userID, "deviceId", req.DeviceID)

	c.JSON(http.StatusCreated, models.NewSuccessResponse(device))
}
