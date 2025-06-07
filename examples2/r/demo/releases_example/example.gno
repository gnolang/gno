// this package demonstrates a way to manage contract releases.
package releases_example

import (
	"std"

	"gno.land/p/demo/releases"
)

var (
	changelog = releases.NewChangelog("example_app")
	admin     = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @administrator
)

func init() {
	// FIXME: admin = std.GetCreator()
}

func NewRelease(name, url, notes string) {
	caller := std.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	changelog.NewRelease(name, url, notes)
}

func UpdateAdmin(address_XXX std.Address) {
	caller := std.OriginCaller()
	if caller != admin {
		panic("restricted area")
	}
	admin = address_XXX
}

func Render(path string) string {
	return changelog.Render(path)
}
