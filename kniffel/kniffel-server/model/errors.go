package model

import "fmt"

// ErrBadRequest represents a 400 error
type ErrBadRequest struct {
	Message string
}

func (e *ErrBadRequest) Error() string {
	return e.Message
}

// ErrNotFound represents a 404 error
type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}
