package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/crisis/internal/types"
	"github.com/tendermint/classic/sdk/x/params"
)

func testPassingInvariant(_ sdk.Context) (string, bool) {
	return "", false
}

func testFailingInvariant(_ sdk.Context) (string, bool) {
	return "", true
}

func testKeeper(checkPeriod uint) Keeper {
	paramsKeeper := params.NewKeeper(
		sdk.NewKVStoreKey(params.StoreKey), sdk.NewTransientStoreKey(params.TStoreKey), params.DefaultCodespace,
	)

	return NewKeeper(paramsKeeper.Subspace(types.DefaultParamspace), checkPeriod, nil, "test")
}

func TestLogger(t *testing.T) {
	k := testKeeper(5)

	ctx := sdk.Context{}.WithLogger(log.NewNopLogger())
	require.Equal(t, ctx.Logger(), k.Logger(ctx))
}

func TestInvariants(t *testing.T) {
	k := testKeeper(5)
	require.Equal(t, k.InvCheckPeriod(), uint(5))

	k.RegisterRoute("testModule", "testRoute", testPassingInvariant)
	require.Len(t, k.Routes(), 1)
}

func TestAssertInvariants(t *testing.T) {
	k := testKeeper(5)
	ctx := sdk.Context{}.WithLogger(log.NewNopLogger())

	k.RegisterRoute("testModule", "testRoute1", testPassingInvariant)
	require.NotPanics(t, func() { k.AssertInvariants(ctx) })

	k.RegisterRoute("testModule", "testRoute2", testFailingInvariant)
	require.Panics(t, func() { k.AssertInvariants(ctx) })
}
