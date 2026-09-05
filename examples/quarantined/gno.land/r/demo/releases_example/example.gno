// this package demonstrates a way to manage contract releases.
package releases_example

import (
	"chain/runtime/unsafe"

	"gno.land/p/demo/releases"
)

var (
	changelog = releases.NewChangelog("example_app")
	admin     = address("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf") // govdao t1 multisig
)

func init() {
	// FIXME: admin = std.GetCreator()
}

func NewRelease(_ realm, name, url, notes string) {
	caller := unsafe.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	changelog.NewRelease(name, url, notes)
}

func UpdateAdmin(_ realm, address_XXX address) {
	caller := unsafe.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	admin = address_XXX
}

func Render(path string) string {
	return changelog.Render(path)
}
