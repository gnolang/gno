package config

import (
	"std"

	"gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func init() {
	// Replace these with actual gno addresses
	moul := std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	leon := std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	michelle := std.Address("g1778y2yphxs2wpuaflsy5y9qwcd4gttn4g5yjx5")

	// Initialize with two admins
	auth = authz.NewWithAuthority(authz.NewMemberAuthority(moul, michelle, leon))
}

// ListAdmins returns the list of current admin addresses.
func ListAdmins() []std.Address {
	return auth.Current().(*authz.MemberAuthority).Tree().Values()
}

// AddAdmin adds a new admin address (can only be called by an admin).
func AddAdmin(addr std.Address) error {
	if addr.
	return auth.Current().(*authz.MemberAuthority).AddMember(addr)
}

// RemoveAdmin removes an admin address (can only be called by an admin).
func RemoveAdmin(addr std.Address) error {
	return auth.Current().(*authz.MemberAuthority).RemoveMember(addr)
}

// IsAdmin returns true if the address is an admin.
func IsAdmin(addr std.Address) bool {
	return auth.Current().(*authz.MemberAuthority).Has(addr)
}
