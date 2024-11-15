package params

import (
	"fmt"
	"strings"
	"testing"
	"time"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

type testEnv struct {
	ctx    sdk.Context
	store  store.Store
	keeper Keeper
}

type testable interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

func setupTestEnv(t testable, prefix string) testEnv {
	var (
		db           = memdb.NewMemDB()
		paramsCapKey = store.NewStoreKey("paramsCapKey")
		ms           = store.NewCommitMultiStore(db)
	)

	ms.MountStoreWithDB(paramsCapKey, iavl.StoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())

	keeper := NewParamsKeeper(paramsCapKey, prefix)

	ctx := sdk.NewContext(
		sdk.RunTxModeDeliver,
		ms,
		&bft.Header{Height: 1, ChainID: "test-chain-id"},
		log.NewNoopLogger(),
	).WithConsensusParams(&abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:    1024,
			MaxDataBytes:  1024 * 100,
			MaxBlockBytes: 1024 * 100,
			MaxGas:        10 * 1000 * 1000,
			TimeIotaMS:    10,
		},
		Validator: &abci.ValidatorParams{
			PubKeyTypeURLs: []string{},
		},
	})

	return testEnv{
		ctx:    ctx,
		store:  ctx.Store(paramsCapKey),
		keeper: keeper,
	}
}

func TestParamsHandler_Process(t *testing.T) {
	h := NewHandler(Keeper{})
	res := h.Process(sdk.Context{}, nil)

	assert.False(t, res.IsOK())
	assert.True(t, strings.Contains(res.Log, errInvalidMsgType.Error()))
}

func TestParamsHandler_Query(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		var (
			prefix = "params_test"

			testTable = []struct {
				name string
				path string
			}{
				{
					"empty path",
					"",
				},
				{
					"empty key",
					fmt.Sprintf("params/%s/", prefix),
				},
				{
					"empty prefix, empty key",
					"params//",
				},
				{
					"malformed path, no key",
					fmt.Sprintf("params/%s", prefix),
				},
				{
					"malformed path, no prefix",
					"params",
				},
				{
					"malformed path, missing prefix",
					fmt.Sprintf("params/%d", time.Now().Unix()),
				},
			}
		)

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				var (
					env = setupTestEnv(t, prefix)
					h   = NewHandler(env.keeper)
				)

				req := abci.RequestQuery{
					Path: testCase.path,
				}

				res := h.Query(env.ctx, req)
				require.False(t, res.IsOK())

				assert.Contains(t, res.Log, errUnknownQueryEndpoint.Error())
				assert.Nil(t, res.Data)
			})
		}
	})

	t.Run("valid path, value missing", func(t *testing.T) {
		var (
			prefix = "params_test"
			key    = "foo/bar.string"

			env = setupTestEnv(t, prefix)
			h   = NewHandler(env.keeper)
		)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("params/%s/%s", prefix, key),
		}

		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)

		assert.Nil(t, res.Data)
	})

	t.Run("valid paths, values present", func(t *testing.T) {
		var (
			module = "params"
			prefix = "params_test"

			env = setupTestEnv(t, prefix)
			h   = NewHandler(env.keeper)
		)

		tcs := []struct {
			path     string
			expected string
		}{
			{path: fmt.Sprintf("%s/%s/foo/bar.string", module, prefix), expected: `"baz"`},
			{path: fmt.Sprintf("%s/%s/foo/bar.int64", module, prefix), expected: `"-12345"`},
			{path: fmt.Sprintf("%s/%s/foo/bar.uint64", module, prefix), expected: `"4242"`},
			{path: fmt.Sprintf("%s/%s/foo/bar.bool", module, prefix), expected: "true"},
			{path: fmt.Sprintf("%s/%s/foo/bar.bytes", module, prefix), expected: `"YmF6"`},
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

		require.NoError(t, env.keeper.SetString(env.ctx, "foo/bar.string", "baz"))
		require.NoError(t, env.keeper.SetInt64(env.ctx, "foo/bar.int64", -12345))
		require.NoError(t, env.keeper.SetUint64(env.ctx, "foo/bar.uint64", 4242))
		require.NoError(t, env.keeper.SetBool(env.ctx, "foo/bar.bool", true))
		require.NoError(t, env.keeper.SetBytes(env.ctx, "foo/bar.bytes", []byte("baz")))

		for _, tc := range tcs {
			req := abci.RequestQuery{
				Path: tc.path,
			}

			res := h.Query(env.ctx, req)
			require.Nil(t, res.Error)
			require.NotNil(t, res)
			assert.Equal(t, string(res.Data), tc.expected)
		}
	})
}
