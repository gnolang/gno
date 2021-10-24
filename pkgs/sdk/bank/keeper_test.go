package bank

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
)

func TestKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	addr3 := crypto.AddressFromPreimage([]byte("addr3"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bank.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, env.bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, env.bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	// Test AddCoins
	env.bank.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 25))))

	env.bank.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 15)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 15), std.NewCoin("foocoin", 25))))

	// Test SubtractCoins
	env.bank.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	env.bank.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bank.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 11)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bank.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 10)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 1))))

	// Test SendCoins
	env.bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	_ = env.bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bank.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	env.bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := NewInput(addr2, std.NewCoins(std.NewCoin("foocoin", 2)))
	output1 := NewOutput(addr, std.NewCoins(std.NewCoin("foocoin", 2)))
	env.bank.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 7))))
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 8))))

	inputs := []Input{
		NewInput(addr, std.NewCoins(std.NewCoin("foocoin", 3))),
		NewInput(addr2, std.NewCoins(std.NewCoin("barcoin", 3), std.NewCoin("foocoin", 2))),
	}

	outputs := []Output{
		NewOutput(addr, std.NewCoins(std.NewCoin("barcoin", 1))),
		NewOutput(addr3, std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))),
	}
	env.bank.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 21), std.NewCoin("foocoin", 4))))
	require.True(t, env.bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 7), std.NewCoin("foocoin", 6))))
	require.True(t, env.bank.GetCoins(ctx, addr3).IsEqual(std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))))
}

func TestBankKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	bank := NewBankKeeper(env.acck)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, bank.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bank.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, bank.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	env.bank.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))

	// Test SendCoins
	bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	err := bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.True(t, bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bank.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	bank.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, bank.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, bank.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	err = bank.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{"FOOCOIN", -5}})
	require.Error(t, err)
}

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
