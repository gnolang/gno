package vm

/*

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
)

func TestViewKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	view := NewViewKeeper(env.acck)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bank.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))
}
*/
