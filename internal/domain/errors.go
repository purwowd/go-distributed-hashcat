package domain

import "fmt"

// NotFoundError is a custom error for entities that are not found
type NotFoundError struct {
	Entity string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Entity)
}

// Helper to check if error is NotFoundError
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// Example specific error for Agent
var ErrAgentNotFound = &NotFoundError{Entity: "agent"}

// Add other custom errors as needed, for example:
// var ErrUserNotFound = &NotFoundError{Entity: "user"}
