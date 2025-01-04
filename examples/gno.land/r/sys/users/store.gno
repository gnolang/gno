package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/pausable"
)

var (
	callerWhitelist = avl.NewTree() // address > struct{}{}

	nameStore    = avl.NewTree() // name/aliases > *UserData
	addressStore = avl.NewTree() // address > *UserData - always points to the latest alias

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
)

const (
	SetUserEvent     = "UserSet"
	CreateAliasEvent = "AliasCreated"
	DeleteUserEvent  = "UserDeleted"

	maxUsernameLen = 64
)

type UserData struct {
	addr     std.Address
	username string
	alias    *UserData
	deleted  bool
}

func (u UserData) IsDeleted() bool {
	// No need to resolve, any UserData this is called upon will already be resolved
	return u.deleted
}

const maxNameLength = 64

// Register adds a new user to the system.
func Register(name string, address std.Address) error {
	// Validate inputs
	if err := validateInputs(name, address); err != nil {
		return err
	}

	if err := validateCall(std.PrevRealm().Addr()); err != nil {
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
		alias:    nil,
		deleted:  false,
	}

	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address.String(), data)

	return nil
}

// RegisterAlias adds an alias pointing to an existing user.
func RegisterAlias(alias string, address std.Address) error {
	// Validate inputs
	if err := validateInputs(alias, address); err != nil {
		return err
	}

	// Check if requested alias is already taken
	if nameStore.Has(alias) {
		return ErrNameTaken
	}

	// Check if user has a name before an alias
	raw, ok := addressStore.Get(address.String())
	if !ok {
		return ErrAliasBeforeName
	}

	if raw.(*UserData).IsDeleted() {
		return ErrDeletedUser
	}

	// Create new Alias
	newData := &UserData{
		addr:     address,
		username: alias,
		alias:    nil,
		deleted:  false,
	}

	oldData := raw.(*UserData)
	// Point to new alias
	oldData.alias = newData
	oldData.addr = ""

	// Set latest alias to stores
	addressStore.Set(address.String(), newData)
	nameStore.Set(alias, newData)

	return nil
}

// Delete marks a user and all their aliases as deleted.
func Delete(name string) error {
	if err := validateCall(std.PrevRealm().Addr()); err != nil {
		return err
	}

	data := ResolveName(name)
	if data == nil {
		return ErrUserNotExist
	}

	if data.deleted {
		return ErrAlreadyDeleted
	}

	data.deleted = true
	// data.addr = "" // I don't think we should do this, it will likely cause user errors
	// data.username = "ghost" // could cause issues with users named "ghost"

	return nil
}

func validateCall(caller std.Address) error {
	if !IsOnWhitelist(caller) {
		return ErrNotWhitelisted
	}

	if Pausable.IsPaused() {
		return pausable.ErrPaused
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
