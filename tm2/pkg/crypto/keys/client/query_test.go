package client

import (
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"testing"
)

func TestHandleBalanceQuery(t *testing.T) {

	pkgPath := "gno.land/r/demo/wugnot"
	queryPath := balancesQuery + pkgPath

	addr := gnolang.DerivePkgAddr(pkgPath)

	parsedPath := handleBalanceQuery(queryPath)

	if parsedPath != balancesQuery+addr.String() {
		t.Fatalf("expected %s, got %s", balancesQuery+addr.String(), parsedPath)
	}
}
