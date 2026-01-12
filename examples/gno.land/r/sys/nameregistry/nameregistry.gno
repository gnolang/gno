package nameregistry

import (
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
)

// TODO: Allow updating checkers (reset, replace, append), GovDAO Tier1 only
// TODO: Use tries to save storage

var (
	reserved    = avl.NewTree()
	blacklisted = avl.NewTree()
	custom      = avl.NewTree()

	// Name constraints contains checkers that defines the rules for valid names
	constraints = []Checker{
		NewMinLengthChecker(4),
		NewMaxLengthChecker(14),
		NewLowercaseChecker(),
		NewForbiddenPrefixChecker("g1", "gi", "gl"),
		NewAlphaNumChecker(),
	}
)

// IsNameValid checks is a name is a valid name.
func IsNameValid(name string) error {
	for _, c := range constraints {
		if err := c.Check(name); err != nil {
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
	b.WriteString(md.H2("Name Checks"))
	b.WriteString(md.Paragraph("Names must match the following rules:"))

	items := make([]string, len(constraints))
	for _, c := range constraints {
		items = append(items, c.Description())
	}

	b.WriteString(md.OrderedList(items))

	return b.String()
}
