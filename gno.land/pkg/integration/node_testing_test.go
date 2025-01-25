package integration

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTestinGenesisState(t *testing.T) {
	// Generate a test private key and address
	privKey := secp256k1.GenPrivKey()
	creatorAddr := privKey.PubKey().Address()

	// Create sample packages
	pkg1 := GenerateMemPackage("pkg1", "file.gno", "package1")
	pkg2 := GenerateMemPackage("pkg2", "file.gno", "package2")

	t.Run("single package genesis", func(t *testing.T) {
		genesis := GenerateTestinGenesisState(privKey, pkg1)

		// Verify transactions
		require.Len(t, genesis.Txs, 1)
		tx := genesis.Txs[0].Tx

		// Check the transaction's message
		require.Len(t, tx.Msgs, 1)
		msg, ok := tx.Msgs[0].(vm.MsgAddPackage)
		require.True(t, ok, "expected MsgAddPackage")
		assert.Equal(t, pkg1, *msg.Package, "package mismatch")

		// Verify transaction signatures
		require.Len(t, tx.Signatures, 1)
		assert.NotEmpty(t, tx.Signatures[0], "signature should not be empty")

		// Verify balances
		require.Len(t, genesis.Balances, 1)
		balance := genesis.Balances[0]
		assert.Equal(t, creatorAddr, balance.Address)
		assert.Equal(t, std.MustParseCoins(ugnot.ValueString(10_000_000_000_000)), balance.Amount)
	})

	t.Run("multiple packages genesis", func(t *testing.T) {
		genesis := GenerateTestinGenesisState(privKey, pkg1, pkg2)

		// Verify two transactions are created
		require.Len(t, genesis.Txs, 2)

		// Check each transaction's package
		for i, expectedPkg := range []gnovm.MemPackage{pkg1, pkg2} {
			tx := genesis.Txs[i].Tx
			require.Len(t, tx.Msgs, 1)
			msg, ok := tx.Msgs[0].(vm.MsgAddPackage)
			require.True(t, ok, "expected MsgAddPackage")
			assert.Equal(t, expectedPkg, *msg.Package, "package mismatch in tx %d", i)
		}
	})
}

func TestGenerateMemPackage(t *testing.T) {
	t.Run("valid file pairs", func(t *testing.T) {
		// Create a MemPackage with valid file pairs
		pkg := GenerateMemPackage(
			"test/path",
			"file1.gno", "content1",
			"file2.gno", "content2",
		)

		// Verify the package metadata
		assert.Equal(t, "path", pkg.Name)
		assert.Equal(t, "test/path", pkg.Path)

		// Verify the included files
		require.Len(t, pkg.Files, 2)
		assert.Equal(t, "file1.gno", pkg.Files[0].Name)
		assert.Equal(t, "content1", pkg.Files[0].Body)
		assert.Equal(t, "file2.gno", pkg.Files[1].Name)
		assert.Equal(t, "content2", pkg.Files[1].Body)
	})

	t.Run("odd number of pairs panics", func(t *testing.T) {
		// Ensure the function panics with odd number of arguments
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for odd number of file pairs")
			}
		}()

		GenerateMemPackage("test/path", "file1.gno") // Invalid: missing content
	})
}
