package nameregistry

import (
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
)

// TODO: Allow updating validators
// TODO: Use tries to save storage

var (
	reserved    = avl.NewTree()
	blacklisted = avl.NewTree()
	custom      = avl.NewTree()

	// Name validators that defines the rules for valid names
	validators = []Validator{
		NewMinLengthValidator(4),
		NewMaxLengthValidator(14),
		NewLowercaseValidator(),
		NewForbiddenPrefixValidator("g1", "gi", "gl"),
		NewAlphaNumValidator(),
	}
)

// IsNameValid checks is a name is a valid name.
func IsNameValid(name string) error {
	for _, v := range validators {
		if err := v.Validate(name); err != nil {
			return err
		}
	}
	return nil
}

// GetReservedNames returns reserved registry names.
func GetReservedNames() rotree.IReadOnlyTree {
	return rotree.Wrap(reserved, func(v any) any { return struct{}{} })
}

// GetBlacklistedNames returns blacklisted registry names.
func GetBlacklistedNames() rotree.IReadOnlyTree {
	return rotree.Wrap(blacklisted, func(v any) any { return struct{}{} })
}

// GetCustomNames returns custom registry names.
func GetCustomNames() rotree.IReadOnlyTree {
	return rotree.Wrap(custom, func(v any) any { return struct{}{} })
}

func Render(string) string {
	var b strings.Builder

	b.WriteString(md.H1("Name Registry"))
	b.WriteString(md.H2("Validation"))
	b.WriteString(md.Paragraph("Names must match the following rules:"))

	for _, v := range validators {
		b.WriteString(md.BulletItem(v.Description()))
	}

	return b.String()
}
