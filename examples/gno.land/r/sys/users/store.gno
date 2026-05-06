package users

import (
	"chain"
	"chain/runtime"
	"regexp"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	nameStore    = bptree.NewBPTree32() // name/aliases > *UserData
	addressStore = bptree.NewBPTree32() // address > *UserData

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)

	// reName mirrors gno's package-name shape (gnovm/pkg/gnolang/mempackage.go
	// `Re_name`): start with a lowercase letter, optional alphanumeric body,
	// then any number of (separator + alphanumeric run) — so single hyphens
	// or underscores are allowed BETWEEN alphanumerics, but consecutive
	// separators (`--`, `__`, `-_`, `_-`) are rejected, and so are leading
	// or trailing separators. Lowercase-only — closes the case-confusable
	// squatting concern (Alice vs alice were two distinct names under the
	// previous case-preserving regex). Length cap of 64 enforced separately
	// in validateName.
	reName = regexp.MustCompile(`^[a-z][a-z0-9]*([_-][a-z0-9]+)*$`)
)

const maxNameLen = 64

const (
	RegisterUserEvent = "Registered"
	UpdateNameEvent   = "Updated"
	DeleteUserEvent   = "Deleted"
)

type UserData struct {
	addr     address
	username string // contains the latest name of a user
	deleted  bool
}

func (u UserData) Name() string {
	return u.username
}

func (u UserData) Addr() address {
	return u.addr
}

// IsDeleted reports whether this user record is missing or marked deleted.
// A nil receiver returns true — "the user does not exist" is semantically
// indistinguishable from "the user was deleted" for callers that need to
// gate further state changes. This lets call sites collapse the nil check
// and the deleted check into a single guard:
//
//	if u.IsDeleted() {
//	    return ErrUserNotExistOrDeleted
//	}
func (u *UserData) IsDeleted() bool {
	if u == nil {
		return true
	}
	return u.deleted
}

// RenderLink provides a render link to the user page on gnoweb
// `linkText` is optional
func (u UserData) RenderLink(linkText string) string {
	if linkText == "" {
		return ufmt.Sprintf("[@%s](/u/%s)", u.username, u.username)
	}

	return ufmt.Sprintf("[%s](/u/%s)", linkText, u.username)
}

// registerUser adds a new user to the system without checking controllers.
// The ignoreCanonical flag suppresses ErrCanonicalCollision; the canonical
// store is written either way (decision #14: later-wins on bypass).
func registerUser(cur realm, name string, address_XXX address, ignoreCanonical bool) error {
	// Validate name
	if err := validateName(name); err != nil {
		return err
	}

	// Validate address
	if !address_XXX.IsValid() {
		return ErrInvalidAddress
	}

	// Check if name is taken (exact-string match precedes canonical check)
	if nameStore.Has(name) {
		return ErrNameTaken
	}

	canonical := Canonicalize(name)
	if !ignoreCanonical {
		if _, taken := canonicalStore.Get(canonical); taken {
			return ErrCanonicalCollision
		}
	}

	raw, ok := addressStore.Get(address_XXX.String())
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
		addr:     address_XXX,
		username: name,
		deleted:  false,
	}

	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address_XXX.String(), data)
	canonicalStore.Set(canonical, name)

	chain.Emit(RegisterUserEvent,
		"name", name,
		"address", address_XXX.String(),
	)
	return nil
}

// RegisterUser adds a new user to the system. Enforces canonical-
// collision detection: a name whose Canonicalize-form matches a prior
// registration returns ErrCanonicalCollision.
func RegisterUser(cur realm, name string, address_XXX address) error {
	// At genesis (height 0), allow any caller to register users.
	// After genesis, only whitelisted controllers can register.
	if runtime.ChainHeight() > 0 && !controllers.Has(runtime.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	return registerUser(cur, name, address_XXX, false)
}

// RegisterUserIgnoreCanonical is the bypass path: same controller-
// whitelist gate, but ErrCanonicalCollision is suppressed. The canonical
// store is still written; a prior entry with the same canonical key is
// silently overwritten (decision #14, later-wins). Use sparingly — names
// registered here can canonical-collide with existing ones, weakening
// confusable protection for everyone.
func RegisterUserIgnoreCanonical(cur realm, name string, address_XXX address) error {
	if runtime.ChainHeight() > 0 && !controllers.Has(runtime.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	return registerUser(cur, name, address_XXX, true)
}

// updateName adds a name that is associated with a specific address without
// checking controllers. The ignoreCanonical flag suppresses
// ErrCanonicalCollision; the canonical store is written either way (decision
// #14: later-wins on bypass).
//
// All previous names are preserved and resolvable.
// The new name is the default value returned for address lookups.
func (u *UserData) updateName(newName string, ignoreCanonical bool) error {
	// IsDeleted handles both branches: nil receiver (user never existed)
	// AND a non-nil receiver whose .deleted is true (a controller cached
	// the *UserData pointer before the user was deleted by a separate
	// controller or governance proposal). Without the deleted-flag branch,
	// nameStore.Set(newName, u) would insert an alias pointing at a
	// deleted user — Has(newName) returns true forever but Resolve(newName)
	// returns nil (Resolve* APIs filter deleted), so the name is squatted
	// with no recovery path. (audit finding #3)
	if u.IsDeleted() {
		return ErrUserNotExistOrDeleted
	}

	// Validate name
	if err := validateName(newName); err != nil {
		return err
	}

	// Check if the requested Alias is already taken (exact-string match)
	if nameStore.Has(newName) {
		return ErrNameTaken
	}

	canonical := Canonicalize(newName)
	if !ignoreCanonical {
		// No self-collision filter (decision #15): even the user's OWN
		// prior canonical claim blocks the rename. Prevents accumulating
		// confusable aliases of one's own name through free renames. The
		// only path to a self-confusable rename is DAO governance via
		// ProposeUpdateName.
		if _, taken := canonicalStore.Get(canonical); taken {
			return ErrCanonicalCollision
		}
	}

	u.username = newName
	nameStore.Set(newName, u)
	canonicalStore.Set(canonical, newName)

	chain.Emit(UpdateNameEvent,
		"alias", newName,
		"address", u.addr.String(),
	)
	return nil
}

// UpdateName adds a name that is associated with a specific address.
// Enforces canonical-collision detection.
// All previous names are preserved and resolvable.
// The new name is the default value returned for address lookups.
func (u *UserData) UpdateName(newName string) error {
	if u.IsDeleted() {
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(runtime.CurrentRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	return u.updateName(newName, false)
}

// UpdateNameIgnoreCanonical is the bypass path: same controller-
// whitelist gate, but ErrCanonicalCollision is suppressed. The canonical
// store is still written; a prior entry with the same canonical key is
// silently overwritten (decision #14, later-wins).
func (u *UserData) UpdateNameIgnoreCanonical(newName string) error {
	if u.IsDeleted() {
		return ErrUserNotExistOrDeleted
	}

	if !controllers.Has(runtime.CurrentRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	return u.updateName(newName, true)
}

// delete marks a user and all their aliases as deleted without checking controllers.
func (u *UserData) delete() error {
	if u.IsDeleted() {
		return ErrUserNotExistOrDeleted
	}

	u.deleted = true

	chain.Emit(DeleteUserEvent, "address", u.addr.String())
	return nil
}

// Delete marks a user and all their aliases as deleted.
func (u *UserData) Delete() error {
	if u.IsDeleted() {
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(runtime.CurrentRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	return u.delete()
}

// Validate validates username and address passed in
// Most of the validation is done in the controllers
// This provides more flexibility down the line
func validateName(username string) error {
	if username == "" {
		return ErrEmptyUsername
	}

	if len(username) > maxNameLen {
		return ErrInvalidUsername
	}

	if !reName.MatchString(username) {
		return ErrInvalidUsername
	}

	// Check if the username can be decoded or looks like a valid address
	if address(username).IsValid() || reAddressLookalike.MatchString(username) {
		return ErrNameLikeAddress
	}

	return nil
}
