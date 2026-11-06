package vm

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// A run that sets no MaxSend sends nothing, live or in a fresh chain's genesis.
// Replayed history, which carries its source block height, runs it uncapped,
// as it ran before the field existed; a replayed run that sets one is capped.
func TestRunMaxSendGenesisReplay(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("replay-runner"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins("1000ugnot"))
	to := crypto.AddressFromPreimage([]byte("replay-payee"))

	msg := NewMsgRun(addr, std.Coins{}, []*std.MemFile{{Name: "script.gno", Body: `package main

import (
	"chain"
	"chain/banker"
)

func main(cur realm) {
	b := banker.NewBanker(banker.BankerTypeRealmSend, cur)
	b.SendCoins(cur.Address(), "` + to.String() + `", chain.Coins{{"ugnot", 5}})
}
`}})

	_, err := env.vmk.Run(ctx, msg)
	require.ErrorContains(t, err, "run sends more coins than its max send allows")
	require.Zero(t, env.bankk.GetCoins(ctx, to).AmountOf("ugnot"))

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)
	genesisCtx := replayCtx.WithBlockHeader(&bft.Header{ChainID: ctx.ChainID()})
	_, err = env.vmk.Run(genesisCtx, msg)
	require.ErrorContains(t, err, "run sends more coins than its max send allows")
	require.Zero(t, env.bankk.GetCoins(genesisCtx, to).AmountOf("ugnot"))

	_, err = env.vmk.Run(replayCtx, msg)
	require.NoError(t, err)
	require.Equal(t, int64(5), env.bankk.GetCoins(replayCtx, to).AmountOf("ugnot"))

	msg.MaxSend = std.MustParseCoins("1ugnot")
	_, err = env.vmk.Run(replayCtx, msg)
	require.ErrorContains(t, err, "run sends more coins than its max send allows")
	require.Equal(t, int64(5), env.bankk.GetCoins(replayCtx, to).AmountOf("ugnot"))
}

// A banker spends a user's account only as that user's own run banker, so a
// run banker a realm kept spends nothing in a later message.
func TestSDKBankerRefusesUserAccount(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	pub := ed25519.GenPrivKey().PubKey()
	user := pub.Address()
	acc := env.acck.NewAccountWithAddress(ctx, user)
	require.NoError(t, acc.SetPubKey(pub))
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, user, std.MustParseCoins("100ugnot"))

	pkgAddr := gnolang.DerivePkgCryptoAddr("gno.land/r/test/payer")
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, pkgAddr))
	env.bankk.SetCoins(ctx, pkgAddr, std.MustParseCoins("100ugnot"))

	payee := crypto.AddressFromPreimage([]byte("payee"))
	five := std.MustParseCoins("5ugnot")
	const refused = "a banker can spend a user's account only inside that user's run"

	require.PanicsWithValue(t, refused, func() {
		NewSDKBanker(env.vmk, ctx).SendCoins(user.Bech32(), payee.Bech32(), five)
	})
	other := crypto.AddressFromPreimage([]byte("other-runner"))
	require.PanicsWithValue(t, refused, func() {
		NewRunSDKBanker(env.vmk, ctx, other, five, false).SendCoins(user.Bech32(), payee.Bech32(), five)
	})

	NewSDKBanker(env.vmk, ctx).SendCoins(pkgAddr.Bech32(), payee.Bech32(), five)
	NewRunSDKBanker(env.vmk, ctx, user, five, false).SendCoins(user.Bech32(), payee.Bech32(), five)
	require.Equal(t, int64(10), env.bankk.GetCoins(ctx, payee).AmountOf("ugnot"))
}
