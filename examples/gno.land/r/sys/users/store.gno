package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
)

var (
	nameStore    = avl.NewTree() // name/aliases > *UserData
	addressStore = avl.NewTree() // address > *UserData

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
)

const (
	RegisterUserEvent  = "UserRegistered"
	RegisterAliasEvent = "AliasRegistered"
	DeleteUserEvent    = "UserDeleted"

	maxUsernameLen = 64
)

type UserData struct {
	addr     std.Address
	username string // contains the latest name of a user
	deleted  bool
}

func (u UserData) Name() string {
	return u.username
}

func (u UserData) Addr() std.Address {
	return u.addr
}

func (u UserData) IsDeleted() bool {
	return u.deleted
}

// RegisterUser adds a new user to the system.
func RegisterUser(name string, address std.Address) error {
	// Validate caller
	if err := validateCall(); err != nil {
		return err
	}

	// Validate inputs
	if err := validateInputs(name, address); err != nil {
		return err
	}

	// Check if name is taken
	if nameStore.Has(name) {
		return ErrNameTaken
	}

	raw, ok := addressStore.Get(address.String())
	if ok {
		// Cannot re-register after deletion
		if raw.(*UserData).IsDeleted() {
			return ErrDeletedUser
		}

		// For a second name, user RegisterAlias
		return ErrAlreadyHasName
	}

	// Create UserData
	data := &UserData{
		addr:     address,
		username: name,
		deleted:  false,
	}

	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address.String(), data)

	std.Emit(RegisterUserEvent, "name", name, "address", address.String())
	return nil
}

// ChangeName adds an Alias pointing to an existing user.
func ChangeName(newName string, address std.Address) error {
	// Validate caller
	if err := validateCall(); err != nil {
		return err
	}

	// Validate inputs
	if err := validateInputs(newName, address); err != nil {
		return err
	}

	// Check if the requested Alias is already taken
	if nameStore.Has(newName) {
		return ErrNameTaken
	}

	// Check if user has a name before an Alias
	raw, ok := addressStore.Get(address.String())
	if !ok {
		return ErrAliasBeforeName
	}

	userData := raw.(*UserData)

	if userData.IsDeleted() {
		return ErrDeletedUser
	}

	userData.username = newName
	nameStore.Set(newName, userData)

	std.Emit(RegisterAliasEvent, "Alias", newName, "address", address.String())
	return nil
}

// Delete marks a user and all their aliases as deleted.
func Delete(addr std.Address) error {
	if err := validateCall(); err != nil {
		return err
	}

	data := ResolveAddress(addr)
	if data == nil {
		return ErrUserNotExist
	}

	data.deleted = true

	std.Emit(DeleteUserEvent, "address", data.addr.String())
	return nil
}

func validateCall() error {
	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrNotWhitelisted
	}

	if Pausable.IsPaused() {
		panic("paused")
	}

	return nil
}

// Validate validates username and address passed in
func validateInputs(username string, address std.Address) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	if strings.TrimSpace(username) == "" {
		return ErrEmptyUsername
	}

	if len(username) > maxUsernameLen {
		return ErrNameTooLong
	}

	// Check if the username can be decoded or looks like a valid address
	if _, _, ok := std.DecodeBech32(std.Address(username)); ok || reAddressLookalike.MatchString(username) {
		return ErrNameLikeAddress
	}

	return nil
}
