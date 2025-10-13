package models

import (
	"time"

	"github.com/google/uuid"
)

type SyncStatus struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"userId" db:"user_id"`
	DeviceID           string     `json:"deviceId" db:"device_id"`
	DeviceName         string     `json:"deviceName" db:"device_name"`
	LastSyncTime       *time.Time `json:"lastSyncTime,omitempty" db:"last_sync_time"`
	LastSyncType       string     `json:"lastSyncType,omitempty" db:"last_sync_type"`
	PendingCount       int        `json:"pendingCount" db:"pending_count"`
	SyncedCount        int        `json:"syncedCount" db:"synced_count"`
	Status             string     `json:"status" db:"status"`
	ErrorMessage       string     `json:"errorMessage,omitempty" db:"error_message"`
	ConflictsResolved  int        `json:"conflictsResolved" db:"conflicts_resolved"`
	CreatedAt          time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at"`
}

type Device struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	DeviceID     string     `json:"deviceId" db:"device_id" binding:"required"`
	DeviceName   string     `json:"deviceName" db:"device_name" binding:"required"`
	RegisteredAt time.Time  `json:"registeredAt" db:"registered_at"`
	LastSyncAt   *time.Time `json:"lastSyncAt,omitempty" db:"last_sync_at"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

type RegisterDeviceRequest struct {
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName" binding:"required"`
}

type IncrementalSyncRequest struct {
	DeviceID          string                    `json:"deviceId" binding:"required"`
	LastSyncTimestamp *time.Time                `json:"lastSyncTimestamp"`
	Changes           IncrementalSyncChanges    `json:"changes"`
}

type IncrementalSyncChanges struct {
	Transactions      []Transaction      `json:"transactions"`
	Expenses          []Expense          `json:"expenses"`
	MerchantPatterns  []MerchantPattern  `json:"merchantPatterns"`
}

type IncrementalSyncResponse struct {
	Success       bool                      `json:"success"`
	SyncTimestamp time.Time                 `json:"syncTimestamp"`
	Conflicts     []SyncConflict            `json:"conflicts"`
	IDMappings    SyncIDMappings            `json:"idMappings"`
	ServerChanges IncrementalSyncChanges    `json:"serverChanges"`
}

type SyncConflict struct {
	Type       string `json:"type"`
	LocalID    string `json:"localId"`
	ServerID   string `json:"serverId"`
	Resolution string `json:"resolution"`
	Reason     string `json:"reason"`
}

type SyncIDMappings struct {
	Transactions     map[string]string `json:"transactions"`
	Expenses         map[string]string `json:"expenses"`
	MerchantPatterns map[string]string `json:"merchantPatterns"`
}

type UpdateSyncStatusRequest struct {
	DeviceID          string `json:"deviceId" binding:"required"`
	SyncType          string `json:"syncType" binding:"required,oneof=realtime batch manual"`
	SyncedCount       int    `json:"syncedCount"`
	PendingCount      int    `json:"pendingCount"`
	ConflictsResolved int    `json:"conflictsResolved"`
}
