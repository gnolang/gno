package v1

import (
	"regexp"
	"std"

	"gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
	RegisterEvent   = "Registered"
	UpdateNameEvent = "NameUpdated"
	DeleteUserEvent = "UserDeleted"
)

var reUsername = regexp.MustCompile(reValidUsername)

// Register registers a new username for the caller
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`.
// Only calls from EOAs are supported.
func Register(username string) error {
	std.AssertOriginCall()

	if paused {
		return ErrPaused
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	registrant := std.PrevRealm().Addr()
	if err := users.RegisterUser(username, registrant); err != nil {
		return err
	}

	std.Emit(RegisterEvent, "registrant", registrant.String(), "username", username)
	return nil
}

// UpdateName allows a user to update their name.
// The associated address and all previous names of a user that changes a name
// are preserved, and all resolve to the new name.
func UpdateName(newName string) error {
	std.AssertOriginCall()

	if paused {
		return ErrPaused
	}

	registrant := std.PrevRealm().Addr()
	if err := users.UpdateName(newName, registrant); err != nil {
		return err
	}

	std.Emit(UpdateNameEvent, "registrant", registrant.String(), "newName", newName)
	return nil
}

// DeleteUser makes all names associated with the PrevRealm() address unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func DeleteUser() error {
	std.AssertOriginCall()

	if paused {
		return ErrPaused
	}

	addr := std.PrevRealm().Addr()
	if err := users.Delete(addr); err != nil {
		return err
	}

	std.Emit(DeleteUserEvent, "address", addr.String())
	return nil
}
