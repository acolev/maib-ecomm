package maib

import (
	"fmt"
	"strings"
)

// APIError represents a structured error from the Maib API.
type APIError struct {
	Errors []ErrorItem
}

func (e *APIError) Error() string {
	var msgs []string
	for _, err := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("[%s] %s", err.ErrorCode, err.ErrorMessage))
	}
	return fmt.Sprintf("API Error: %s", strings.Join(msgs, "; "))
}
