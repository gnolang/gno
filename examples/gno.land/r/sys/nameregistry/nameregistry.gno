package nameregistry

import (
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
)

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
		NewHomoglyphChecker(reserved, custom, blacklisted),
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
