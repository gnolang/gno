// Package test contains the code to parse and execute Gno tests and filetests.
package test

import (
	"io"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	// DefaultHeight is the default height used in the [Context].
	DefaultHeight = 123
	// DefaultTimestamp is the Timestamp value used by default in [Context]. It is
	// Friday, 29 January 2021 17:00:00 UTC - approximation of gnolang/gno's initial commit
	DefaultTimestamp = 1611939600
	// DefaultCaller is the result of gno.DerivePkgAddr("user1.gno")
	DefaultCaller crypto.Bech32Address = "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
)

// Context returns a TestExecContext. Usable for test purpose only.
// The returned context has a mock banker, params and event logger. It will give
// the pkgAddr the coins in `send` by default, and only that.
// The Height and Timestamp parameters are set to the [DefaultHeight] and
// [DefaultTimestamp].
func Context(pkgPath string, send std.Coins) *teststd.TestExecContext {
	// FIXME: create a better package to manage this, with custom constructors
	pkgAddr := gno.DerivePkgAddr(pkgPath) // the addr of the pkgPath called.

	banker := &teststd.TestBanker{
		CoinTable: map[crypto.Bech32Address]std.Coins{
			pkgAddr.Bech32(): send,
		},
	}
	ctx := stdlibs.ExecContext{
		ChainID:       "dev",
		Height:        DefaultHeight,
		Timestamp:     DefaultTimestamp,
		OrigCaller:    DefaultCaller,
		OrigPkgAddr:   pkgAddr.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		Banker:        banker,
		Params:        newTestParams(),
		EventLogger:   sdk.NewEventLogger(),
	}
	return &teststd.TestExecContext{
		ExecContext: ctx,
		RealmFrames: make(map[*gno.Frame]teststd.RealmOverride),
	}
}

// Machine is a minimal machine, set up with just the Store, Output and Context.
func Machine(testStore gno.Store, output io.Writer, pkgPath string) *gno.Machine {
	return gno.NewMachineWithOptions(gno.MachineOptions{
		Store:   testStore,
		Output:  output,
		Context: Context(pkgPath, nil),
	})
}

// ----------------------------------------
// testParams

type testParams struct{}

func newTestParams() *testParams {
	return &testParams{}
}

func (tp *testParams) SetBool(key string, val bool)     { /* noop */ }
func (tp *testParams) SetBytes(key string, val []byte)  { /* noop */ }
func (tp *testParams) SetInt64(key string, val int64)   { /* noop */ }
func (tp *testParams) SetUint64(key string, val uint64) { /* noop */ }
func (tp *testParams) SetString(key string, val string) { /* noop */ }
