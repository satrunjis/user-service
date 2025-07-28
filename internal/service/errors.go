package service

import (
	"errors"
	"strings"
)

type ErrorCode string

const (
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
)

type ServiceError struct {
	Code    ErrorCode
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewServiceError(code ErrorCode, messages ...string) *ServiceError {
	message := strings.Join(messages, " ")
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

func mapRepositoryError(err error, operation string) error {
	_ = operation
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		switch serviceErr.Code {
		case ErrCodeNotFound:
			return NewServiceError(ErrCodeNotFound, "User not found")
		case ErrCodeAlreadyExists:
			return NewServiceError(ErrCodeAlreadyExists, "User already exists")
		case ErrCodeInternal:
			return NewServiceError(ErrCodeInternal, "Database error occurred")
		default:
			return NewServiceError(ErrCodeInternal, "Unexpected error occurred")
		}
	}

	return NewServiceError(ErrCodeInternal, "System error occurred")
}
