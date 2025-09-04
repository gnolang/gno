// this package demonstrates a way to manage contract releases.
package releases_example

import (
	"chain"
	"chain/runtime"

	"gno.land/p/demo/releases"
)

var (
	changelog = releases.NewChangelog("example_app")
	admin     = chain.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @administrator
)

func init() {
	// FIXME: admin = std.GetCreator()
}

func NewRelease(_ realm, name, url, notes string) {
	caller := runtime.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	changelog.NewRelease(name, url, notes)
}

func UpdateAdmin(_ realm, address_XXX chain.Address) {
	caller := runtime.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	admin = address_XXX
}

func Render(path string) string {
	return changelog.Render(path)
}
