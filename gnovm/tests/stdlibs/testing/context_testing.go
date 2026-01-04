package testing

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/chain/banker"
	"github.com/gnolang/gno/gnovm/tests/stdlibs/chain/runtime"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func X_getContext(m *gno.Machine) (
	originCaller string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
) {
	ctx := m.Context.(*runtime.TestExecContext)

	originCaller = ctx.OriginCaller.String()

	for _, coin := range ctx.OriginSend {
		origSendDenoms = append(origSendDenoms, coin.Denom)
		origSendAmounts = append(origSendAmounts, coin.Amount)
	}

	for _, coin := range *ctx.OriginSendSpent {
		origSpendDenoms = append(origSpendDenoms, coin.Denom)
		origSpendAmounts = append(origSpendAmounts, coin.Amount)
	}

	chainID = ctx.ChainID
	height = ctx.Height
	timeUnix = ctx.Timestamp
	timeNano = ctx.TimestampNano
	return
}

func X_setContext(
	m *gno.Machine,
	originCaller string,
	currRealmAddr string, currRealmPkgPath string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
) {
	ctx := m.Context.(*runtime.TestExecContext)

	ctx.ChainID = chainID
	ctx.Height = height
	ctx.Timestamp = timeUnix
	ctx.TimestampNano = timeNano
	ctx.OriginCaller = crypto.Bech32Address(originCaller)

	if currRealmAddr != "" {
		// Associate the given Realm with the caller's frame.
		var frameIdx int
		// NOTE: the frames are different from when calling testing.SetRealm (has been refactored to this code)
		//
		// When calling this function from Gno, the 3 top frames are the following:
		// #7: [FRAME FUNC:setContext RECV:(undefined) (15 args) 11/3/0/6/4 LASTPKG:testing ...]
		// #6: [FRAME FUNC:SetContext RECV:(undefined) (1 args) 8/2/0/4/3 LASTPKG:testing ...]
		// #5: [FRAME FUNC:SetRealm RECV:(undefined) (1 args) 5/1/0/2/2 LASTPKG:gno.land/r/demo/groups ...]
		// We want to set the Realm of the frame where testing.SetRealm is being called, hence -3-1.
		for i := m.NumFrames() - 4; i >= 0; i-- {
			// Must be a frame from calling a function.
			if fr := m.Frames[i]; fr.Func != nil && fr.Func.PkgPath != "testing" {
				frameIdx = i
				break
			}
		}

		m.Frames[frameIdx].TestOverridden = true // in case frame gets popped
		ctx.RealmFrames[frameIdx] = runtime.RealmOverride{
			Addr:    crypto.Bech32Address(currRealmAddr),
			PkgPath: currRealmPkgPath,
		}
	}

	ctx.OriginSend = banker.CompactCoins(origSendDenoms, origSendAmounts)
	coins := banker.CompactCoins(origSpendDenoms, origSpendAmounts)
	ctx.OriginSendSpent = &coins

	m.Context = ctx
}

func X_testIssueCoins(m *gno.Machine, addr string, denom []string, amt []int64) {
	ctx := m.Context.(*runtime.TestExecContext)
	banker := ctx.Banker
	for i := range denom {
		banker.IssueCoin(crypto.Bech32Address(addr), denom[i], amt[i])
	}
}

func X_newRealm(m *gno.Machine, addr, pkgPath string) gno.TypedValue {
	return gno.TypedValue{
		// testing imports chain/runtime, so this type is always available.
		T: m.Store.GetType("chain/runtime.Realm"),
		V: m.Alloc.NewStructWithFields(
			// addr address
			gno.TypedValue{T: m.Store.GetType(".uverse.address"), V: gno.StringValue(addr)},
			// pkgPath string
			gno.TypedValue{T: gno.StringType, V: gno.StringValue(pkgPath)},
		),
	}
}

func X_isRealm(m *gno.Machine, pkgPath string) bool {
	return gno.IsRealmPath(pkgPath)
}
