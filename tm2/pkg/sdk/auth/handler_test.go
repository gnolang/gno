package auth

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestInvalidMsg(t *testing.T) {
	t.Parallel()

	h := NewHandler(AccountKeeper{}, GasPriceKeeper{})
	res := h.Process(sdk.NewContext(sdk.RunTxModeDeliver, nil, &bft.Header{ChainID: "test-chain"}, nil), tu.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized auth message type"))
}

func TestQueryAccount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	_, pubkey, addr := tu.KeyTestPubAddr()

	acc := env.acck.NewAccountWithAddress(env.ctx, addr)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s/%s", QueryAccount, addr),
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error) // no account is set and no error returned anyway
	require.NotNil(t, res)

	// set account
	acc.SetAccountNumber(uint64(1))
	acc.SetSequence(uint64(20))
	acc.SetCoins(std.NewCoins(std.NewCoin("foo", 10)))
	acc.SetPubKey(pubkey)
	env.acck.SetAccount(env.ctx, acc)

	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	bz, err := amino.MarshalJSONIndent(acc, "", "  ")
	require.NoError(t, err)

	require.True(t, bytes.Equal(res.Data, bz))
}

func TestQueryGasPrice(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s", QueryGasPrice),
		Data: []byte{},
	}

	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error) // no gasprice is set and no error returned anyway
	require.NotNil(t, res)
	gp := std.GasPrice{
		Gas: 100,
		Price: std.Coin{
			Denom:  "token",
			Amount: 10,
		},
	}
	env.gk.SetGasPrice(env.ctx, gp)

	var gp2 std.GasPrice
	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.NoError(t, amino.UnmarshalJSON(res.Data, &gp2))
	require.True(t, gp == gp2)
}

func TestQuerySessions(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	_, _, masterAddr := tu.KeyTestPubAddr()

	// Create master account.
	masterAcc := env.acck.NewAccountWithAddress(env.ctx, masterAddr)
	env.acck.SetAccount(env.ctx, masterAcc)

	// Query sessions when none exist — should return empty list.
	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s/%s/%s", QueryAccount, masterAddr, QuerySessions),
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)

	var empty []std.Account
	require.NoError(t, amino.UnmarshalJSON(res.Data, &empty))
	require.Empty(t, empty)

	// Create two session accounts.
	_, sessPub1, sessAddr1 := tu.KeyTestPubAddr()
	_, sessPub2, sessAddr2 := tu.KeyTestPubAddr()

	sa1 := env.acck.NewSessionAccount(env.ctx, masterAddr, sessPub1)
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa1)
	sa2 := env.acck.NewSessionAccount(env.ctx, masterAddr, sessPub2)
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa2)

	// Query all sessions.
	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)

	var sessions []std.Account
	require.NoError(t, amino.UnmarshalJSON(res.Data, &sessions))
	require.Len(t, sessions, 2)

	// Verify both session addresses are present.
	addrs := map[string]bool{}
	for _, s := range sessions {
		addrs[s.GetAddress().String()] = true
	}
	require.True(t, addrs[sessAddr1.String()])
	require.True(t, addrs[sessAddr2.String()])
}

func TestQuerySessionAccount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	_, _, masterAddr := tu.KeyTestPubAddr()

	// Create master account.
	masterAcc := env.acck.NewAccountWithAddress(env.ctx, masterAddr)
	env.acck.SetAccount(env.ctx, masterAcc)

	// Create a session account.
	_, sessPub, sessAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(env.ctx, masterAddr, sessPub)
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa)

	// Query existing session.
	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s/%s/%s/%s", QueryAccount, masterAddr, QuerySessionAccount, sessAddr),
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)

	bz, err := amino.MarshalJSONIndent(sa, "", "  ")
	require.NoError(t, err)
	require.True(t, bytes.Equal(res.Data, bz))

	// Query non-existent session.
	_, _, bogusAddr := tu.KeyTestPubAddr()
	req2 := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s/%s/%s/%s", QueryAccount, masterAddr, QuerySessionAccount, bogusAddr),
		Data: []byte{},
	}
	res2 := h.Query(env.ctx, req2)
	require.Error(t, res2.Error)
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	req := abci.RequestQuery{
		Path: "auth/notexist",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
