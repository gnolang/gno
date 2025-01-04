package users

import (
	"gno.land/p/demo/avl"
	"regexp"
	"std"
)

var (
	callerWhitelist = avl.NewTree() // addr > struct{}{}

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
)

const (
	SetUserEvent     = "UserSet"
	CreateAliasEvent = "AliasCreated"
	DeleteUserEvent  = "UserDeleted"

	maxUsernameLen = 64
)

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
