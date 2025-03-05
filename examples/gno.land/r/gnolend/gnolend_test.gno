package gnolend

import (
	"std"
	"testing"
)

func TestGetGnolendAddress(t *testing.T) {
	// Get the package address
	pkgPath := "gno.land/r/gnolend"
	addr := std.DerivePkgAddr(pkgPath)

	// Print the address for use in tests
	println("Gnolend package address:", addr.String())
} 