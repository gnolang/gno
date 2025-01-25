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
	prevRealmAddr string,
	prevRealmPath string,
	originPkgAddress string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
	banker bool, // TODO
	logger bool, // TODO
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

	if prevRealmAddr != "" {
		ctx.OrigCaller = crypto.Bech32Address(prevRealmAddr)
	}

	if originPkgAddress != "" {
		ctx.OrigPkgAddr = crypto.Bech32Address(originPkgAddress)
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

func X_matchString(pat, str string) (result bool, err error) {
	return false, errors.New("only implemented in testing stdlibs")
}
