package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sooraj1002/expense-tracker/api/middleware"
	"github.com/sooraj1002/expense-tracker/config"
	"github.com/sooraj1002/expense-tracker/db"
	"github.com/sooraj1002/expense-tracker/logger"
	"github.com/sooraj1002/expense-tracker/models"
	"github.com/sooraj1002/expense-tracker/utils"
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
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		logger.Log.Errorw("Failed to check user existence", "error", err)
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeDatabaseError,
			"Failed to check user existence",
		))
		return
	}

	if exists {
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
	var user models.User
	err = db.DB.QueryRow(`
		INSERT INTO users (email, password_hash, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, name, created_at, last_login_at, updated_at
	`, req.Email, passwordHash, req.Name, time.Now(), time.Now()).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.LastLoginAt, &user.UpdatedAt,
	)
	if err != nil {
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
	var passwordHash string
	err := db.DB.QueryRow(`
		SELECT id, email, password_hash, name, created_at, last_login_at, updated_at
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &passwordHash, &user.Name, &user.CreatedAt, &user.LastLoginAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
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
	if !utils.CheckPassword(req.Password, passwordHash) {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Invalid email or password",
		))
		return
	}

	// Update last login time
	now := time.Now()
	_, err = db.DB.Exec("UPDATE users SET last_login_at = $1 WHERE id = $2", now, user.ID)
	if err != nil {
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
	err = db.DB.QueryRow(`
		SELECT id, email, name, created_at, last_login_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.LastLoginAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
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
	var existingID uuid.UUID
	err = db.DB.QueryRow("SELECT id FROM devices WHERE device_id = $1", req.DeviceID).Scan(&existingID)
	if err == nil {
		// Device exists, update it
		var device models.Device
		err = db.DB.QueryRow(`
			UPDATE devices
			SET device_name = $1, updated_at = $2
			WHERE device_id = $3
			RETURNING id, user_id, device_id, device_name, registered_at, last_sync_at, created_at, updated_at
		`, req.DeviceName, time.Now(), req.DeviceID).Scan(
			&device.ID, &device.UserID, &device.DeviceID, &device.DeviceName,
			&device.RegisteredAt, &device.LastSyncAt, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
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

	// Create new device
	var device models.Device
	now := time.Now()
	err = db.DB.QueryRow(`
		INSERT INTO devices (user_id, device_id, device_name, registered_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, device_id, device_name, registered_at, last_sync_at, created_at, updated_at
	`, userID, req.DeviceID, req.DeviceName, now, now, now).Scan(
		&device.ID, &device.UserID, &device.DeviceID, &device.DeviceName,
		&device.RegisteredAt, &device.LastSyncAt, &device.CreatedAt, &device.UpdatedAt,
	)
	if err != nil {
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
