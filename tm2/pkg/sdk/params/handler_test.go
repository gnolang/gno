package params

import (
	"fmt"
	"strings"
	"testing"

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
	keeper ParamsKeeper
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
	t.Parallel()

	h := NewHandler(ParamsKeeper{})
	res := h.Process(sdk.Context{}, nil)

	assert.False(t, res.IsOK())
	assert.True(t, strings.Contains(res.Log, errInvalidMsgType.Error()))
}

func TestParamsHandler_Query(t *testing.T) {
	t.Parallel()

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()

		var (
			prefix = "params_test"

			testTable = []struct {
				name string
				path string
			}{
				{
					"empty key",
					fmt.Sprintf("params/%s/", prefix),
				},
				{
					"malformed path, no key",
					fmt.Sprintf("params/%s", prefix),
				},
				{
					"malformed path, no prefix",
					"params",
				},
			}
		)

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

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
		t.Parallel()

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

	t.Run("valid path, value present", func(t *testing.T) {
		t.Parallel()

		var (
			prefix = "params_test"
			key    = "foo/bar.string"
			value  = "baz"

			env = setupTestEnv(t, prefix)
			h   = NewHandler(env.keeper)
		)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("params/%s/%s", prefix, key),
		}

		env.keeper.SetString(env.ctx, key, value)

		res := h.Query(env.ctx, req)
		require.Nil(t, res.Error)
		require.NotNil(t, res)

		assert.Equal(t, fmt.Sprintf("%q", value), string(res.Data))
	})
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv(t, "params_test")
	h := NewHandler(env.keeper)
	req := abci.RequestQuery{
		Path: "params/notfound",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
