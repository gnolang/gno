package vm

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// A run script can write a realm param, but its package is never saved, so the
// storage deposit has no realm to charge against. processStorageDeposit refuses
// the whole message rather than take either wrong turn: charging would deref a
// realm that is not there, and skipping would commit the bytes with no deposit
// behind them and no baseline to refund from later.
func TestRunCannotWriteParamsWithoutARealmToChargeFor(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("run-params-writer"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)

	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/x/run")},
		{Name: "main.gno", Body: `package main

import params "chain/params"

func main(cur realm) {
	params.SetString("k", "v")
}`},
	}
	msg := NewMsgRun(addr, std.Coins{}, files)
	// Well clear of the cost, so a deposit ceiling cannot be what refuses this.
	msg.MaxDeposit = std.MustParseCoins(ugnot.ValueString(10_000_000))

	_, err := env.vmk.Run(ctx, msg)
	require.Error(t, err, "a run script writing params must not be accepted")
	require.Contains(t, err.Error(), "unknown realm",
		"the refusal must be the missing realm, not some other fault")
}
