package nameregistry

import (
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/rotree"
	"gno.land/p/nt/ownable"
)

// TODO: Use tries to save storage

const (
	TypeReserved    = "reserved"
	TypeBlacklisted = "blacklisted"
	TypeCustom      = "custom"
)

var (
	reserved    = avl.NewTree()
	blacklisted = avl.NewTree()
	custom      = avl.NewTree()

	// TODO: Remove once realm fully integrates with GovDAO
	admin = ownable.NewWithOrigin()

	// Name constraints contains checkers that defines the rules for valid names
	constraints = []Checker{
		NewMinLengthChecker(4),
		NewMaxLengthChecker(14),
		NewLowercaseChecker(),
		NewForbiddenPrefixChecker("g1", "gi", "gl"),
		NewAlphaNumChecker(),
		NewSimilarityChecker(),
	}
)

// Size returns the number of names in the registry.
func Size() int {
	return reserved.Size() + blacklisted.Size() + custom.Size()
}

// IsNameValid checks is a name is a valid name.
func IsNameValid(name string) error {
	for _, c := range constraints {
		if err := c.Check(name); err != nil {
			return err
		}
	}
	return nil
}

// GetNames returns reserved, blacklisted or custom registry names.
func GetNames(nameType string) rotree.IReadOnlyTree {
	var storage *avl.Tree
	switch nameType {
	case TypeReserved:
		storage = reserved
	case TypeBlacklisted:
		storage = blacklisted
	case TypeCustom:
		storage = custom
	default:
		panic("invalid name type: " + nameType)
	}
	return rotree.Wrap(storage, func(v any) any { return struct{}{} })
}

// SetNames sets the list of reserved, blacklisted or custom registry names.
// Existing names are replaced by the new ones.
func SetNames(_ realm, nameType string, names ...string) {
	if len(names) == 0 {
		panic("list of names must not be empty")
	}

	// TODO: Names be setted though GovDAO proposals
	admin.AssertOwnedByPrevious()

	storage := avl.NewTree()
	switch nameType {
	case TypeReserved:
		reserved = storage
	case TypeBlacklisted:
		blacklisted = storage
	case TypeCustom:
		custom = storage
	default:
		panic("invalid name type: " + nameType)
	}

	for _, n := range names {
		storage.Set(n, struct{}{})
	}
}

// AddNames adds new names to the list of reserved, blacklisted or custom registry names.
// Adding new names keeps the existing names intact.
func AddNames(_ realm, nameType string, names ...string) {
	if len(names) == 0 {
		panic("list of names must not be empty")
	}

	// TODO: Names be added though GovDAO proposals
	admin.AssertOwnedByPrevious()

	storage := avl.NewTree()
	switch nameType {
	case TypeReserved:
		reserved = storage
	case TypeBlacklisted:
		blacklisted = storage
	case TypeCustom:
		custom = storage
	default:
		panic("invalid name type: " + nameType)
	}

	for _, n := range names {
		storage.Set(n, struct{}{})
	}
}
