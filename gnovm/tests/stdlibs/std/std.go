package std

import (
	"fmt"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tm2std "github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !IsOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}

func IsOriginCall(m *gno.Machine) bool {
	tname := m.Frames[0].Func.Name
	switch tname {
	case "main": // test is a _filetest
		return len(m.Frames) == 3
	case "runtest": // test is a _test
		return len(m.Frames) == 7
	}
	// support init() in _filetest
	// XXX do we need to distinguish from 'runtest'/_test?
	// XXX pretty hacky even if not.
	if strings.HasPrefix(string(tname), "init.") {
		return len(m.Frames) == 3
	}
	panic("unable to determine if test is a _test or a _filetest")
}

func TestCurrentRealm(m *gno.Machine) string {
	return m.Realm.Path
}

func TestSkipHeights(m *gno.Machine, count int64) {
	ctx := m.Context.(std.ExecContext)
	ctx.Height += count
	m.Context = ctx
}

func ClearStoreCache(m *gno.Machine) {
	if gno.IsDebug() && testing.Verbose() {
		m.Store.Print()
		fmt.Println("========================================")
		fmt.Println("CLEAR CACHE (RUNTIME)")
		fmt.Println("========================================")
	}
	m.Store.ClearCache()
	m.PreprocessAllFilesAndSaveBlockNodes()
	if gno.IsDebug() && testing.Verbose() {
		m.Store.Print()
		fmt.Println("========================================")
		fmt.Println("CLEAR CACHE DONE")
		fmt.Println("========================================")
	}
}

func GetCallerAt(m *gno.Machine, n int) crypto.Bech32Address {
	if n <= 0 {
		m.Panic(typedString("GetCallerAt requires positive arg"))
		return ""
	}
	if n > m.NumFrames()-1 {
		// NOTE: the last frame's LastPackage
		// is set to the original non-frame
		// package, so need this check.
		m.Panic(typedString("frame not found"))
		return ""
	}
	if n == m.NumFrames()-1 {
		// This makes it consistent with GetOrigCaller and TestSetOrigCaller.
		ctx := m.Context.(stdlibs.ExecContext)
		return ctx.OrigCaller
	}
	return m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32()
}

func TestSetOrigCaller(m *gno.Machine, addr crypto.Bech32Address) {
	ctx := m.Context.(std.ExecContext)
	ctx.OrigCaller = crypto.Bech32Address(addr)
	m.Context = ctx
}

func TestSetOrigPkgAddr(m *gno.Machine, addr crypto.Bech32Address) {
	ctx := m.Context.(stdlibs.ExecContext)
	ctx.OrigPkgAddr = addr
	m.Context = ctx
}

func TestSetOrigSend(m *gno.Machine, sent, spent tm2std.Coins) {
	ctx := m.Context.(stdlibs.ExecContext)
	ctx.OrigSend = sent
	ctx.OrigSendSpent = &spent
	m.Context = ctx
}

func TestIssueCoins(m *gno.Machine, addr crypto.Bech32Address, coins tm2std.Coins) {
	ctx := m.Context.(stdlibs.ExecContext)
	banker := ctx.Banker
	for _, coin := range coins {
		banker.IssueCoin(addr, coin.Denom, coin.Amount)
	}
}
