package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

const admin = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul

var (
	owner = ownable.NewWithAddress(admin) // Package owner
)

// Admin calls

// AdminEnable enables this package.
func AdminEnable() {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	enabled = true
}

// AdminDisable disables this package.
func AdminDisable() {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	enabled = false
}

// AdminUpdateVerifyCall updates the method that verifies the namespace.
func AdminUpdateVerifyCall(check VerifyNameFunc) {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	checkFunc = check
}

// AdminTransferOwnership transfers the ownership to a new owner.
func AdminTransferOwnership(newOwner std.Address) error {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	return owner.TransferOwnership(newOwner)
}
