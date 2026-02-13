package antisquat

import (
	"regexp"
	"strings"
)

// validNamePattern matches names that start with a lowercase letter
// followed by lowercase letters, digits, or underscores.
var validNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// NormalizeName converts a name to its canonical form (lowercase, trimmed).
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ValidateName checks whether a name meets the format and length requirements.
func ValidateName(name string, cfg *Config) error {
	if name == "" {
		return ErrNameEmpty
	}

	nameLen := len(name)

	if nameLen < cfg.MinNameLength {
		return ErrNameTooShort
	}

	if nameLen > cfg.MaxNameLength {
		return ErrNameTooLong
	}

	if !validNamePattern.MatchString(name) {
		return ErrNameInvalidChars
	}

	return nil
}
