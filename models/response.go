package models

// APIResponse is a standard response format for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *APIError   `json:"error"`
}

// APIError represents an error in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Common error codes
const (
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeDatabaseError    = "DATABASE_ERROR"
)

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string) APIResponse {
	return APIResponse{
		Success: false,
		Data:    nil,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
}
