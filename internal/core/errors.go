package core

import "errors"

var (
	// ErrNotFound is returned when a requested domain entity cannot be located.
	ErrNotFound = errors.New("resource not found")

	// ErrConflict is returned when an entity with duplicate identifiers or conflicting state exists.
	ErrConflict = errors.New("resource conflict")

	// ErrInvalidInput is returned when payload data violates business invariants.
	ErrInvalidInput = errors.New("invalid input data")

	// ErrInvalidState is returned when attempting an illegal state machine transition.
	ErrInvalidState = errors.New("invalid state transition")

	// ErrUnauthorized is returned when authentication credentials are missing or invalid.
	ErrUnauthorized = errors.New("unauthorized access")

	// ErrForbidden is returned when the caller lacks necessary authorization privileges.
	ErrForbidden = errors.New("forbidden operation")
)
