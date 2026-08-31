package hub

import (
	"gno.land/p/gnoland/boards"
)

// Flag defines a type for thread and comment flags.
type Flag struct {
	// user is the user that flagged.
	user address

	// reason is the reason for flagging.
	reason string
}

// User returns the user that flagged.
func (f Flag) User() address { return f.user }

// Reason returns the reason for flagging.
func (f Flag) Reason() string { return f.reason }

// NewSafeFlag creates a safe flag.
func NewSafeFlag(ref boards.Flag) Flag {
	return Flag{
		user:   ref.User,
		reason: ref.Reason,
	}
}
