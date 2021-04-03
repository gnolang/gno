package cli

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	abciServer "github.com/tendermint/classic/abci/server"
	tcmd "github.com/tendermint/classic/cmd/tendermint/commands"
	"github.com/tendermint/classic/libs/cli"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/client"
	"github.com/tendermint/classic/sdk/server"
	"github.com/tendermint/classic/sdk/server/mock"
	"github.com/tendermint/classic/sdk/tests"
	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/types/module"
	"github.com/tendermint/classic/sdk/x/genutil"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func TestInitCmd(t *testing.T) {
	defer server.SetupViper(t)()
	defer setupClientHome(t)()
	home, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)

	ctx := server.NewContext(cfg, logger)
	cmd := InitCmd(ctx, testMbm, home)

	require.NoError(t, cmd.RunE(nil, []string{"appnode-test"}))
}

func setupClientHome(t *testing.T) func() {
	clientDir, cleanup := tests.NewTestCaseDir(t)
	viper.Set(flagClientHome, clientDir)
	return cleanup
}

func TestEmptyState(t *testing.T) {
	defer server.SetupViper(t)()
	defer setupClientHome(t)()

	home, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)

	ctx := server.NewContext(cfg, logger)

	cmd := InitCmd(ctx, testMbm, home)
	require.NoError(t, cmd.RunE(nil, []string{"appnode-test"}))

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd = server.ExportCmd(ctx, nil)

	err = cmd.RunE(nil, nil)
	require.NoError(t, err)

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC

	require.Contains(t, out, "genesis_time")
	require.Contains(t, out, "chain_id")
	require.Contains(t, out, "consensus_params")
	require.Contains(t, out, "app_hash")
	require.Contains(t, out, "app_state")
}

func TestStartStandAlone(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	viper.Set(cli.HomeFlag, home)
	defer setupClientHome(t)()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := server.NewContext(cfg, logger)
	initCmd := InitCmd(ctx, testMbm, home)
	require.NoError(t, initCmd.RunE(nil, []string{"appnode-test"}))

	app, err := mock.NewApp(home, logger)
	require.Nil(t, err)
	svrAddr, _, err := server.FreeTCPAddr()
	require.Nil(t, err)
	svr, err := abciServer.NewServer(svrAddr, "socket", app)
	require.Nil(t, err, "error creating listener")
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	timer := time.NewTimer(time.Duration(2) * time.Second)
	select {
	case <-timer.C:
		svr.Stop()
	}
}

func TestInitNodeValidatorFiles(t *testing.T) {
	home, cleanup := tests.NewTestCaseDir(t)
	defer cleanup()
	viper.Set(cli.HomeFlag, home)
	viper.Set(client.FlagName, "moniker")
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(cfg)
	require.Nil(t, err)
	require.NotEqual(t, "", nodeID)
	require.NotEqual(t, 0, len(valPubKey.Bytes()))
}
