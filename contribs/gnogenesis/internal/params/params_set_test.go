package params

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsSetCmd(t *testing.T) {
	t.Parallel()

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		args := []string{"auth.unrestricted_addrs"} // missing value
		err := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, err, "invalid number of params set arguments")
	})

	t.Run("invalid genesis path", func(t *testing.T) {
		t.Parallel()

		cfg := &paramsCfg{}
		cfg.GenesisPath = "invalid-path"

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		args := []string{"auth.unrestricted_addrs", "addr1"}
		err := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, err, "unable to load genesis")
	})

	t.Run("set string field", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.VM.Params.ChainDomain = "gno.land"
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		args := []string{"vm.chain_domain", "gui.land"}
		err := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, err)

		// Reload and check
		updated, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		state := updated.AppState.(gnoland.GnoGenesisState)
		assert.Equal(t, "gui.land", state.VM.Params.ChainDomain)
	})

	t.Run("set crypto.Address field", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.Auth.Params.FeeCollector = crypto.Address{}
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		addr := common.DummyKey(t)
		args := []string{"auth.fee_collector", addr.Address().String()}
		err := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, err)

		updated, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		state := updated.AppState.(gnoland.GnoGenesisState)

		assert.Equal(t, addr.Address().String(), state.Auth.Params.FeeCollector.String())
	})

	t.Run("set []crypto.Address field", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.Auth.Params.UnrestrictedAddrs = []crypto.Address{}
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		addr1, addr2 := common.DummyKey(t), common.DummyKey(t)
		args := []string{"auth.unrestricted_addrs", addr1.Address().String(), addr2.Address().String()}
		err := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, err)

		updated, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		state := updated.AppState.(gnoland.GnoGenesisState)
		require.Len(t, state.Auth.Params.UnrestrictedAddrs, 2)
		assert.Equal(t, addr1.Address().String(), state.Auth.Params.UnrestrictedAddrs[0].String())
		assert.Equal(t, addr2.Address().String(), state.Auth.Params.UnrestrictedAddrs[1].String())
	})

	t.Run("set int field", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.Auth.Params.MaxMemoBytes = 2048
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		args := []string{"auth.max_memo_bytes", "4096"}
		err := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, err)

		updated, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		state := updated.AppState.(gnoland.GnoGenesisState)
		assert.Equal(t, int64(4096), state.Auth.Params.MaxMemoBytes)
	})

	t.Run("set gas price field", func(t *testing.T) {
		t.Parallel()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.Auth.Params.InitialGasPrice = std.GasPrice{
			Gas: 10, Price: std.MustParseCoin("3ugnot"),
		}
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()

		io := commands.NewTestIO()
		cmd := newParamsSetCmd(cfg, io)

		newGas := std.GasPrice{
			Gas: 2000, Price: std.MustParseCoin("400ugnot"),
		}

		args := []string{"auth.initial_gasprice", newGas.String()}
		err := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, err)

		updated, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		state := updated.AppState.(gnoland.GnoGenesisState)
		assert.Equal(t, newGas, state.Auth.Params.InitialGasPrice)
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		// Use a type not supported by saveStringToValue
		type dummy struct {
			Unsupported map[string]string
		}
		val := reflect.ValueOf(&dummy{}).Elem().FieldByName("Unsupported")
		err := saveStringToValue([]string{"foo"}, val)
		assert.ErrorContains(t, err, "unsupported type")
	})

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()
		type testField struct {
			Addr crypto.Address
		}
		params := testField{}
		v := reflect.ValueOf(&params).Elem().FieldByName("Addr")
		err := saveStringToValue([]string{"badaddress"}, v)
		assert.ErrorContains(t, err, "unable to parse address")
	})

	t.Run("invalid int", func(t *testing.T) {
		t.Parallel()
		type testField struct {
			IntVal int
		}
		params := testField{}
		v := reflect.ValueOf(&params).Elem().FieldByName("IntVal")
		err := saveStringToValue([]string{"notanint"}, v)
		assert.ErrorContains(t, err, "invalid character")
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Parallel()
		type testField struct {
			Dur time.Duration
		}
		params := testField{}
		v := reflect.ValueOf(&params).Elem().FieldByName("Dur")
		err := saveStringToValue([]string{"notaduration"}, v)
		assert.ErrorContains(t, err, "unable to parse time.Duration")
	})
}

func Test_saveStringToValue_JSONTypes(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		I int     `json:"i"`
		U uint    `json:"u"`
		F float64 `json:"f"`
		B bool    `json:"b"`
	}

	val := reflect.ValueOf(&testStruct{}).Elem()

	// int
	err := saveStringToValue([]string{"42"}, val.FieldByName("I"))
	require.NoError(t, err)
	assert.Equal(t, 42, int(val.FieldByName("I").Int()))

	// uint
	err = saveStringToValue([]string{"42"}, val.FieldByName("U"))
	require.NoError(t, err)
	assert.Equal(t, uint64(42), val.FieldByName("U").Uint())

	// float
	err = saveStringToValue([]string{"3.14"}, val.FieldByName("F"))
	require.NoError(t, err)
	assert.InDelta(t, 3.14, val.FieldByName("F").Float(), 0.0001)

	// bool
	err = saveStringToValue([]string{"true"}, val.FieldByName("B"))
	require.NoError(t, err)
	assert.True(t, val.FieldByName("B").Bool())
}

func Test_saveStringToValue_noValues(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Foo string
	}
	val := reflect.ValueOf(&testStruct{}).Elem().FieldByName("Foo")
	err := saveStringToValue([]string{}, val)
	assert.ErrorContains(t, err, "no value(s) to set")
}

// TestParamsSetCmd_Validation covers the validation added after the field is
// set, including the bank branch.
//
// The command writes the field by reflection and consults nothing else, so
// without this a bad value reached the genesis file and surfaced as a node
// panic at boot. Validation is scoped to the module named by the key, so an
// unrelated half-filled section cannot block an edit.
func TestParamsSetCmd_Validation(t *testing.T) {
	t.Parallel()

	// run writes genesis, applies mutate, then runs `params set key vals...`.
	run := func(t *testing.T, mutate func(*gnoland.GnoGenesisState), args ...string) error {
		t.Helper()
		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		if mutate != nil {
			mutate(&appState)
		}
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cfg := &paramsCfg{}
		cfg.GenesisPath = tempGenesis.Name()
		cmd := newParamsSetCmd(cfg, commands.NewTestIO())
		return cmd.ParseAndRun(context.Background(), args)
	}

	t.Run("bank rejects an invalid restricted denom", func(t *testing.T) {
		t.Parallel()
		err := run(t, nil, "bank.restricted_denoms", "!!!not-a-denom")
		require.Error(t, err, "an invalid denom must not reach the genesis file")
		assert.ErrorContains(t, err, "invalid restricted denom")
	})

	t.Run("bank accepts a valid restricted denom", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, run(t, nil, "bank.restricted_denoms", "ugnot"))
	})

	t.Run("an unrelated invalid module does not block the edit", func(t *testing.T) {
		t.Parallel()
		// vm is left invalid on purpose; the key being set is a bank one, so
		// validating all three modules would refuse a genesis still being built.
		err := run(t, func(gs *gnoland.GnoGenesisState) {
			gs.VM.Params.StoragePrice = ""
		}, "bank.restricted_denoms", "ugnot")
		assert.NoError(t, err, "validation must be scoped to the module named by the key")
	})
}
