package std

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tm2std "github.com/gnolang/gno/tm2/pkg/std"
)

// TestExecContext is the testing extension of the exec context.
type TestExecContext struct {
	std.ExecContext

	// These are used to set up the result of CurrentRealm() and PrevRealm().
	RealmFrames map[*gno.Frame]RealmOverride
}

var _ std.ExecContexter = &TestExecContext{}

type RealmOverride struct {
	Addr    crypto.Bech32Address
	PkgPath string
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}

func TestSkipHeights(m *gno.Machine, count int64) {
	ctx := m.Context.(*TestExecContext)
	ctx.Height += count
	ctx.Timestamp += (count * 5)
	m.Context = ctx
}

func X_callerAt(m *gno.Machine, n int) string {
	if n <= 0 {
		m.Panic(typedString("GetCallerAt requires positive arg"))
		return ""
	}
	// Add 1 to n to account for the GetCallerAt (gno fn) frame.
	n++
	if n > m.NumFrames()-1 {
		// NOTE: the last frame's LastPackage
		// is set to the original non-frame
		// package, so need this check.
		m.Panic(typedString("frame not found"))
		return ""
	}
	if n == m.NumFrames()-1 {
		// This makes it consistent with GetOrigCaller and TestSetOrigCaller.
		ctx := m.Context.(*TestExecContext)
		return string(ctx.OrigCaller)
	}
	return string(m.MustLastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
}

func X_testSetOrigCaller(m *gno.Machine, addr string) {
	ctx := m.Context.(*TestExecContext)
	ctx.OrigCaller = crypto.Bech32Address(addr)
	m.Context = ctx
}

func X_testSetOrigPkgAddr(m *gno.Machine, addr string) {
	ctx := m.Context.(*TestExecContext)
	ctx.OrigPkgAddr = crypto.Bech32Address(addr)
	m.Context = ctx
}

func X_testSetRealm(m *gno.Machine, addr, pkgPath string) {
	// Associate the given Realm with the caller's frame.
	var frame *gno.Frame
	// When calling this function from Gno, the two top frames are the following:
	// #6 [FRAME FUNC:testSetRealm RECV:(undefined) (2 args) 17/6/0/10/8 LASTPKG:std ...]
	// #5 [FRAME FUNC:TestSetRealm RECV:(undefined) (1 args) 14/5/0/8/7 LASTPKG:gno.land/r/tyZ1Vcsta ...]
	// We want to set the Realm of the frame where TestSetRealm is being called, hence -3.
	for i := m.NumFrames() - 3; i >= 0; i-- {
		// Must be a frame from calling a function.
		if fr := m.Frames[i]; fr.Func != nil {
			frame = fr
			break
		}
	}

	ctx := m.Context.(*TestExecContext)
	ctx.RealmFrames[frame] = RealmOverride{
		Addr:    crypto.Bech32Address(addr),
		PkgPath: pkgPath,
	}
}

func X_getRealm(m *gno.Machine, height int) (address string, pkgPath string) {
	// NOTE: keep in sync with stdlibs/std.getRealm

	var (
		ctx           = m.Context.(*TestExecContext)
		currentCaller crypto.Bech32Address
		// Keeps track of the number of times currentCaller
		// has changed.
		changes int
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := m.Frames[i]
		override, overridden := ctx.RealmFrames[m.Frames[max(i-1, 0)]]
		if !overridden &&
			(fr.LastPackage == nil || !fr.LastPackage.IsRealm()) {
			continue
		}

		// LastPackage is a realm. Get caller and pkgPath, and compare against
		// currentCaller.
		caller, pkgPath := override.Addr, override.PkgPath
		if !overridden {
			caller = fr.LastPackage.GetPkgAddr().Bech32()
			pkgPath = fr.LastPackage.PkgPath
		}
		if caller != currentCaller {
			if changes == height {
				return string(caller), pkgPath
			}
			currentCaller = caller
			changes++
		}
	}

	// Fallback case: return OrigCaller.
	return string(ctx.OrigCaller), ""
}

func X_isRealm(m *gno.Machine, pkgPath string) bool {
	return gno.IsRealmPath(pkgPath)
}

func X_testSetOrigSend(m *gno.Machine,
	sentDenom []string, sentAmt []int64,
	spentDenom []string, spentAmt []int64,
) {
	ctx := m.Context.(*TestExecContext)
	ctx.OrigSend = std.CompactCoins(sentDenom, sentAmt)
	spent := std.CompactCoins(spentDenom, spentAmt)
	ctx.OrigSendSpent = &spent
	m.Context = ctx
}

// TestBanker is a banker that can be used as a mock banker in test contexts.
type TestBanker struct {
	CoinTable map[crypto.Bech32Address]tm2std.Coins
}

var _ std.BankerInterface = &TestBanker{}

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
	tcoins, _ := tb.CoinTable[to]
	tsum := tcoins.Add(amt)
	tb.CoinTable[to] = tsum
}

// TotalCoin implements the Banker interface.
func (tb *TestBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

// IssueCoin implements the Banker interface.
func (tb *TestBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.CoinTable[addr]
	sum := coins.Add(tm2std.Coins{{Denom: denom, Amount: amt}})
	tb.CoinTable[addr] = sum
}

// RemoveCoin implements the Banker interface.
func (tb *TestBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.CoinTable[addr]
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
