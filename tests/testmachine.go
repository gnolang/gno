package tests

import (
	"io"

	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/stdlibs"
)

func TestMachine(store gno.Store, stdout io.Writer, pkgPath string) *gno.Machine {
	// default values
	var (
		send     std.Coins
		maxAlloc int64
	)

	return testMachineCustom(store, pkgPath, stdout, maxAlloc, send)
}

func testMachineCustom(store gno.Store, pkgPath string, stdout io.Writer, maxAlloc int64, send std.Coins) *gno.Machine {
	// FIXME (Manfred): create a better package to manage this, with custom constructors
	pkgAddr := gno.DerivePkgAddr(pkgPath)                      // the addr of the pkgPath called.
	caller := gno.DerivePkgAddr(pkgPath)                       // NOTE: for the purpose of testing, the caller is generally the "main" package, same as pkgAddr.
	pkgCoins := std.MustParseCoins("200000000ugnot").Add(send) // >= send.
	banker := newTestBanker(pkgAddr.Bech32(), pkgCoins)
	ctx := stdlibs.ExecContext{
		ChainID:       "dev",
		Height:        123,
		Timestamp:     1234567890,
		Msg:           nil,
		OrigCaller:    caller.Bech32(),
		OrigPkgAddr:   pkgAddr.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		Banker:        banker,
	}
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       "", // set later.
		Output:        stdout,
		Store:         store,
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
	})
	return m
}
