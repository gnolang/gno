package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	p "gno.land/p/demo/pausable"
)

var (
	usernameStore   = avl.NewTree() // username > address
	addressStore    = avl.NewTree() // address > username
	aliasStore      = avl.NewTree() // username > address
	callerWhitelist = avl.NewTree() // addr > struct{}{}

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{38}$`)
)

const maxUsernameLen = 50

// WriteUser writes a *new* user to the store
// Only whitelisted callers can call this function
// Writing to the store can be paused by the admin
func WriteUser(name string, addr std.Address) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	trimmedName := strings.TrimSpace(name)
	if err := validate(trimmedName, addr); err != nil {
		return err
	}

	if usernameStore.Has(trimmedName) {
		return ErrAlreadyExists
	}

	if addressStore.Has(addr) {
		return ErrAlreadyHasName // use CreateAlias
	}

	// Cannot register name which is someone's alias
	if aliasStore.Has(trimmedName) {
		return ErrAlreadyExists
	}

	return nil
}

// CreateAlias creates an alias username for a specific address
// The address must already have a registered username
func CreateAlias(name string, addr std.Address) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	trimmedName := strings.TrimSpace(name)
	if err := validate(trimmedName, addr); err != nil {
		return err
	}

	// Cannot register alias without an already registered username
	if !addressStore.Has(addr) {
		return ErrAliasBeforeName
	}

	// Cannot register taken username as alias
	if usernameStore.Has(trimmedName) || aliasStore.Has(trimmedName) {
		return ErrAlreadyExists
	}

	aliasStore.Set(trimmedName, addr)

	return nil
}

// XXX: should we have a DeleteUser? Maybe with a separate whitelist?

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr.String())
}

// UserCount returns the number of registered users
func UserCount() int {
	return aStore.Size()
}

// XXX: Should we have this? it's not modifiable directly.
func GetUserStore() *avl.Tree {
	return uStore
}

// Validate validates username and address passed in
func validate(username string, address std.Address) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	if username == "" {
		return ErrEmptyUsername
	}

	if len(username) > maxUsernameLen {
		return ErrUsernameTooLong
	}

	// Check if the username can be decoded or looks like a valid address
	if _, _, ok := std.DecodeBech32(std.Address(username)); ok || reAddressLookalike.MatchString(username) {
		return ErrUsernameLikeAddress
	}

	return nil
}
