package params

import "fmt"

// MustParamString asserts value is a string and returns it.
// Panics with a descriptive message if the type assertion fails.
func MustParamString(key string, value any) string {
	s, ok := value.(string)
	if !ok {
		panic(fmt.Sprintf("invalid type for %s param: expected string, got %T", key, value))
	}
	return s
}

// MustParamInt64 asserts value is an int64 and returns it.
// Panics with a descriptive message if the type assertion fails.
func MustParamInt64(key string, value any) int64 {
	i, ok := value.(int64)
	if !ok {
		panic(fmt.Sprintf("invalid type for %s param: expected int64, got %T", key, value))
	}
	return i
}

// MustParamStrings asserts value is a []string and returns it.
// Panics with a descriptive message if the type assertion fails.
func MustParamStrings(key string, value any) []string {
	s, ok := value.([]string)
	if !ok {
		panic(fmt.Sprintf("invalid type for %s param: expected []string, got %T", key, value))
	}
	return s
}
