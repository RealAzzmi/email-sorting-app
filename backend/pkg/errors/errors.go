package errors

import "fmt"

// Common application errors
var (
	ErrNotFound      = fmt.Errorf("resource not found")
	ErrInvalidInput  = fmt.Errorf("invalid input")
	ErrUnauthorized  = fmt.Errorf("unauthorized")
	ErrInternalError = fmt.Errorf("internal server error")
)

// Custom error types
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Err:     ErrNotFound,
	}
}

func NewInvalidInputError(message string) *AppError {
	return &AppError{
		Code:    "INVALID_INPUT",
		Message: message,
		Err:     ErrInvalidInput,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Err:     err,
	}
}