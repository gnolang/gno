package txs

import (
	"context"
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
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Add_Packages(t *testing.T) {
	t.Parallel()

	keyFromMnemonic := func(mnemonic string) crypto.PrivKey {
		t.Helper()

		// Generate seed from mnemonic
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
		require.NoError(t, err)

		// Derive Private Key
		hdPath := hd.NewFundraiserParams(0, crypto.CoinType, 0)
		masterPriv, ch := hd.ComputeMastersFromSeed(seed)

		derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath.String())
		require.NoError(t, err)

		// Convert to secp256k1 private key
		return secp256k1.PrivKeySecp256k1(derivedPriv)
	}

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

		genesis := common.DefaultGenesis()
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

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()

		var (
			keybaseDir = t.TempDir()
			name       = "beep-boop"
		)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
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
			name, // non-existent key name
			"--gno-home",
			keybaseDir,
			"--insecure-password-stdin",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "Key "+name+" not found")
	})

	t.Run("existing key, invalid password", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
			keybaseDir  = t.TempDir()
			name        = "beep-boop"
			password    = "somepass"
		)

		createValidFile(t, dir, packagePath)

		// Create key
		kb, err := keys.NewKeyBaseFromDir(keybaseDir)
		require.NoError(t, err)
		mnemonic, err := client.GenerateMnemonic(256)
		require.NoError(t, err)
		_, err = kb.CreateAccount(name, mnemonic, "", password, 0, 0)
		require.NoError(t, err)

		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n",
					password+"wrong", // invalid password
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
			name,
			"--gno-home",
			keybaseDir,
			"--insecure-password-stdin",
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.Error(t, cmdErr)
	})

	t.Run("existing key, valid password", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
			keybaseDir  = t.TempDir()
			name        = "beep-boop"
			password    = "somepass"
		)

		createValidFile(t, dir, packagePath)

		// Create key
		kb, err := keys.NewKeyBaseFromDir(keybaseDir)
		require.NoError(t, err)
		info, err := kb.CreateAccount(name, defaultAccount_Seed, "", password, 0, 0)
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
			name,
			"--gno-home",
			keybaseDir,
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

		tx := state.Txs[0].Tx

		msgAddPkg, ok := tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)

		signPayload, err := tx.GetSignBytes(common.DefaultChainID, 0, 0)
		require.NoError(t, err)

		pubKey := info.GetPubKey()
		assert.True(t, pubKey.Equals(tx.Signatures[0].PubKey))
		assert.True(t, pubKey.VerifyBytes(signPayload, tx.Signatures[0].Signature))

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})

	t.Run("ok default key", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := keyFromMnemonic(defaultAccount_Seed)

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

		tx := state.Txs[0].Tx

		msgAddPkg, ok := tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)

		signPayload, err := tx.GetSignBytes(common.DefaultChainID, 0, 0)
		require.NoError(t, err)

		pubKey := key.PubKey()
		assert.True(t, pubKey.Equals(tx.Signatures[0].PubKey))
		assert.True(t, pubKey.VerifyBytes(signPayload, tx.Signatures[0].Signature))

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})

	t.Run("valid package", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := keyFromMnemonic(defaultAccount_Seed)

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

		tx := state.Txs[0].Tx

		msgAddPkg, ok := tx.Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)

		signPayload, err := tx.GetSignBytes(common.DefaultChainID, 0, 0)
		require.NoError(t, err)

		pubKey := key.PubKey()
		assert.True(t, pubKey.Equals(tx.Signatures[0].PubKey))
		assert.True(t, pubKey.VerifyBytes(signPayload, tx.Signatures[0].Signature))

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
