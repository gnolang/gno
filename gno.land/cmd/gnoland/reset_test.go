package main

import (
	"path/filepath"
	"testing"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/stretchr/testify/require"

	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

func TestResetAll(t *testing.T) {
	config := cfg.TestConfig()
	dir := t.TempDir()
	config.SetRootDir(dir)
	config.EnsureDirs()

	require.NoError(t, initFilesWithConfig(config))
	pv := privval.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	pv.LastSignState.Height = 10
	pv.Save()

	require.NoError(t, resetAll(config.DBDir(), config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile(), logger))

	require.DirExists(t, config.DBDir())
	require.NoFileExists(t, filepath.Join(config.DBDir(), "block.db"))
	require.NoFileExists(t, filepath.Join(config.DBDir(), "state.db"))
	require.NoFileExists(t, filepath.Join(config.DBDir(), "gnolang.db"))
	require.FileExists(t, config.PrivValidatorStateFile())
	require.FileExists(t, config.GenesisFile())
	pv = privval.LoadFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	require.Equal(t, int64(0), pv.LastSignState.Height)
}

func initFilesWithConfig(config *cfg.Config) error {
	// private validator
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()
	var pv *privval.FilePV
	pv = privval.GenFilePV(privValKeyFile, privValStateFile)
	pv.Save()
	nodeKeyFile := config.NodeKeyFile()
	if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
		return err
	}

	genFile := config.GenesisFile()
	genDoc := bft.GenesisDoc{
		ChainID:         "test-chain-%v",
		GenesisTime:     tmtime.Now(),
		ConsensusParams: bft.DefaultConsensusParams(),
	}
	key := pv.GetPubKey()
	genDoc.Validators = []bft.GenesisValidator{{
		Address: key.Address(),
		PubKey:  key,
		Power:   10,
	}}
	if err := genDoc.SaveAs(genFile); err != nil {
		return err
	}
	return nil
}
