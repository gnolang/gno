package txs

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Add_Packages(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("invalid package dir", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidPackageDir.Error())
	})

	t.Run("non existent key", func(t *testing.T) {
		t.Parallel()
		keybaseDir := t.TempDir()
		keyname := "beep-boop"

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n",
					"password",
				),
			),
		)
		// Create the command
		cmd := NewTxsCmd(io)
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			t.TempDir(), // package dir
			"--key-name",
			keyname, // non-existent key name
			"--gno-home",
			keybaseDir, // temporaryDir for keybase
			"--insecure-password-stdin",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "Key "+keyname+" not found")
	})

	t.Run("existent key wrong password", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
			keybaseDir  = t.TempDir()
			keyname     = "beep-boop"
			password    = "somepass"
		)
		createValidFile(t, dir, packagePath)
		// Create key
		kb, err := keys.NewKeyBaseFromDir(keybaseDir)
		require.NoError(t, err)
		mnemonic, err := client.GenerateMnemonic(256)
		require.NoError(t, err)
		_, err = kb.CreateAccount(keyname, mnemonic, "", password+"wrong", 0, 0)
		require.NoError(t, err)

		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n",
					password,
				),
			),
		)

		// Create the command
		cmd := NewTxsCmd(io)
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			"--key-name",
			keyname, // non-existent key name
			"--gno-home",
			keybaseDir, // temporaryDir for keybase
			"--insecure-password-stdin",
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.Equal(t, cmdErr.Error(), "unable to sign txs, unable sign tx invalid account password")
	})

	t.Run("existent key correct password", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
			keybaseDir  = t.TempDir()
			keyname     = "beep-boop"
			password    = "somepass"
		)
		createValidFile(t, dir, packagePath)
		// Create key
		kb, err := keys.NewKeyBaseFromDir(keybaseDir)
		require.NoError(t, err)
		_, err = kb.CreateAccount(keyname, DefaultAccount_Seed, "", password, 0, 0)
		require.NoError(t, err)

		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n",
					password,
				),
			),
		)

		// Create the command
		cmd := NewTxsCmd(io)
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			"--key-name",
			keyname, // non-existent key name
			"--gno-home",
			keybaseDir, // temporaryDir for keybase
			"--insecure-password-stdin",
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		require.Equal(t, 1, len(state.Txs))
		require.Equal(t, 1, len(state.Txs[0].Tx.Msgs))

		msgAddPkg, ok := state.Txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)
		require.Equal(t, "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj", state.Txs[0].Tx.Signatures[0].PubKey.String())
		require.Equal(t, "cfe5a15d8def04cbdaf9d08e2511db7928152b26419c4577cbfa282c83118852411f3de5d045ce934555572c21bda8042ce5c64b793a01748e49cf2cff7c2983", hex.EncodeToString(state.Txs[0].Tx.Signatures[0].Signature))

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})
	t.Run("ok default key", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
			keybaseDir  = t.TempDir()
		)
		createValidFile(t, dir, packagePath)

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			"--gno-home",
			keybaseDir, // temporaryDir for keybase
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		require.Equal(t, 1, len(state.Txs))
		require.Equal(t, 1, len(state.Txs[0].Tx.Msgs))

		msgAddPkg, ok := state.Txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)
		require.Equal(t, "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj", state.Txs[0].Tx.Signatures[0].PubKey.String())
		require.Equal(t, "cfe5a15d8def04cbdaf9d08e2511db7928152b26419c4577cbfa282c83118852411f3de5d045ce934555572c21bda8042ce5c64b793a01748e49cf2cff7c2983", hex.EncodeToString(state.Txs[0].Tx.Signatures[0].Signature))

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})

	t.Run("valid package", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
		)
		createValidFile(t, dir, packagePath)

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		require.Equal(t, 1, len(state.Txs))
		require.Equal(t, 1, len(state.Txs[0].Tx.Msgs))

		msgAddPkg, ok := state.Txs[0].Tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)
		require.Equal(t, "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj", state.Txs[0].Tx.Signatures[0].PubKey.String())
		require.Equal(t, "cfe5a15d8def04cbdaf9d08e2511db7928152b26419c4577cbfa282c83118852411f3de5d045ce934555572c21bda8042ce5c64b793a01748e49cf2cff7c2983", hex.EncodeToString(state.Txs[0].Tx.Signatures[0].Signature))

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})
}

func createValidFile(t *testing.T, dir string, packagePath string) {
	t.Helper()
	createFile := func(path, data string) {
		file, err := os.Create(path)
		require.NoError(t, err)

		_, err = file.WriteString(data)
		require.NoError(t, err)
	}

	// Create the gno.mod file
	createFile(
		filepath.Join(dir, "gno.mod"),
		fmt.Sprintf("module %s\n", packagePath),
	)

	// Create a simple main.gno
	createFile(
		filepath.Join(dir, "main.gno"),
		"package cuttlas\n\nfunc Example() string {\nreturn \"Manos arriba!\"\n}",
	)
}
