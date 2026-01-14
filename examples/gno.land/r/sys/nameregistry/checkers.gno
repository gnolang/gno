package nameregistry

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"gno.land/p/nt/avl/rotree"
)

// Checks that a string ends with with 3 numbers.
var reEndsWithThreeNumbers = regexp.MustCompile(`[0-9]{3}$`)

// Checker defines an interface for name checkers.
type Checker interface {
	// Description returns a short description of the name checker.
	Description() string

	// Check checks that a name is valid.
	Check(name string) error
}

// NewChecker creates a new name checker.
func NewChecker(description string, fn func(string) error) Checker {
	description = strings.TrimSpace(description)
	if description == "" {
		panic("checker description is required")
	}

	if fn == nil {
		panic("checker function is required")
	}

	return &checker{description, fn}
}

// NewMinLengthChecker creates a checker to checks  min length of names.
func NewMinLengthChecker(length int) Checker {
	if length == 0 {
		panic("min length checker requires a length > 0")
	}

	desc := "Minimum name length allowed is " + strconv.Itoa(length)
	return NewChecker(desc, func(name string) error {
		if len(name) < length {
			return errors.New("name is too short, min length is " + strconv.Itoa(length))
		}
		return nil
	})
}

// NewMaxLengthChecker creates a checker to checks max length of names.
func NewMaxLengthChecker(length int) Checker {
	if length < 1 {
		panic("max length checker requires a length > 1")
	}

	desc := "Maximum name length allowed is " + strconv.Itoa(length)
	return NewChecker(desc, func(name string) error {
		if len(name) > length {
			return errors.New("name is too long, max length is " + strconv.Itoa(length))
		}
		return nil
	})
}

// NewLowercaseChecker creates a checker to check that all alphanumeric
// chars are lowercase.
func NewLowercaseChecker() Checker {
	desc := "Alphanumeric chars must all have lowercase chars"
	return NewChecker(desc, func(name string) error {
		if strings.ToLower(name) != name {
			return errors.New("name characters must be in lowercase")
		}
		return nil
	})
}

// NewForbiddenPrefixChecker creates a checker to checks that name doesn't
// start with one or more prefixes.
func NewForbiddenPrefixChecker(prefixes ...string) Checker {
	if len(prefixes) == 0 {
		panic("name prefix checker requires at least one prefix")
	}

	forbidden := strings.Join(prefixes, ", ")
	desc := "Name must not start with: " + forbidden
	return NewChecker(desc, func(name string) error {
		for _, s := range prefixes {
			if strings.HasPrefix(name, s) {
				return errors.New("name must not start these prefixes: " + forbidden)
			}
		}
		return nil
	})
}

// NewAlphaNumChecker creates a checker to check that a name starts with
// a letter and ends with three numbers.
func NewAlphaNumChecker() Checker {
	desc := "Name must start with a letter and end with 3 numbers"
	return NewChecker(desc, func(name string) error {
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

// NewHomoglyphChecker creates a checker to check that a name is not reserved
// or blacklisted. Checker considers typosquatting with o=0, i=1 and i=l.
// Names must be given within one or more AVL trees where key values are names.
func NewHomoglyphChecker(names ...rotree.IReadOnlyTree) Checker {
	if len(names) == 0 {
		panic("homoglyph checker requires at least one names tree")
	}

	reNumbersSuffix := regexp.MustCompile(`[0-9]{1}[a-z0-9]*$`)
	desc := "Reserved or blacklisted names are not allowed"
	return NewChecker(desc, func(name string) error {
		variants := []string{name}

		// Remove all numbers from name prefix to avoid cases like "jesus000"
		prefix := reNumbersSuffix.ReplaceAllString(name, "")
		if prefix != name {
			variants = append(variants, prefix)
		}

		variants = append(variants, GenHomoglyphVariants(name)...)
		for i, n := range variants {
			for _, storage := range names {
				if !storage.Has(n) {
					continue
				}

				// First name in list is the one being checked
				if i == 0 {
					return errors.New("name is reserved")
				} else {
					return errors.New("similar name already exists: " + n)
				}
			}
		}
		return nil
	})
}

type checker struct {
	description string
	fn          func(string) error
}

func (c checker) Description() string     { return c.description }
func (c checker) Check(name string) error { return c.fn(name) }
