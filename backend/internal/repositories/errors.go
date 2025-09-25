package repositories

import "errors"

var (
	// ErrNotFound indicates the requested record does not exist.
	ErrNotFound = errors.New("record not found")
	// ErrConflict indicates the attempted write would violate a uniqueness constraint.
	ErrConflict = errors.New("record conflict")
)
