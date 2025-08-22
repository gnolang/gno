package params

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsGetCmd(t *testing.T) {
	t.Parallel()

	makeGenesis := func(t *testing.T) (string, gnoland.GnoGenesisState, func()) {
		t.Helper()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		genesis := common.DefaultGenesis()
		appState := genesis.AppState.(gnoland.GnoGenesisState)
		appState.Auth.Params.MaxMemoBytes = 12345
		appState.Auth.Params.FeeCollector = [20]byte{1, 2, 3}
		appState.VM.Params.ChainDomain = "gno.land"
		appState.Bank.Params.RestrictedDenoms = []string{"ugnot"}
		genesis.AppState = appState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		return tempGenesis.Name(), appState, cleanup
	}

	t.Run("get all params", func(t *testing.T) {
		t.Parallel()
		genPath, _, cleanup := makeGenesis(t)
		t.Cleanup(cleanup)

		cfg := &paramsCfg{}
		cfg.GenesisPath = genPath

		var out bytes.Buffer
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(&out))

		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{})
		require.NoError(t, err)

		// Output should be valid JSON and contain all sections
		var decoded map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
		assert.Contains(t, out.String(), "auth")
		assert.Contains(t, out.String(), "vm")
		assert.Contains(t, out.String(), "bank")
	})

	t.Run("get auth section", func(t *testing.T) {
		t.Parallel()
		genPath, appState, cleanup := makeGenesis(t)
		t.Cleanup(cleanup)

		cfg := &paramsCfg{}
		cfg.GenesisPath = genPath
		var out bytes.Buffer
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(&out))

		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{"auth"})
		require.NoError(t, err)

		var decoded map[string]any
		require.NoError(t, json.Unmarshal(out.Bytes(), &decoded))
		assert.EqualValues(t, float64(appState.Auth.Params.MaxMemoBytes), decoded["max_memo_bytes"])
		assert.Contains(t, decoded, "fee_collector")
	})

	t.Run("get leaf value", func(t *testing.T) {
		t.Parallel()
		genPath, _, cleanup := makeGenesis(t)
		t.Cleanup(cleanup)

		cfg := &paramsCfg{}
		cfg.GenesisPath = genPath

		var out bytes.Buffer
		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(&out))

		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{"auth.max_memo_bytes"})
		require.NoError(t, err)

		// Should be a single value (not JSON object)
		val := strings.TrimSpace(out.String())
		assert.Equal(t, "12345", val)
	})

	t.Run("invalid key", func(t *testing.T) {
		t.Parallel()
		genPath, _, cleanup := makeGenesis(t)
		t.Cleanup(cleanup)

		cfg := &paramsCfg{}
		cfg.GenesisPath = genPath

		io := commands.NewTestIO()
		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{"foo.bar"})
		assert.Error(t, err)
	})

	t.Run("invalid genesis path", func(t *testing.T) {
		t.Parallel()
		cfg := &paramsCfg{}
		cfg.GenesisPath = "not-a-file"
		io := commands.NewTestIO()
		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{})
		assert.ErrorContains(t, err, "unable to load genesis")
	})

	t.Run("too many arguments", func(t *testing.T) {
		t.Parallel()
		genPath, _, cleanup := makeGenesis(t)
		t.Cleanup(cleanup)

		cfg := &paramsCfg{}
		cfg.GenesisPath = genPath

		io := commands.NewTestIO()
		cmd := newParamsGetCmd(cfg, io)
		err := cmd.ParseAndRun(context.Background(), []string{"auth", "extra"})
		assert.ErrorContains(t, err, "invalid number of params get arguments")
	})
}
