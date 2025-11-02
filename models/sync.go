package models

import (
	"time"

	"github.com/google/uuid"
)

type SyncStatus struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID  `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	DeviceID          string     `json:"deviceId" gorm:"column:device_id;not null"`
	DeviceName        string     `json:"deviceName" gorm:"column:device_name"`
	LastSyncTime      *time.Time `json:"lastSyncTime,omitempty" gorm:"column:last_sync_time"`
	LastSyncType      string     `json:"lastSyncType,omitempty" gorm:"column:last_sync_type"`
	PendingCount      int        `json:"pendingCount" gorm:"column:pending_count;default:0"`
	SyncedCount       int        `json:"syncedCount" gorm:"column:synced_count;default:0"`
	Status            string     `json:"status" gorm:"default:'pending'"`
	ErrorMessage      string     `json:"errorMessage,omitempty" gorm:"column:error_message;type:text"`
	ConflictsResolved int        `json:"conflictsResolved" gorm:"column:conflicts_resolved;default:0"`
	CreatedAt         time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

type Device struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	DeviceID     string     `json:"deviceId" gorm:"column:device_id;uniqueIndex;not null" binding:"required"`
	DeviceName   string     `json:"deviceName" gorm:"column:device_name;not null" binding:"required"`
	RegisteredAt time.Time  `json:"registeredAt" gorm:"column:registered_at"`
	LastSyncAt   *time.Time `json:"lastSyncAt,omitempty" gorm:"column:last_sync_at"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
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
