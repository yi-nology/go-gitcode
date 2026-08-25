package gitcode

import (
	"errors"
	"fmt"
)

// ValidationErrorField represents a client-side validation error.
type ValidationErrorField struct {
	Field   string
	Message string
}

func (e *ValidationErrorField) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// validateOwnerRepo validates that owner and repo are not empty.
func validateOwnerRepo(owner, repo string) error {
	if owner == "" {
		return &ValidationErrorField{Field: "owner", Message: "owner is required"}
	}
	if repo == "" {
		return &ValidationErrorField{Field: "repo", Message: "repo is required"}
	}
	return nil
}

// validateOwner validates that owner is not empty.
//
//nolint:unused
func validateOwner(owner string) error {
	if owner == "" {
		return &ValidationErrorField{Field: "owner", Message: "owner is required"}
	}
	return nil
}

// validateNonEmpty validates that a value is not empty.
func validateNonEmpty(field, value string) error {
	if value == "" {
		return &ValidationErrorField{Field: field, Message: field + " is required"}
	}
	return nil
}

// IsValidationError checks if an error is a client-side validation error.
func IsValidationError(err error) bool {
	var target *ValidationErrorField
	return errors.As(err, &target)
}
