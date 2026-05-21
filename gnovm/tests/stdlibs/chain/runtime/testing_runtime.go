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
	// Count only actual function call frames (excludes closures and
	// control-flow basic frames like for/range/switch).
	callFrames := m.NumCallFrames()
	switch tname {
	case "main": // test is a _filetest
		// Non-closure frames expected:
		// 0. main
		// 1. $RealmFuncName
		// 2. runtime.AssertOriginCall
		return callFrames == 3
	case "RunTest", "runTest_cur": // _test, with or without (cur realm, t *testing.T)
		// Non-closure frames expected:
		// 0. testing.RunTest / runTest_cur
		// 1. tRunner / tRunner_cur
		// 2. $TestFuncName / $TestFuncName_cur
		// 3. $RealmFuncName
		// 4. runtime.AssertOriginCall
		return callFrames == 5
	}
	// support init() in _filetest
	// XXX do we need to distinguish from 'runtest'/_test?
	// XXX pretty hacky even if not.
	if strings.HasPrefix(string(tname), "init.") {
		return callFrames == 3
	}
	panic("unable to determine if test is a _test or a _filetest")
}

// realmsEqual mirrors execctx.realmsEqual: nil-safe Realm comparison.
func realmsEqual(a, b *gno.Realm) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID
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
				// v3a-friendly behavior: a user-realm override
				// reached before we've accumulated `height`
				// crossings is the EOA boundary — skip it and
				// let the switch fallthrough return
				// ctx.OriginCaller for the matching height.
				// Previously this panicked with
				// "cannot seek beyond origin caller override"
				// which blocked runtime.Caller() in v3a-style
				// tests where there is no manual cross()
				// scaffolding above the override.
				lfr = fr
				continue
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

// X_getRealmV3a is the auto-cross-aware identity walk used by the
// v3a primitives (runtime.Caller / Self / CallerN). It differs from
// X_getRealm in that a frame counts as a realm-cross when the callee
// is /r/-declared in a realm different from fr.LastRealm — mirroring
// the Layer-1 borrow rule in PushFrameCall. v2 primitives
// (CurrentRealm / PreviousRealm) keep using X_getRealm so their
// established semantics remain unchanged.
//
// NOTE: keep in sync with execctx.GetRealmV3a.
func X_getRealmV3a(m *gno.Machine, height int) (addr string, pkgPath string) {
	var (
		ctx       = m.Context.(*TestExecContext)
		lfr       = m.LastFrame()
		crosses   int
		postShift = m.Realm
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := &m.Frames[i]

		override, overridden := getOverride(m, i)
		if overridden {
			if override.PkgPath == "" && crosses < height {
				lfr = fr
				continue
			}
		}
		shifted := false
		if !overridden {
			if !fr.IsCall() {
				continue
			}
			shifted = !realmsEqual(postShift, fr.LastRealm)
			if !fr.WithCross && !shifted {
				lfr = fr
				postShift = fr.LastRealm
				continue
			}
		}

		if !overridden {
			if fr.WithCross && !fr.DidCrossing {
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
		postShift = fr.LastRealm
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
				panic("should not happen")
			} else {
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
