package host

import (
	"errors"
	"regexp"
	"strings"

	"gno.land/p/nt/ufmt"
)

// DefaultMaxCharacterLength defines the default maximum character length used
// in validation of identifiers including the client, connection, port and
// channel identifiers.
//
// NOTE: this restriction is specific to this golang implementation of IBC. If
// your use case demands a higher limit, please open an issue and we will consider
// adjusting this restriction.
const DefaultMaxCharacterLength = 64

// DefaultMaxPortCharacterLength defines the default maximum character length used
// in validation of port identifiers.
var DefaultMaxPortCharacterLength = 128

// IsValidID defines regular expression to check if the string consist of
// characters in one of the following categories only:
// - Alphanumeric
// - `.`, `_`, `+`, `-`, `#`
// - `[`, `]`, `<`, `>`
var IsValidID = regexp.MustCompile(`^[a-zA-Z0-9\.\_\+\-\#\[\]\<\>]+$`).MatchString

// ICS 024 Identifier and Path Validation Implementation
//
// This file defines ValidateFn to validate identifier and path strings
// The spec for ICS 024 can be located here:
// https://github.com/cosmos/ibc/tree/master/spec/core/ics-024-host-requirements

// ValidateFn function type to validate path and identifier bytestrings
type ValidateFn func(string) error

func defaultIdentifierValidator(id string, minLength, maxLength int) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("identifier cannot be blank")
	}
	// valid id MUST NOT contain "/" separator
	if strings.Contains(id, "/") {
		return ufmt.Errorf("identifier %s cannot contain separator '/'", id)
	}
	// valid id must fit the length requirements
	if len(id) < minLength || len(id) > maxLength {
		return ufmt.Errorf("identifier %s has invalid length: %d, must be between %d-%d characters", id, len(id), minLength, maxLength)
	}
	// valid id must contain only alphanumeric characters and some allowed symbols.
	if !IsValidID(id) {
		return ufmt.Errorf(
			"identifier %s must contain only alphanumeric or the following characters: '.', '_', '+', '-', '#', '[', ']', '<', '>'",
			id,
		)
	}
	return nil
}

// ClientIdentifierValidator is the default validator function for Client identifiers.
// A valid Identifier must be between 4-64 characters and only contain alphanumeric and some allowed
// special characters (see IsValidID).
func ClientIdentifierValidator(id string) error {
	err := defaultIdentifierValidator(id, 4, DefaultMaxCharacterLength)
	if err != nil {
		return ufmt.Errorf("validate client identifier: %v", err)
	}
	return nil
}

// ConnectionIdentifierValidator is the default validator function for Connection identifiers.
// A valid Identifier must be between 10-64 characters and only contain alphanumeric and some allowed
// special characters (see IsValidID).
// TODO probably not needed remove
func ConnectionIdentifierValidator(id string) error {
	err := defaultIdentifierValidator(id, 10, DefaultMaxCharacterLength)
	if err != nil {
		return ufmt.Errorf("validate connection identifier: %v", err)
	}
	return nil
}

// ChannelIdentifierValidator is the default validator function for Channel identifiers.
// A valid Identifier must be between 8-64 characters and only contain alphanumeric and some allowed
// special characters (see IsValidID).
// TODO remove ? there's no channel in IBCv2
func ChannelIdentifierValidator(id string) error {
	err := defaultIdentifierValidator(id, 8, DefaultMaxCharacterLength)
	if err != nil {
		return ufmt.Errorf("validate channel identifier: %v", err)
	}
	return nil
}

// PortIdentifierValidator is the default validator function for Port identifiers.
// A valid Identifier must be between 2-64 characters and only contain alphanumeric and some allowed
// special characters (see IsValidID).
func PortIdentifierValidator(id string) error {
	err := defaultIdentifierValidator(id, 2, DefaultMaxPortCharacterLength)
	if err != nil {
		return ufmt.Errorf("validate port identifier: %v", err)
	}
	return nil
}

// NewPathValidator takes in an Identifier Validator function and returns
// a Path Validator function which requires path to consist of `/`-separated valid identifiers,
// where a valid identifier is between 1-64 characters, contains only alphanumeric and some allowed
// special characters (see IsValidID), and satisfies the custom `idValidator` function.
func NewPathValidator(idValidator ValidateFn) ValidateFn {
	return func(path string) error {
		pathArr := strings.Split(path, "/")
		if len(pathArr) > 0 && pathArr[0] == path {
			return ufmt.Errorf("path %s doesn't contain any separator '/'", path)
		}

		for _, p := range pathArr {
			// a path beginning or ending in a separator returns empty string elements.
			if p == "" {
				return ufmt.Errorf("path %s cannot begin or end with '/'", path)
			}

			if err := idValidator(p); err != nil {
				return err
			}
			// Each path element must either be a valid identifier or constant number
			if err := defaultIdentifierValidator(p, 1, DefaultMaxCharacterLength); err != nil {
				return ufmt.Errorf("path %s contains an invalid identifier: '%s': %v", path, p, err)
			}
		}

		return nil
	}
}
