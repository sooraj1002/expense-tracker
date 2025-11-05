package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID   `json:"userId" gorm:"type:uuid;column:user_id;not null;index"`
	Amount      float64     `json:"amount" gorm:"type:decimal(15,2);not null" binding:"required,gt=0"`
	CategoryID  uuid.UUID   `json:"categoryId" gorm:"type:uuid;column:category_id;not null;index" binding:"required"`
	AccountID   uuid.UUID   `json:"accountId" gorm:"type:uuid;column:account_id;not null;index" binding:"required"`
	Date        time.Time   `json:"date" gorm:"not null;index" binding:"required"`
	Description string      `json:"description,omitempty" gorm:"type:text"`
	MerchantID  *uuid.UUID  `json:"merchantId,omitempty" gorm:"type:uuid;column:merchant_id;index"`
	LocationID  *uuid.UUID  `json:"locationId,omitempty" gorm:"type:uuid;column:location_id"`
	RawData     string      `json:"rawData,omitempty" gorm:"column:raw_data;type:text"`
	Tags        StringArray `json:"tags" gorm:"type:text[];default:'{misc}'"`
	Verified    bool        `json:"verified" gorm:"default:false"`
	CreatedAt   time.Time   `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updatedAt" gorm:"autoUpdateTime"`
}

type CreateExpenseRequest struct {
	Amount      float64   `json:"amount" binding:"required,gt=0"`
	CategoryID  uuid.UUID `json:"categoryId" binding:"required"`
	AccountID   uuid.UUID `json:"accountId" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
}

type UpdateExpenseRequest struct {
	Amount      *float64   `json:"amount" binding:"omitempty,gt=0"`
	CategoryID  *uuid.UUID `json:"categoryId"`
	AccountID   *uuid.UUID `json:"accountId"`
	Date        *time.Time `json:"date"`
	Description *string    `json:"description"`
	Tags        []string   `json:"tags"`
	Verified    *bool      `json:"verified"`
}

type BatchExpenseRequest struct {
	DeviceID string    `json:"deviceId" binding:"required"`
	Expenses []Expense `json:"expenses" binding:"required"`
}

type BatchExpenseResponse struct {
	Success    bool                `json:"success"`
	Synced     int                 `json:"synced"`
	Failed     int                 `json:"failed"`
	Conflicts  int                 `json:"conflicts"`
	IDMappings map[string]string   `json:"idMappings"`
}

// CategoryBasic contains basic category information embedded in expense responses
type CategoryBasic struct {
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
	Color        string `json:"color"`
	IsDefault    bool   `json:"isDefault"`
}

// ExpenseResponse is the response structure for a single expense with embedded category
type ExpenseResponse struct {
	ID          string         `json:"id"`
	UserID      string         `json:"userId"`
	Amount      float64        `json:"amount"`
	Category    CategoryBasic  `json:"category"`
	AccountID   string         `json:"accountId"`
	Date        time.Time      `json:"date"`
	Description string         `json:"description,omitempty"`
	Tags        StringArray    `json:"tags"`
	Verified    bool           `json:"verified"`
	MerchantID  *string        `json:"merchantId,omitempty"`
	LocationID  *string        `json:"locationId,omitempty"`
	RawData     string         `json:"rawData,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// ExpenseListResponse is the new response structure with success field and embedded data
type ExpenseListResponse struct {
	Success    bool                `json:"success"`
	Data       []ExpenseResponse   `json:"data"`
	Pagination *PaginationMetadata `json:"pagination,omitempty"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Page        int     `json:"page"`
	Limit       int     `json:"limit"`
	TotalCount  int     `json:"totalCount"`
	TotalPages  int     `json:"totalPages"`
	TotalAmount float64 `json:"totalAmount"` // Total amount for filtered results
}

type PeriodSummary struct {
	Month        int     `json:"month"`
	Year         int     `json:"year"`
	TotalSpent   float64 `json:"totalSpent"`
	ExpenseCount int     `json:"expenseCount"`
}

// StringArray is a custom type for handling PostgreSQL text arrays
type StringArray []string

// Scan implements the sql.Scanner interface for StringArray
func (s *StringArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		// PostgreSQL returns arrays as {item1,item2,item3}
		return s.parseArray(string(src))
	case string:
		return s.parseArray(src)
	case nil:
		*s = []string{"misc"}
		return nil
	}
	return errors.New("incompatible type for StringArray")
}

// parseArray parses PostgreSQL array format {item1,item2}
func (s *StringArray) parseArray(str string) error {
	if str == "" || str == "{}" {
		*s = []string{"misc"}
		return nil
	}

	// Remove surrounding braces
	if len(str) > 2 && str[0] == '{' && str[len(str)-1] == '}' {
		str = str[1 : len(str)-1]
	}

	if str == "" {
		*s = []string{"misc"}
		return nil
	}

	// Simple split - this works for simple strings without commas in values
	parts := []string{}
	for _, part := range splitArray(str) {
		if part != "" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		*s = []string{"misc"}
	} else {
		*s = parts
	}
	return nil
}

// splitArray splits on commas while respecting quotes
func splitArray(s string) []string {
	var result []string
	var current string
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
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

// Value implements the driver.Valuer interface for StringArray
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{misc}", nil
	}

	// Format as PostgreSQL array: {item1,item2,item3}
	result := "{"
	for i, v := range s {
		if i > 0 {
			result += ","
		}
		// Quote strings that contain special characters
		if needsQuoting(v) {
			result += "\"" + v + "\""
		} else {
			result += v
		}
	}
	result += "}"
	return result, nil
}

// needsQuoting checks if a string needs to be quoted in PostgreSQL array
func needsQuoting(s string) bool {
	for _, c := range s {
		if c == ',' || c == '{' || c == '}' || c == '"' || c == ' ' {
			return true
		}
	}
	return false
}

// GormDataType gorm common data type
func (StringArray) GormDataType() string {
	return "text[]"
}
