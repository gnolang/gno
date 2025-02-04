package testing

import (
	"errors"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func X_unixNano() int64 {
	// only implemented in testing stdlibs
	return 0
}

func X_testSetContext(
	m *gno.Machine,
	isOrigin bool,
	originCaller string,
	originPkgAddress string,
	currRealmAddr string, currRealmPkgPath string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
) {
	ctx := m.Context.(*teststd.TestExecContext)

	if chainID != "" {
		ctx.ChainID = chainID
	}

	if height != 0 {
		ctx.Height = height
	}

	if timeUnix != 0 {
		ctx.Timestamp = timeUnix
	}

	if timeNano != 0 {
		ctx.TimestampNano = timeNano
	}

	if originCaller != "" {
		ctx.OrigCaller = crypto.Bech32Address(originCaller)
	}

	if originPkgAddress != "" {
		ctx.OrigPkgAddr = crypto.Bech32Address(originPkgAddress)
	}

	if currRealmAddr != "" {
		// Associate the given Realm with the caller's frame.
		var frame *gno.Frame
		// NOTE: the frames are different from when calling std.TestSetRealm (has been refactored to this code)
		//
		// When calling this function from Gno, the 3 top frames are the following:
		// #7: [FRAME FUNC:testSetContext RECV:(undefined) (15 args) 11/3/0/6/4 LASTPKG:testing ...]
		// #6: [FRAME FUNC:TestSetContext RECV:(undefined) (1 args) 8/2/0/4/3 LASTPKG:testing ...]
		// #5: [FRAME FUNC:SetRealm RECV:(undefined) (1 args) 5/1/0/2/2 LASTPKG:gno.land/r/demo/groups ...]
		// We want to set the Realm of the frame where t/testing.SetRealm is being called, hence -4.
		for i := m.NumFrames() - 4; i >= 0; i-- {
			// Must be a frame from calling a function.
			if fr := m.Frames[i]; fr.Func != nil {
				frame = fr
				break
			}
		}

		ctx.RealmFrames[frame] = teststd.RealmOverride{
			Addr:    crypto.Bech32Address(currRealmAddr),
			PkgPath: currRealmPkgPath,
		}
	}

	if len(origSendDenoms) > 0 && len(origSendDenoms) == len(origSendAmounts) {
		ctx.OrigSend = std.CompactCoins(origSendDenoms, origSendAmounts)
	}

	if len(origSpendDenoms) > 0 && len(origSpendDenoms) == len(origSpendAmounts) {
		coins := std.CompactCoins(origSpendDenoms, origSpendAmounts)
		ctx.OrigSendSpent = &coins
	}

	m.Context = ctx
}

func X_testIssueCoins(m *gno.Machine, addr string, denom []string, amt []int64) {
	ctx := m.Context.(*teststd.TestExecContext)
	banker := ctx.Banker
	for i := range denom {
		banker.IssueCoin(crypto.Bech32Address(addr), denom[i], amt[i])
	}
}

func X_matchString(pat, str string) (result bool, err error) {
	return false, errors.New("only implemented in testing stdlibs")
}
