package params

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestArbitraryParamsQuery(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.keeper)

	tcs := []struct {
		path     string
		expected string
	}{
		{path: "params/" + dummyModuleName + ":bar_string", expected: `"baz"`},
		{path: "params/" + dummyModuleName + ":bar_int64", expected: `"-12345"`},
		{path: "params/" + dummyModuleName + ":bar_uint64", expected: `"4242"`},
		{path: "params/" + dummyModuleName + ":bar_bool", expected: "true"},
		{path: "params/" + dummyModuleName + ":bar_bytes", expected: `baz`},
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

	env.keeper.SetString(env.ctx, dummyModuleName+":bar_string", "baz")
	env.keeper.SetInt64(env.ctx, dummyModuleName+":bar_int64", -12345)
	env.keeper.SetUint64(env.ctx, dummyModuleName+":bar_uint64", 4242)
	env.keeper.SetBool(env.ctx, dummyModuleName+":bar_bool", true)
	env.keeper.SetBytes(env.ctx, dummyModuleName+":bar_bytes", []byte("baz"))

	for _, tc := range tcs {
		req := abci.RequestQuery{
			Path: tc.path,
		}
		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)
		assert.Equal(t, tc.expected, string(res.Data))
	}
}

func TestModuleParamsQuery(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.keeper)
	tcs := []struct {
		path     string
		expected string
	}{
		{path: "params/params_test:foo/bar.string", expected: `"baz"`},
		{path: "params/params_test:foo/bar.int64", expected: `"-12345"`},
		{path: "params/params_test:foo/bar.uint64", expected: `"4242"`},
		{path: "params/params_test:foo/bar.bool", expected: "true"},
		{path: "params/params_test:foo/bar.bytes", expected: `baz`},
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

	env.keeper.SetString(env.ctx, "params_test:foo/bar.string", "baz")
	env.keeper.SetInt64(env.ctx, "params_test:foo/bar.int64", -12345)
	env.keeper.SetUint64(env.ctx, "params_test:foo/bar.uint64", 4242)
	env.keeper.SetBool(env.ctx, "params_test:foo/bar.bool", true)
	env.keeper.SetBytes(env.ctx, "params_test:foo/bar.bytes", []byte("baz"))

	for _, tc := range tcs {
		req := abci.RequestQuery{
			Path: tc.path,
		}
		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)
		assert.Equal(t, tc.expected, string(res.Data))
	}
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()
	env := setupTestEnv()
	h := NewHandler(env.keeper)
	req := abci.RequestQuery{
		Path: "params/notfound:",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
