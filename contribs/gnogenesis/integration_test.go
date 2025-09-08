package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This integration test try to ensure `gnogenesis` is up to date against gno.land, and
// the generated `genesis.json` is valid and can be loaded by the node.
func TestIntegration(t *testing.T) {
	const (
		chainid     = "test"
		genesisTime = "1750402800" // Friday, June 20th 2025 09:00 GMT+0200 (Central European Summer Time)
	)

	ctx := context.Background()

	logger := log.NewTestingLogger(t)
	gnoroot := gnoenv.RootDir()

	genesisfile := filepath.Join(t.TempDir(), "genesis.json")

	// Generate file
	runGnoGenesisCommand(t, ctx, "generate",
		"-chain-id", chainid,
		"-genesis-time", genesisTime,
		"-output-path", genesisfile,
	)

	// Create keybase
	gnoHomeDir := filepath.Join(t.TempDir(), "gno")

	sk := ed25519.GenPrivKey()
	myVal := bft.NewMockPVWithPrivKey(sk)
	kb, err := keys.NewKeyBaseFromDir(gnoHomeDir)
	require.NoError(t, err)
	kb.ImportPrivKey("mykey", sk, "")

	// Add my validator
	runGnoGenesisCommand(t, ctx, "validator", "add",
		"--name", "myval",
		"--genesis-path", genesisfile,
		"--address", myVal.PubKey().Address().String(),
		"--pub-key", myVal.PubKey().String(),
		"--power", "1",
	)

	// Dummy account
	dKeys := common.DummyKeys(t, 3)

	// Generate balance sheet
	defaultBalanceAmount := std.Coins{std.NewCoin(ugnot.Denom, 10e8)}
	balances := []string{
		fmt.Sprintf("%s=%s", myVal.PubKey().Address().String(), defaultBalanceAmount.String()),
		fmt.Sprintf("%s=%s", dKeys[0].Address().String(), defaultBalanceAmount.String()),
		fmt.Sprintf("%s=%s", dKeys[1].Address().String(), defaultBalanceAmount.String()),
		fmt.Sprintf("%s=%s", dKeys[2].Address().String(), defaultBalanceAmount.String()),
	}

	balanceSheet := filepath.Join(t.TempDir(), "balance-sheet.txt")
	err = os.WriteFile(balanceSheet, []byte(strings.Join(balances, "\n")), 0o644)
	require.NoError(t, err)

	// Add balance sheet
	runGnoGenesisCommand(t, ctx, "balances", "add",
		"--genesis-path", genesisfile,
		"--balance-sheet", balanceSheet,
	)

	// Set some whitelisted addresses
	runGnoGenesisCommand(t, ctx, "params", "set",
		"--genesis-path", genesisfile,
		"auth.unrestricted_addrs",
		dKeys[0].Address().String(),
		dKeys[1].Address().String(),
		dKeys[2].Address().String(),
	)

	// Set restricted denom
	runGnoGenesisCommand(t, ctx, "params", "set",
		"--genesis-path", genesisfile,
		"bank.restricted_denoms", "ugnot",
	)

	// XXX(gfanton): add some validators, they should load a genesis as well and we
	// should start them in parallel

	io := commands.NewTestIO()

	// XXX: Add the whole example folders (?), keep this commented until we
	// feel it's useful, other integrations tests should already cover this.
	//
	// io.SetIn(strings.NewReader("\n")) // send an empty password
	// examplesDir := filepath.Join(gnoroot, "examples")
	// runGnoGenesisCommandWithIO(t, ctx, io, "txs", "add", "packages",
	// 	"--insecure-password-stdin",
	// 	"--key-name", "mykey",
	// 	"--gno-home", gnoHomeDir,
	// 	"--genesis-path", genesisfile,
	// 	examplesDir)

	// Add dummy bar package
	barPkg := std.MemPackage{
		Name: "bar",
		Path: "gno.land/r/dev/bar",
		Files: []*std.MemFile{
			{
				Name: "bar.gno",
				Body: `package bar
func Render(_ string) string { return "bar" }
`,
			},
			{
				Name: "gnomod.toml",
				Body: gnolang.GenGnoModLatest("gno.land/r/dev/bar"),
			},
		},
	}
	barDir := t.TempDir()
	err = barPkg.WriteTo(barDir)
	require.NoError(t, err)

	io.SetIn(strings.NewReader("\n")) // send an empty password
	runGnoGenesisCommandWithIO(t, ctx, io, "txs", "add", "packages",
		"--insecure-password-stdin",
		"--key-name", "mykey",
		"--gno-home", gnoHomeDir,
		"--genesis-path", genesisfile,
		barDir)

	genesis, err := bft.GenesisDocFromFile(genesisfile)
	require.NoError(t, err)

	file, err := os.ReadFile(genesisfile)
	require.NoError(t, err)
	fmt.Println(string(file))

	// Ensure genesis is valid
	err = genesis.Validate()
	require.NoError(t, err)

	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.PrivValidator = myVal
	cfg.Genesis = genesis

	node, address := integration.TestingInMemoryNode(t, logger, cfg)
	defer node.Stop()

	cli, err := client.NewHTTPClient(address)
	require.NoError(t, err)

	s, err := cli.Status(ctx, nil)
	require.NoError(t, err)

	assert.Equal(t, s.NodeInfo.Network, chainid)

	// call bar package
	res, err := cli.ABCIQuery(ctx, "vm/qrender", []byte("gno.land/r/dev/bar:"))
	require.NoError(t, err)
	require.NoError(t, res.Response.Error)
	assert.Equal(t, string(res.Response.Data), "bar")

	// ... XXX: ensure everything else is correctly setup
}

func runGnoGenesisCommand(t *testing.T, ctx context.Context, args ...string) {
	t.Helper()
	runGnoGenesisCommandWithIO(t, ctx, commands.NewTestIO(), args...)
}

func runGnoGenesisCommandWithIO(t *testing.T, ctx context.Context, io commands.IO, args ...string) {
	t.Helper()

	t.Logf("running: gnogenesis %s", strings.Join(args, " "))
	cmd := newGenesisCmd(io)
	err := cmd.ParseAndRun(ctx, args)
	require.NoError(t, err)
}
