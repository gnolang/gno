package nameregistry

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// Checks that a string ends with with 3 numbers.
var reEndsWithThreeNumbers = regexp.MustCompile(`[0-9]{3}$`)

// Validator defines an interface for name validators.
type Validator interface {
	// Description returns a short description of the name validator.
	Description() string

	// Validate checks that a name is valid.
	Validate(name string) error
}

// NewValidator creates a new name validator.
func NewValidator(description string, fn func(string) error) Validator {
	description = strings.TrimSpace(description)
	if description == "" {
		panic("validator description is required")
	}

	if fn == nil {
		panic("validator function is required")
	}

	return &validator{description, fn}
}

// NewMinLengthValidator creates a validator to checks  min length of names.
func NewMinLengthValidator(length int) Validator {
	if length == 0 {
		panic("min length validator requires a length > 0")
	}

	desc := "Minimum name length allowed is " + strconv.Itoa(length)
	return NewValidator(desc, func(name string) error {
		if len(name) < length {
			return errors.New("name is too short, min length is " + strconv.Itoa(length))
		}
		return nil
	})
}

// NewMaxLengthValidator creates a validator to checks max length of names.
func NewMaxLengthValidator(length int) Validator {
	if length < 1 {
		panic("max length validator requires a length > 1")
	}

	desc := "Maximum name length allowed is " + strconv.Itoa(length)
	return NewValidator(desc, func(name string) error {
		if len(name) > length {
			return errors.New("name is too long, max length is " + strconv.Itoa(length))
		}
		return nil
	})
}

// NewLowercaseValidator creates a validator to check that all alphanumeric
// chars are lowercase.
func NewLowercaseValidator() Validator {
	desc := "Alphanumeric chars must all have lowercase chars"
	return NewValidator(desc, func(name string) error {
		if strings.ToLower(name) != name {
			return errors.New("name characters must be in lowercase")
		}
		return nil
	})
}

// NewForbiddenPrefixValidator creates a validator to checks that name doesn't
// start with one or more prefixes.
func NewForbiddenPrefixValidator(prefixes ...string) Validator {
	if len(prefixes) == 0 {
		panic("name prefix validator requires at least one prefix")
	}

	forbidden := strings.Join(prefixes, ", ")
	desc := "Name must not start with: " + forbidden
	return NewValidator(desc, func(name string) error {
		for _, s := range prefixes {
			if strings.HasPrefix(name, s) {
				return errors.New("name must not start these prefixes: " + forbidden)
			}
		}
		return nil
	})
}

// NewAlphaNumValidator creates a validator to check that a name starts with
// a letter and ends with three numbers.
func NewAlphaNumValidator() Validator {
	desc := "Name must start with a letter and end with 3 numbers"
	return NewValidator(desc, func(name string) error {
		name = strings.ToLower(name)
		if name == "" || name[0] > 'z' || name[0] < 'a' {
			return errors.New(`name must start with a letter between "a" and "z"`)
		}

		if len(name) < 4 || !reEndsWithThreeNumbers.MatchString(name) {
			return errors.New("name must end with three numbers")
		}
		return nil
	})
}

type validator struct {
	description string
	fn          func(string) error
}

func (v validator) Description() string        { return v.description }
func (v validator) Validate(name string) error { return v.fn(name) }
