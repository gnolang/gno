package main

// start_tmkms_lazy_test.go pins the -lazy + tmkms_listener guard
// (PR #5718 review nit): -lazy derives the genesis validator set from a
// locally-available signing key, but in tmkms_listener mode the key lives in
// tmkms and only signs votes/proposals — NewSignerFromConfig would silently
// fall back to a local key and the node would come up as a non-validator.
// start must refuse this and point the operator at an explicit genesis.

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestStart_LazyRejectsTmkmsListener(t *testing.T) {
	t.Parallel()

	nodeDir := t.TempDir()
	configPath := constructConfigPath(nodeDir)

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(new(bytes.Buffer)))
	io.SetErr(commands.WriteNopCloser(new(bytes.Buffer)))

	// Create a default config (and its directory) via `config init`.
	initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer initCancel()
	require.NoError(t, newRootCmd(io).ParseAndRun(initCtx,
		[]string{"config", "init", "--config-path", configPath}))

	// Enable tmkms_listener. lazyInitNodeDir preserves an existing config, so
	// this survives into LoadConfig during start. (A list field is awkward via
	// `config set`, so edit the struct directly.)
	cfg, err := config.LoadConfig(nodeDir)
	require.NoError(t, err)
	cfg.Consensus.PrivValidator.TmkmsListener.ListenAddr = "tcp://127.0.0.1:0"
	cfg.Consensus.PrivValidator.TmkmsListener.ChainID = "test-chain"
	cfg.Consensus.PrivValidator.TmkmsListener.AllowedKMSPubKeys = []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
	}
	require.NoError(t, config.WriteConfigFile(configPath, cfg))

	// start --lazy with a missing genesis must refuse rather than seed one
	// from the local key. The guard returns before the node starts.
	runCtx, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer runCancel()
	err = newRootCmd(io).ParseAndRun(runCtx, []string{
		"start",
		"--lazy",
		"--data-dir", nodeDir,
		"--genesis", filepath.Join(nodeDir, "genesis.json"),
	})
	require.ErrorIs(t, err, errLazyTmkmsListener)
}
