package domain

import "fmt"

// NotFoundError adalah error khusus untuk entitas yang tidak ditemukan
type NotFoundError struct {
    Entity string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found", e.Entity)
}

// Helper untuk cek apakah error adalah NotFoundError
func IsNotFoundError(err error) bool {
    _, ok := err.(*NotFoundError)
    return ok
}

// Contoh error spesifik untuk Agent
var ErrAgentNotFound = &NotFoundError{Entity: "agent"}

// Tambahkan error khusus lain sesuai kebutuhan, misal:
// var ErrUserNotFound = &NotFoundError{Entity: "user"}
