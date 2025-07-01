package integration

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTestingGenesisState(t *testing.T) {
	// Generate a test private key and address
	privKey := secp256k1.GenPrivKey()
	creatorAddr := privKey.PubKey().Address()

	// Create sample packages
	pkg1 := std.MemPackage{
		Name: "pkg1",
		Path: "pkg1",
		Files: []*std.MemFile{
			{Name: "file.gno", Body: "package1"},
		},
	}
	pkg2 := std.MemPackage{
		Name: "pkg2",
		Path: "pkg2",
		Files: []*std.MemFile{
			{Name: "file.gno", Body: "package2"},
		},
	}

	t.Run("single package genesis", func(t *testing.T) {
		genesis := GenerateTestingGenesisState(privKey, pkg1)

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
		genesis := GenerateTestingGenesisState(privKey, pkg1, pkg2)

		// Verify two transactions are created
		require.Len(t, genesis.Txs, 2)

		// Check each transaction's package
		for i, expectedPkg := range []std.MemPackage{pkg1, pkg2} {
			tx := genesis.Txs[i].Tx
			require.Len(t, tx.Msgs, 1)
			msg, ok := tx.Msgs[0].(vm.MsgAddPackage)
			require.True(t, ok, "expected MsgAddPackage")
			assert.Equal(t, expectedPkg, *msg.Package, "package mismatch in tx %d", i)
		}
	})
}
