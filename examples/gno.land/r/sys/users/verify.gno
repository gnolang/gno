package users

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/r/demo/users"
	"gno.land/r/sys/teams"
)

const adminAddress = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul

type (
	VerifyNameFunc    func(enabled bool, address std.Address, name string) bool
	VerifyPackageFunc func(enabled bool, address std.Address, name, path string) bool
)

var (
	owner          = ownable.NewWithAddress(adminAddress)
	nameCheck      = VerifyNameByUser
	packageCheck   = VerifyPackageAuthorization
	packageEnabled = false
)

// Core authorization functions
func VerifyNameByUser(enable bool, address std.Address, name string) bool {
	if checkBasicAuth(enable, address, name) {
		return true
	}

	if user := users.GetUserByName(name); user != nil {
		return user.Address == address
	}

	return false
}

func VerifyPackageAuthorization(enable bool, address std.Address, name, path string) bool {
	if checkBasicAuth(enable, address, name) {
		return true
	}

	user := users.GetUserByName(name)
	if user == nil {
		return false // user doesn't exist
	}

	// Check if user is a team
	if team := teams.Get(user.Address); team != nil {
		return team.Can(address, teams.ActionAddPackage{Path: path})
	}

	// Else user need to own the name
	return user.Address == address
}

// Public authorization interfaces
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return nameCheck(packageEnabled, address, name)
}

func IsAuthorizedToAddPackage(address std.Address, name, path string) bool {
	return packageCheck(packageEnabled, address, name, path)
}

// System status management
func IsEnabled() bool { return packageEnabled }

// Helper functions
func checkBasicAuth(enable bool, address std.Address, name string) (authorized bool) {
	if !enable {
		return true
	}

	return address.String() == name
}

// Administration functions
func AdminEnable() {
	if !owner.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	packageEnabled = true
}

func AdminDisable() {
	if !owner.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	packageEnabled = false
}

func AdminUpdateNameCheck(fn VerifyNameFunc) {
	if !owner.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	nameCheck = fn
}

func AdminUpdatePackageCheck(fn VerifyPackageFunc) {
	if !owner.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	packageCheck = fn
}

func AdminTransferOwnership(newOwner std.Address) {
	if !owner.CallerIsOwner() {
		panic(ownable.ErrUnauthorized)
	}

	owner.TransferOwnership(newOwner)
}
