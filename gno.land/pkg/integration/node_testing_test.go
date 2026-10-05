package integration

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
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

// TestGenesisParamsReachTheHarness guards a silent trap in the txtar harness.
//
// gnolandCmd merges a script's genesis into the base config by copying named
// fields. LoadGenesisParamsFile writes vm params in from
// gno.land/genesis/genesis_params.toml, so a field the loader sets but the merge
// does not copy is silently replaced by the default -- a txtar test would pass
// while a real chain, which reads the file directly, used a different value.
//
// The loader rejects an unknown key, so the file cannot drift ahead of the
// loader. This closes the other half: the loader cannot drift ahead of the
// merge. Written by diffing a loaded state against a pristine default, so a
// newly handled param shows up here on its own rather than needing to be
// remembered.
func TestGenesisParamsReachTheHarness(t *testing.T) {
	t.Parallel()

	// Fields gnolandCmd copies out of the script genesis into the base config.
	// Keep in step with the merge in testscript_gnoland.go.
	carried := map[string]bool{
		"ChainDomain":     true,
		"SysNamesPkgPath": true,
		"RunSubmitters":   true,
	}

	// A params file whose vm values are deliberately NOT the defaults. Diffing
	// against the defaults then names exactly the fields the loader writes --
	// which a file that happens to match the defaults could never reveal.
	dir := t.TempDir()
	paramFile := filepath.Join(dir, "genesis_params.toml")
	require.NoError(t, os.WriteFile(paramFile, []byte(`["vm"]
  chain_domain = "example.test"
  sysnames_pkgpath = "gno.land/r/sys/othernames"
`), 0o600))

	loaded := gnoland.DefaultGenState()
	require.NoError(t, gnoland.LoadGenesisParamsFile(paramFile, &loaded))

	base := gnoland.DefaultGenState()
	for _, name := range changedParamFields(t, base.VM.Params, loaded.VM.Params) {
		assert.True(t, carried[name],
			"genesis_params.toml can set VM.Params.%s, but gnolandCmd does not "+
				"carry it into the harness genesis -- txtar tests would silently "+
				"use the default while a real chain used the file's value", name)
	}
}

// changedParamFields names the exported fields that differ between two Params.
func changedParamFields(t *testing.T, a, b vm.Params) []string {
	t.Helper()

	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	var out []string
	for i := 0; i < va.NumField(); i++ {
		name := va.Type().Field(i).Name
		if !va.Field(i).CanInterface() {
			continue
		}
		if !reflect.DeepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
			out = append(out, name)
		}
	}
	require.NotEmpty(t, out,
		"the fixture must set something different from the defaults, or this "+
			"test proves nothing")
	return out
}
