package proposal

import "strings"

// CombinedErrors is a combined execution error
type CombinedErrors struct {
	errors []error
}

// Error returns the combined execution error
func (e *CombinedErrors) Error() string {
	if len(e.errors) == 0 {
		return ""
	}

	var sb strings.Builder

	for _, err := range e.errors {
		sb.WriteString(err.Error() + "; ")
	}

	// Remove the last semicolon and space
	result := sb.String()

	return result[:len(result)-2]
}

// Add adds a new error to the execution error
func (e *CombinedErrors) Add(errs ...error) {
	if len(errs) == 0 {
		return
	}

	e.errors = append(e.errors, errs...)
}

// Size returns a
func (e *CombinedErrors) Size() int {
	return len(e.errors)
}
