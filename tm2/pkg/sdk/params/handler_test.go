package params

import (
	"strings"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	tcs := []struct {
		path     string
		expected string
	}{
		{path: "params/params_test/foo/bar.string", expected: `"baz"`},
		{path: "params/params_test/foo/bar.int64", expected: `"-12345"`},
		{path: "params/params_test/foo/bar.uint64", expected: `"4242"`},
		{path: "params/params_test/foo/bar.bool", expected: "true"},
		{path: "params/params_test/foo/bar.bytes", expected: `"YmF6"`},
	}

	for _, tc := range tcs {
		req := abci.RequestQuery{
			Path: tc.path,
		}
		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)
		require.Nil(t, res.Data)
	}

	env.keeper.SetString(env.ctx, "foo/bar.string", "baz")
	env.keeper.SetInt64(env.ctx, "foo/bar.int64", -12345)
	env.keeper.SetUint64(env.ctx, "foo/bar.uint64", 4242)
	env.keeper.SetBool(env.ctx, "foo/bar.bool", true)
	env.keeper.SetBytes(env.ctx, "foo/bar.bytes", []byte("baz"))

	for _, tc := range tcs {
		req := abci.RequestQuery{
			Path: tc.path,
		}
		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)
		assert.Equal(t, string(res.Data), tc.expected)
	}
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
