package combinederr

import "strings"

// CombinedError is a combined execution error
type CombinedError struct {
	errors []error
}

// Error returns the combined execution error
func (e *CombinedError) Error() string {
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
func (e *CombinedError) Add(err error) {
	if err == nil {
		return
	}

	e.errors = append(e.errors, err)
}

// Size returns a
func (e *CombinedError) Size() int {
	return len(e.errors)
}
