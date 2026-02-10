package commands

import (
	"fmt"
	"strconv"
	"strings"
)

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

// Uint64Flag defines the custom flag type
// that represents a uint64 value and a boolean
// indicating whether it was set or not
type Uint64Flag struct {
	V       uint64
	Defined bool
}

// String outputs the value of the flag
func (u *Uint64Flag) String() string {
	return strconv.FormatUint(u.V, 10)
}

// Set parses and sets the value of the flag
func (u *Uint64Flag) Set(value string) error {
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid uint64 value: %w", err)
	}

	u.V, u.Defined = v, true
	return nil
}
