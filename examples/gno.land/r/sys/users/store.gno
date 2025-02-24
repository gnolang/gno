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
	RegisterUserEvent = "Registered"
	UpdateNameEvent   = "Updated"
	DeleteUserEvent   = "Deleted"

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

// RenderLink provides a render link to the user page on gnoweb
// ufmt.Sprintf("[%s](%s)", userData.Name(), userData.RenderLink())
func (u UserData) RenderLink() string {
	return "/u/" + u.username
}

// RegisterUser adds a new user to the system.
func RegisterUser(name string, address std.Address) error {
	// Validate caller
	if !controllers.Has(std.PreviousRealm().Address()) {
		return ErrNotWhitelisted
	}

	// Validate name
	if err := validateName(name); err != nil {
		return err
	}

	// Validate address
	if !address.IsValid() {
		return ErrInvalidAddress
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

		// For a second name, use UpdateName
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

	std.Emit(RegisterUserEvent,
		"name", name,
		"address", address.String(),
	)
	return nil
}

// UpdateName adds a name that is associated with a specific address
// All previous names are preserved and resolvable.
// The new name is the default value returned for address lookups.
func (u *UserData) UpdateName(newName string) error {
	if u == nil { // either doesnt exists or was deleted
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(std.PreviousRealm().Address()) {
		return ErrNotWhitelisted
	}

	// Validate name
	if err := validateName(newName); err != nil {
		return err
	}

	// Check if the requested Alias is already taken
	if nameStore.Has(newName) {
		return ErrNameTaken
	}

	u.username = newName
	nameStore.Set(newName, u)

	std.Emit(UpdateNameEvent, "alias", newName, "address", u.addr.String())
	return nil
}

// Delete marks a user and all their aliases as deleted.
func (u *UserData) Delete() error {
	if u == nil {
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(std.PreviousRealm().Address()) {
		return ErrNotWhitelisted
	}

	u.deleted = true

	std.Emit(DeleteUserEvent, "address", u.addr.String())
	return nil
}

// Validate validates username and address passed in
func validateName(username string) error {
	if strings.TrimSpace(username) == "" {
		return ErrEmptyUsername
	}

	if len(username) > maxUsernameLen {
		return ErrNameTooLong
	}

	// Check if the username can be decoded or looks like a valid address
	if std.Address(username).IsValid() || reAddressLookalike.MatchString(username) {
		return ErrNameLikeAddress
	}

	return nil
}
