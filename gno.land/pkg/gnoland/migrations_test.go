package gnoland

import (
	"errors"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/migrations"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
)

// TestUpgradesRegistry pins that every registered upgrade names a release tag
// halt_min_version can carry, in strictly increasing order.
func TestUpgradesRegistry(t *testing.T) {
	t.Parallel()

	require.NoError(t, migrations.Validate(Upgrades))
	prev := ""
	for _, u := range Upgrades {
		v, ok := parseReleaseVersion(u.Version)
		require.True(t, ok, "upgrade %q: not a release tag", u.Version)
		if prev != "" {
			require.Positive(t, semver.Compare(v, prev), "upgrade %q: not after the previous one", u.Version)
		}
		prev = v
	}
}

// TestBeginBlocker pins which block runs a migration. Not parallel: it swaps
// the Upgrades registry.
func TestBeginBlocker(t *testing.T) {
	const version = "v1.8.0"
	registry := Upgrades
	t.Cleanup(func() { Upgrades = registry })
	run := func(t *testing.T, haltHeight, height int64, haltVersion string, handler migrations.Handler) {
		t.Helper()
		Upgrades = []migrations.Upgrade{{Version: version, Handler: handler}}
		prmk := &mockConfigurableParamsKeeper{
			int64s:  map[string]int64{nodeParamHaltHeight: haltHeight},
			strings: map[string]string{nodeParamHaltMinVersion: haltVersion},
		}
		bb := BeginBlocker(migrations.Env{Params: prmk}, "")
		ctx := sdk.Context{}.WithLogger(log.NewNoopLogger()).WithBlockHeader(&bft.Header{Height: height})
		bb(ctx, abci.RequestBeginBlock{})
	}
	counting := func(calls *int) migrations.Handler {
		return func(sdk.Context, migrations.Env) error { *calls++; return nil }
	}

	t.Run("runs in the block after the halt", func(t *testing.T) {
		var calls int
		run(t, 100, 101, version, counting(&calls))
		assert.Equal(t, 1, calls)
	})

	t.Run("runs in no other block", func(t *testing.T) {
		var calls int
		for _, h := range []int64{99, 100, 102} {
			run(t, 100, h, version, counting(&calls))
		}
		assert.Zero(t, calls)
	})

	t.Run("no halt never matches block 1", func(t *testing.T) {
		var calls int
		run(t, 0, 1, version, counting(&calls))
		assert.Zero(t, calls)
	})

	t.Run("unregistered or empty version runs nothing", func(t *testing.T) {
		var calls int
		run(t, 100, 101, "v1.9.0", counting(&calls))
		run(t, 100, 101, "", counting(&calls))
		assert.Zero(t, calls)
	})

	t.Run("a failed migration stops the block", func(t *testing.T) {
		failing := func(sdk.Context, migrations.Env) error { return errors.New("boom") }
		assert.PanicsWithValue(t, "upgrade v1.8.0: migration failed at height 101: boom", func() {
			run(t, 100, 101, version, failing)
		})
	})
}

// TestMigrationRunsAfterHalt pins the whole path: a governance halt at H, a
// restart on a binary registering a migration under halt_min_version, and
// the migration running once, in block H+1, while block sync defers commits
// — iterating the database, which nothing held yet allows.
//
// Not parallel: it swaps the Upgrades registry, which parallel tests only
// read once the sequential ones are done.
func TestMigrationRunsAfterHalt(t *testing.T) {
	db := memdb.NewMemDB()
	app, err := NewAppWithOptions(TestAppOptions(db))
	require.NoError(t, err)
	baseApp := app.(*sdk.BaseApp)

	privKey := getDummyKey(t)
	addr := privKey.PubKey().Address()
	gen := gnoGenesisState(t)
	gen.Balances = []Balance{{Address: addr, Amount: std.Coins{{Denom: "ugnot", Amount: 10_000_000_000}}}}
	res := baseApp.InitChain(abci.RequestInitChain{
		AppState:        gen,
		ChainID:         "test-chain",
		ConsensusParams: &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 3_000_000_000}},
	})
	require.True(t, res.IsOK(), "%v", res.Error)
	baseApp.Commit()

	block := func(app *sdk.BaseApp, height int64, seq uint64, msgs ...std.Msg) {
		t.Helper()
		app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: "test-chain", Height: height}})
		if len(msgs) > 0 {
			txBytes, err := amino.Marshal(createAndSignTxWithAccSeq(t, msgs, "test-chain", privKey, 0, seq))
			require.NoError(t, err)
			res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
			require.True(t, res.IsOK(), "block %d: %s", height, res.Log)
		}
		app.EndBlock(abci.RequestEndBlock{Height: height})
		app.Commit()
	}

	// A realm writing the halt params as r/sys/params does, which only a realm
	// at that path may. The test binary's version is "develop".
	const path = "gno.land/r/sys/params"
	block(baseApp, 1, 0, vm.NewMsgAddPackage(addr, path, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
		{Name: "halt.gno", Body: `package params

import prms "sys/params"

func SetHalt(cur realm, height int64, version string) {
	prms.SetSysParamInt64("node", "p", "halt_height", height)
	prms.SetSysParamString("node", "p", "halt_min_version", version)
}
`},
	}))
	block(baseApp, 2, 1, vm.NewMsgCall(addr, nil, path, "SetHalt", []string{"3", "develop"}))
	block(baseApp, 3, 2) // the halt block: BeginBlock 4 would panic here

	var calls, keys int
	registry := Upgrades
	t.Cleanup(func() { Upgrades = registry })
	Upgrades = []migrations.Upgrade{{
		Version: "develop",
		Handler: func(ctx sdk.Context, env migrations.Env) error {
			calls++
			it := ctx.Store(env.BaseKey).Iterator(nil, nil, nil)
			defer it.Close()
			for ; it.Valid(); it.Next() {
				keys++
			}
			return nil
		},
	}}
	restarted, err := NewAppWithOptions(TestAppOptions(db))
	require.NoError(t, err)
	baseApp = restarted.(*sdk.BaseApp)

	require.True(t, baseApp.DeferCommits(true))
	require.NotPanics(t, func() { block(baseApp, 4, 3) })
	assert.Equal(t, 1, calls, "migration runs in block H+1")
	assert.Positive(t, keys, "migration iterated the base store")

	block(baseApp, 5, 3) // commits are held now; the migration must not run again
	assert.Equal(t, 1, calls, "migration runs once")
	require.NoError(t, baseApp.WriteCommits())
	require.False(t, baseApp.DeferCommits(false))
}
