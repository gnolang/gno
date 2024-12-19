package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/pausable"
)

var (
	usernameStore   = avl.NewTree() // username > address
	addressStore    = avl.NewTree() // address > username
	aliasStore      = avl.NewTree() // username > address
	callerWhitelist = avl.NewTree() // addr > struct{}{}

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
)

//type UserData struct {
//	name      string
//	addr      std.Address
//	IsAlias   bool
//	IsDeleted bool
//}
//
//var (
//	usernameStore = avl.NewTree() // username > *UserData
//	addressStore  = avl.NewTree() // address > *UserData
//) todo see if this is the better approach

const (
	SetUserEvent     = "UserSet"
	CreateAliasEvent = "AliasCreated"
	DeleteUserEvent  = "UserDeleted"

	maxUsernameLen = 50
	deletedUser    = "ghost" // todo also need to forbid as username... not cool
)

// SetUser writes a *new* user to the store
// Only whitelisted callers can call this function
// Writing to the store can be paused by the admin
func SetUser(name string, addr std.Address) error {
	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	trimmedName := strings.TrimSpace(name)

	// Check if addr is already registered or has been deleted
	raw, ok := addressStore.Has(addr.String())
	if ok {
		if raw.(string) == deletedUser {
			return ErrDeletedUser
		}

		return ErrAlreadyHasName // use CreateAlias
	}

	// Check if name is taken
	if usernameStore.Has(trimmedName) || aliasStore.Has(trimmedName) {
		return ErrNameTaken
	}

	// Validate fields before writing
	if err := validate(trimmedName, addr); err != nil {
		return err
	}

	usernameStore.Set(trimmedName, addr)
	addressStore.Set(addr.String(), trimmedName)

	std.Emit(SetUserEvent, "name", name, "address", addr.String())

	return nil
}

// CreateAlias creates an alias username for a specific address
// The address must already have a registered username
func CreateAlias(name string, addr std.Address) error {
	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	trimmedName := strings.TrimSpace(name)

	// Check if addr is already registered or has been deleted
	raw, ok := addressStore.Has(addr.String())
	if !ok {
		return ErrAliasBeforeName // must register beforehand
	}

	// Check if user has been deleted
	if raw.(string) == deletedUser {
		return ErrDeletedUser
	}

	// Cannot register taken username as alias
	if usernameStore.Has(trimmedName) || aliasStore.Has(trimmedName) {
		return ErrNameTaken
	}

	// Validate before writing
	if err := validate(trimmedName, addr); err != nil {
		return err
	}

	aliasStore.Set(trimmedName, addr)

	return nil
}

func DeleteUser(addr std.Address) error {
	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	// Check if user is registered
	raw, ok := addressStore.Get(addr.String())
	if !ok {
		return ErrUserNotExist
	}

	username := raw.(string)

	// Check if user has already been deleted
	if username == deletedUser {
		return ErrAlreadyDeleted
	}

	// Find all aliases to delete
	var toDelete []string
	aliasStore.Iterate("", "", func(key string, value interface{}) bool {
		if value.(std.Address) == addr {
			toDelete = append(toDelete, key)
		}

		return false
	})

	// Delete user
	usernameStore.Set(username, std.Address("")) // todo OR dstore as blacklist
	addressStore.Set(addr, deletedUser)
	for _, alias := range toDelete {
		aliasStore.Set(alias, deletedUser)
	}

	std.Emit(DeleteUserEvent, "deletedUser", username)

	return nil
}

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr.String())
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
