package commands

import "strings"

// StringArr defines the custom flag type
// that represents an array of string values
type StringArr []string

// String is a required output method for the flag
func (s *StringArr) String() string {
	if len(*s) <= 0 {
		return "..."
	}

	return strings.Join(*s, ", ")
}

// Set is a required output method for the flag.
// This is where our custom type manipulation actually happens
func (s *StringArr) Set(value string) error {
	*s = append(*s, value)

	return nil
}
