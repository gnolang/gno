package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

const adminAddr = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul

var (
	admin    = ownable.NewWithAddress(adminAddr) // Package owner
	pausable = p.NewFromOwnable(admin)           //
)

// AdminUpdateVerifyCall updates the method that verifies the namespace.
func AdminUpdateVerifyCall(check VerifyNameFunc) {
	if err := admin.CallerIsOwner(); err != nil {
		panic(err)
	}

	checkFunc = check
}

func Pause() error {
	return pausable.Pause()
}

func Unpause() error {
	return pausable.Unpause()
}

func GetOwner() std.Address {
	return admin.Owner()
}

func TransferOwnership(newOwner std.Address) {
	if err := admin.TransferOwnership(newOwner); err != nil {
		panic(err)
	}
}
