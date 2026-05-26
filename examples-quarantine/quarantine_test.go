// Package quarantine drives gno tests for the quarantined examples set.
// See README.md for the rule on what lives in examples/ vs examples-quarantine/.
//
// The driver builds a unified package list spanning examples/ and
// examples-quarantine/ so cross-tree imports resolve, then runs `gno test`
// against each quarantined package.
//
// TestQuarantineRealms runs unit tests across the whole quarantine set.
// TestQuarantineRealmsLoad boots an in-memory gnoland node and loads every
// quarantined package at genesis, asserting that AddPackage succeeds for all.
package quarantine_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestQuarantineRealms runs `gno test`-equivalent against every package under
// examples-quarantine/. We can't invoke the `gno test` CLI directly because
// it's rooted in a single workspace (gnowork.toml) and can't resolve
// quarantine→safe cross-tree imports without reaching for the remote package
// fetcher. Here, opts.Packages carries the quarantine set so the runtime
// resolver finds quarantine→quarantine imports, and WithExamples=true adds
// examples/ as the fallback for quarantine→safe imports.
func TestQuarantineRealms(t *testing.T) {
	rootdir := gnoenv.RootDir()
	quarantineDir := filepath.Join(rootdir, "examples-quarantine")

	quarPkgs, err := packages.ReadPkgListFromDir(quarantineDir, gno.MPUserAll)
	require.NoError(t, err)

	pkgs := make(packages.PkgList, 0, len(quarPkgs))
	for _, p := range quarPkgs {
		pkgs = append(pkgs, &packages.Package{Dir: p.Dir, ImportPath: p.Name})
	}

	opts := &test.TestOptions{
		RootDir: rootdir,
		Output:  os.Stdout,
		Error:   os.Stderr,
	}
	opts.BaseStore, opts.TestStore = test.StoreWithOptions(
		rootdir, opts.WriterForStore(),
		test.StoreOptions{WithExamples: true, Testing: true, Packages: pkgs},
	)
	opts.Verbose = testing.Verbose()

	for _, p := range quarPkgs {
		t.Run(p.Name, func(t *testing.T) {
			mpkg, err := gno.ReadMemPackage(p.Dir, p.Name, gno.MPAnyAll)
			require.NoError(t, err)

			hasTests := slices.ContainsFunc(mpkg.Files, func(f *std.MemFile) bool {
				return strings.HasSuffix(f.Name, "_test.gno") ||
					strings.HasSuffix(f.Name, "_filetest.gno")
			})
			if !hasTests {
				t.Skip("no test files")
			}

			if err := test.Test(mpkg, p.Dir, opts); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestQuarantineRealmsLoad boots an in-memory gnoland node and asserts that
// every quarantined package deploys cleanly at genesis.
//
// The default handler (PanicOnFailingTxResultHandler) aborts genesis on the
// first failing AddPackage tx. We swap in a collecting handler so a single
// CI run surfaces every load failure instead of one at a time.
func TestQuarantineRealmsLoad(t *testing.T) {
	rootdir := gnoenv.RootDir()
	creator := crypto.MustAddressFromString(integration.DefaultAccount_Address)
	quarantineTxs := integration.LoadQuarantinePackages(t, creator, rootdir)

	logger := log.NewTestingLogger(t)
	config, _ := integration.TestingNodeConfig(t, rootdir, quarantineTxs...)
	config.InitChainerConfig.GenesisTxResultHandler = func(_ sdk.Context, tx std.Tx, res sdk.Result) {
		if !res.IsErr() {
			return
		}

		for _, msg := range tx.Msgs {
			if addPkg, ok := msg.(vmm.MsgAddPackage); ok && addPkg.Package != nil {
				pkgPath := addPkg.Package.Path
				t.Errorf("AddPackage failed at genesis: %s: %s", pkgPath, res.Log)
				return
			}
		}

		t.Errorf("Msgs failed at genesis: %s", res.Log)
	}

	node, _ := integration.TestingInMemoryNode(t, logger, config)
	t.Cleanup(func() {
		if err := node.Stop(); err != nil {
			t.Logf("node.Stop: %v", err)
		}
	})
}
