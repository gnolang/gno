package runtime

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tm2std "github.com/gnolang/gno/tm2/pkg/std"
)

// TestExecContext is the testing extension of the exec context.
type TestExecContext struct {
	stdlibs.ExecContext

	// These are used to set up the result of CurrentRealm() and PreviousRealm().
	RealmFrames map[int]RealmOverride
}

var _ stdlibs.ExecContexter = &TestExecContext{}

type RealmOverride struct {
	Addr    crypto.Bech32Address
	PkgPath string
}

func AssertOriginCall(m *gno.Machine) {
	if !isOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
		return
	}
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}

func isOriginCall(m *gno.Machine) bool {
	tname := m.Frames[0].Func.Name
	switch tname {
	case "main": // test is a _filetest
		// 0. main
		// 1. $RealmFuncName
		// 2. td.IsOriginCall
		return len(m.Frames) == 3
	case "RunTest": // test is a _test
		// 0. testing.RunTest
		// 1. tRunner
		// 2. $TestFuncName
		// 3. $RealmFuncName
		// 4. std.IsOriginCall
		return len(m.Frames) == 5
	}
	// support init() in _filetest
	// XXX do we need to distinguish from 'runtest'/_test?
	// XXX pretty hacky even if not.
	if strings.HasPrefix(string(tname), "init.") {
		return len(m.Frames) == 3
	}
	panic("unable to determine if test is a _test or a _filetest")
}

func getOverride(m *gno.Machine, i int) (RealmOverride, bool) {
	fr := &m.Frames[i]
	ctx := m.Context.(*TestExecContext)
	override, overridden := ctx.RealmFrames[i]
	if overridden && !fr.TestOverridden {
		return RealmOverride{}, false // override was replaced
	}
	return override, overridden
}

func X_getRealm(m *gno.Machine, height int) (addr string, pkgPath string) {
	// NOTE: keep in sync with stdlibs/std.getRealm

	var (
		ctx     = m.Context.(*TestExecContext)
		lfr     = m.LastFrame() // last call frame
		crosses int             // track realm crosses
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := &m.Frames[i]

		// Skip over (non-realm) non-crosses.
		override, overridden := getOverride(m, i)
		if overridden {
			if override.PkgPath == "" && crosses < height {
				m.Panic(typedString("frame not found: cannot seek beyond origin caller override"))
			}
		}
		if !overridden {
			if !fr.IsCall() {
				continue
			}
			if !fr.WithCross {
				lfr = fr
				continue
			}
		}

		// Sanity check XXX move check elsewhere
		if !overridden {
			if !fr.DidCrossing {
				panic(fmt.Sprintf(
					"cross(fn) but fn didn't call crossing(): %s.%s",
					fr.Func.PkgPath,
					fr.Func.String()))
			}
		}

		crosses++
		if crosses > height {
			if overridden {
				caller, pkgPath := override.Addr, override.PkgPath
				return string(caller), pkgPath
			} else {
				currlm := lfr.LastRealm
				caller, rlmPath := gno.DerivePkgBech32Addr(currlm.Path), currlm.Path
				return string(caller), rlmPath
			}
		}
		lfr = fr
	}

	switch m.Stage {
	case gno.StageAdd:
		switch height {
		case crosses:
			fr := m.Frames[0]
			path := fr.LastPackage.PkgPath
			return string(gno.DerivePkgBech32Addr(path)), path
		case crosses + 1:
			return string(ctx.OriginCaller), ""
		default:
			m.Panic(typedString("frame not found"))
			return "", ""
		}
	case gno.StageRun:
		switch height {
		case crosses:
			fr := m.Frames[0]
			path := fr.LastPackage.PkgPath
			if path == "" {
				// Not sure what would cause this.
				panic("should not happen")
			} else {
				// e.g. TestFoo(t *testing.Test) in *_test.gno
				// or main() in *_filetest.gno
				return string(gno.DerivePkgBech32Addr(path)), path
			}
		case crosses + 1:
			return string(ctx.OriginCaller), ""
		default:
			m.Panic(typedString("frame not found"))
			return "", ""
		}
	default:
		panic("exec kind unspecified")
	}
}

// TestBanker is a banker that can be used as a mock banker in test contexts.
type TestBanker struct {
	CoinTable map[crypto.Bech32Address]tm2std.Coins
}

var _ stdlibs.BankerInterface = &TestBanker{}

// GetCoins implements the Banker interface.
func (tb *TestBanker) GetCoins(addr crypto.Bech32Address) (dst tm2std.Coins) {
	return tb.CoinTable[addr]
}

// SendCoins implements the Banker interface.
func (tb *TestBanker) SendCoins(from, to crypto.Bech32Address, amt tm2std.Coins) {
	fcoins, fexists := tb.CoinTable[from]
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
	tb.CoinTable[from] = frest
	// Second, add to 'to'.
	// NOTE: even works when from==to, due to 2-step isolation.
	tcoins := tb.CoinTable[to]
	tsum := tcoins.Add(amt)
	tb.CoinTable[to] = tsum
}

// TotalCoin implements the Banker interface.
func (tb *TestBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

// IssueCoin implements the Banker interface.
func (tb *TestBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins := tb.CoinTable[addr]
	sum := coins.Add(tm2std.Coins{{Denom: denom, Amount: amt}})
	tb.CoinTable[addr] = sum
}

// RemoveCoin implements the Banker interface.
func (tb *TestBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins := tb.CoinTable[addr]
	rest := coins.Sub(tm2std.Coins{{Denom: denom, Amount: amt}})
	tb.CoinTable[addr] = rest
}

func X_testIssueCoins(m *gno.Machine, addr string, denom []string, amt []int64) {
	ctx := m.Context.(*TestExecContext)
	banker := ctx.Banker
	for i := range denom {
		banker.IssueCoin(crypto.Bech32Address(addr), denom[i], amt[i])
	}
}
