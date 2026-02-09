package boards

import (
	"strings"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

type (
	// Flag defines a type for post flags
	Flag struct {
		// User is the user that flagged the post.
		User address

		// Reason is the reason that describes why post is flagged.
		Reason string
	}

	// FlagIterFn defines a function type to iterate post flags.
	FlagIterFn func(Flag) bool

	// FlagStorage defines an interface for storing posts flagging information.
	FlagStorage interface {
		// Exists checks if a flag from a user exists
		Exists(address) bool

		// Add adds a new flag from a user.
		Add(Flag) error

		// Remove removes a user flag.
		Remove(address) (removed bool)

		// Size returns the number of flags in the storage.
		Size() int

		// Iterate iterates post flags.
		// To reverse iterate flags use a negative count.
		// If the callback returns true, the iteration is stopped.
		Iterate(start, count int, fn FlagIterFn) bool
	}
)

// NewFlagStorage creates a new storage for post flags.
// The new storage uses an AVL tree to store flagging info.
func NewFlagStorage() FlagStorage {
	return &flagStorage{avl.NewTree()}
}

type flagStorage struct {
	flags *avl.Tree // address -> string(reason)
}

// Exists checks if a flag from a user exists
func (s flagStorage) Exists(addr address) bool {
	return s.flags.Has(addr.String())
}

// Add adds a new flag from a user.
// It fails if a flag from the same user exists.
func (s *flagStorage) Add(f Flag) error {
	if !f.User.IsValid() {
		return ufmt.Errorf("post flagging error, invalid user address: %s", f.User)
	}

	k := f.User.String()
	if s.flags.Has(k) {
		return ufmt.Errorf("flag from user already exists: %s", f.User)
	}

	s.flags.Set(k, strings.TrimSpace(f.Reason))
	return nil
}

// Remove removes a user flag.
func (s *flagStorage) Remove(addr address) bool {
	_, removed := s.flags.Remove(addr.String())
	return removed
}

// Size returns the number of flags in the storage.
func (s flagStorage) Size() int {
	return s.flags.Size()
}

// Iterate iterates post flags.
// To reverse iterate flags use a negative count.
// If the callback returns true, the iteration is stopped.
func (s flagStorage) Iterate(start, count int, fn FlagIterFn) bool {
	if count < 0 {
		return s.flags.ReverseIterateByOffset(start, -count, func(k string, v any) bool {
			return fn(Flag{
				User:   address(k),
				Reason: v.(string),
			})
		})
	}

	return s.flags.IterateByOffset(start, count, func(k string, v any) bool {
		return fn(Flag{
			User:   address(k),
			Reason: v.(string),
		})
	})
}
