package v1

import (
	"regexp"
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"

	"gno.land/r/sys/users"
)

const (
	adminAddr = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul

	reValidUsername    = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
	RegisterEvent      = "Registered"
	RegisterAliasEvent = "AliasRegistered"
	DeleteUserEvent    = "UserDeleted"
)

var (
	Ownable    = ownable.NewWithAddress(adminAddr) // @moul - switch to dao managed in v2
	Pausable   = pausable.NewFromOwnable(Ownable)  //
	reUsername = regexp.MustCompile(reValidUsername)
)

// Register registers a new username for the caller
// A valid username must start with a minimum of 3 letters,
// end with a minimum of 3 numbers, and be less than 20 chars long.
// All letters must be lowercase, and the only valid special char is `_`
// Only calls from EOAs are supported
func Register(username string) error {
	std.AssertOriginCall()

	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	if matched := reUsername.MatchString(username); !matched {
		return ErrInvalidUsername
	}

	registrant := std.PrevRealm().Addr()

	if err := users.Register(username, registrant); err != nil {
		return err
	}

	std.Emit(RegisterEvent, "registrant", registrant.String(), "username", username)
	return nil
}

func RegisterAlias(alias string) error {
	std.AssertOriginCall()

	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	registrant := std.PrevRealm().Addr()

	if err := users.RegisterAlias(username, registrant); err != nil {
		return err
	}

	std.Emit(RegisterAliasEvent, "registrant", registrant.String(), "alias", alias)
	return nil
}

// DeleteUser deletes the name and aliases associated with the PrevRealm() address.
// WARN: After deletion, the same address WILL NOT be able to register a new name or alias.
func DeleteUser() error {
	std.AssertOriginCall()

	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	addr := std.PrevRealm().Addr()
	if err := users.Delete(addr); err != nil {
		return err
	}

	std.Emit(DeleteUserEvent, "address", registrant.String())
	return nil
}
