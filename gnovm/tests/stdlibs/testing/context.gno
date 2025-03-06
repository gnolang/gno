package testing

import (
	"time"

	"github.com/gnolang/gno/gnovm/std"
)

func getContext() (
	originCaller string,
	originPkgAddress string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
)

func setContext(
	originCaller string,
	originPkgAddress string,
	currRealmAddr string, currRealmPkgPath string,
	origSendDenoms []string, origSendAmounts []int64,
	origSpendDenoms []string, origSpendAmounts []int64,
	chainID string,
	height int64,
	timeUnix int64, timeNano int64,
)

func GetContext() Context {
	originCaller, originPkgAddress, origSendDenoms, origSendAmounts, origSpendDenoms, origSpendAmounts, chainID, height, timeUnix, timeNano := getContext()
	return Context{
		OriginCaller:     std.Address(originCaller),
		OriginPkgAddress: std.Address(originPkgAddress),
		OriginSend:       compactNative(origSendDenoms, origSendAmounts),
		OriginSpend:      compactNative(origSpendDenoms, origSpendAmounts),
		ChainID:          chainID,
		Height:           height,
		Time:             time.Unix(timeUnix, timeNano),
	}
}

func SetContext(ctx Context) {
	originSendDenom, originSendAmount := expandNative(ctx.OriginSend)
	originSpendDenom, originSpendAmount := expandNative(ctx.OriginSpend)

	var timestamp, timestampNano int64 = 0, 0

	if !ctx.Time.IsZero() {
		timestamp = ctx.Time.Unix()
		timestampNano = 0
	}

	setContext(
		ctx.OriginCaller.String(),
		ctx.OriginPkgAddress.String(),
		ctx.CurrentRealm.Address().String(), ctx.CurrentRealm.PkgPath(),
		originSendDenom, originSendAmount,
		originSpendDenom, originSpendAmount,
		ctx.ChainID,
		ctx.Height,
		timestamp, timestampNano,
	)
}
