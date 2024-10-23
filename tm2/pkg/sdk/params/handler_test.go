package params

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
)

func TestInvalidMsg(t *testing.T) {
	t.Parallel()

	h := NewHandler(ParamsKeeper{})
	res := h.Process(sdk.NewContext(sdk.RunTxModeDeliver, nil, &bft.Header{ChainID: "test-chain"}, nil), tu.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized params message type"))
}

func TestQuery(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.keeper)

	req := abci.RequestQuery{
		Path: "params/params_test/foo/bar.string",
	}

	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.Nil(t, res.Data)

	env.keeper.SetString(env.ctx, "foo/bar.string", "baz")

	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.Equal(t, string(res.Data), `"baz"`)
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.keeper)
	req := abci.RequestQuery{
		Path: "params/notfound",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
