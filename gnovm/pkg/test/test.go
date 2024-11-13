// Package test contains the code to parse and execute Gno tests and filetests.
package test

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	DefaultHeight = 123
	// Friday, 29 January 2021 17:00:00 UTC - approximation of gnolang/gno's initial commit
	DefaultTimestamp = 1611939600
)

// TestContext returns a TestExecContext. Usable for test purpose only.
func TestContext(pkgPath string, send std.Coins) *teststd.TestExecContext {
	// FIXME: create a better package to manage this, with custom constructors
	pkgAddr := gno.DerivePkgAddr(pkgPath) // the addr of the pkgPath called.
	caller := gno.DerivePkgAddr("user1.gno")

	banker := &testBanker{
		coinTable: map[crypto.Bech32Address]std.Coins{
			pkgAddr.Bech32(): send,
		},
	}
	ctx := stdlibs.ExecContext{
		ChainID:       "dev",
		Height:        DefaultHeight,
		Timestamp:     DefaultTimestamp,
		OrigCaller:    caller.Bech32(),
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

// ----------------------------------------
// testBanker

type testBanker struct {
	coinTable map[crypto.Bech32Address]std.Coins
}

func (tb *testBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return tb.coinTable[addr]
}

func (tb *testBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
	fcoins, fexists := tb.coinTable[from]
	if !fexists {
		panic(fmt.Sprintf(
			"source address %s does not exist",
			from.String()))
	}
	if !fcoins.IsAllGTE(amt) {
		panic(fmt.Sprintf(
			"source address %s has %s; cannot send %s",
			from.String(), fcoins, amt))
	}
	// First, subtract from 'from'.
	frest := fcoins.Sub(amt)
	tb.coinTable[from] = frest
	// Second, add to 'to'.
	// NOTE: even works when from==to, due to 2-step isolation.
	tcoins, _ := tb.coinTable[to]
	tsum := tcoins.Add(amt)
	tb.coinTable[to] = tsum
}

func (tb *testBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (tb *testBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	sum := coins.Add(std.Coins{{Denom: denom, Amount: amt}})
	tb.coinTable[addr] = sum
}

func (tb *testBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	rest := coins.Sub(std.Coins{{Denom: denom, Amount: amt}})
	tb.coinTable[addr] = rest
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
