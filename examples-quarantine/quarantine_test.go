// Package quarantine drives gno tests for the quarantined examples set.
//
// Packages under examples-quarantine/ are not part of the test-13 genesis
// (see misc/quarantine/safe-list.txt for the kept set) but are still
// exercised here so they remain useful as integration test fodder. The driver
// builds a unified package list spanning examples/ and examples-quarantine/
// so cross-tree imports resolve, then runs `gno test` against each
// quarantined package.
//
// TestQuarantineRealms runs unit tests across the whole quarantine set.
// TestQuarantineRealmsLoad boots an in-memory gnoland node and loads every
// quarantined package at genesis, asserting that AddPackage succeeds for all.
package quarantine_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
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
// examples-quarantine/. opts.Packages carries the quarantine set so the
// runtime resolver finds quarantine→quarantine imports; WithExamples=true
// adds examples/ as a fallback for quarantine→safe imports.
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
// every quarantined package deploys cleanly at genesis. integration.
// TestingNodeConfig adds the safe examples/ set via LoadDefaultPackages; we
// append the quarantine txs as additionals so they run on top.
//
// The default handler (PanicOnFailingTxResultHandler) aborts genesis on the
// first failing AddPackage tx. We swap in a collecting handler so a single
// CI run surfaces every load failure instead of one at a time.
func TestQuarantineRealmsLoad(t *testing.T) {
	rootdir := gnoenv.RootDir()
	examplesDir := filepath.Join(rootdir, "examples")
	quarantineDir := filepath.Join(rootdir, "examples-quarantine")

	creator := crypto.MustAddressFromString(integration.DefaultAccount_Address)
	fee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

	quarantineTxs, err := loadQuarantineGenesisTxs(examplesDir, quarantineDir, creator, fee)
	require.NoError(t, err)

	logger := log.NewTestingLogger(t)
	config, _ := integration.TestingNodeConfig(t, rootdir, quarantineTxs...)

	var (
		failuresMu sync.Mutex
		failures   []string
	)
	config.InitChainerConfig.GenesisTxResultHandler = func(_ sdk.Context, tx std.Tx, res sdk.Result) {
		if !res.IsErr() {
			return
		}
		pkgPath := "<unknown>"
		for _, msg := range tx.Msgs {
			if addPkg, ok := msg.(vmm.MsgAddPackage); ok && addPkg.Package != nil {
				pkgPath = addPkg.Package.Path
				break
			}
		}
		failuresMu.Lock()
		failures = append(failures, fmt.Sprintf("%s: %s", pkgPath, res.Log))
		failuresMu.Unlock()
	}

	node, _ := integration.TestingInMemoryNode(t, logger, config)
	t.Cleanup(func() {
		if err := node.Stop(); err != nil {
			t.Logf("node.Stop: %v", err)
		}
	})

	failuresMu.Lock()
	defer failuresMu.Unlock()
	for _, f := range failures {
		t.Errorf("AddPackage failed at genesis: %s", f)
	}
}

// loadQuarantineGenesisTxs builds AddPackage txs for every package under
// quarantineDir, sorted in a dependency-respecting order against the *union*
// of examples and quarantine packages. This is necessary because
// gnoland.LoadPackagesFromDir sorts within a single root and would error on
// quarantine→safe cross-tree imports.
func loadQuarantineGenesisTxs(examplesDir, quarantineDir string, creator crypto.Address, fee std.Fee) ([]gnoland.TxWithMetadata, error) {
	safePkgs, err := packages.ReadPkgListFromDir(examplesDir, gno.MPUserAll)
	if err != nil {
		return nil, err
	}
	quarPkgs, err := packages.ReadPkgListFromDir(quarantineDir, gno.MPUserAll)
	if err != nil {
		return nil, err
	}

	quarDirs := make(map[string]struct{}, len(quarPkgs))
	for _, p := range quarPkgs {
		quarDirs[p.Dir] = struct{}{}
	}

	union := slices.Concat(safePkgs, quarPkgs)
	sorted, err := union.Sort()
	if err != nil {
		return nil, err
	}
	sorted = sorted.GetNonIgnoredPkgs()

	txs := make([]gnoland.TxWithMetadata, 0, len(quarPkgs))
	for _, pkg := range sorted {
		if _, ok := quarDirs[pkg.Dir]; !ok {
			continue
		}
		mpkg, err := gno.ReadMemPackage(pkg.Dir, pkg.Name, gno.MPUserAll)
		if err != nil {
			return nil, fmt.Errorf("read mempackage %q: %w", pkg.Dir, err)
		}

		// Honor [addpkg] creator directives in gnomod.toml (mirrors
		// gnoland.LoadPackagesFromDir). Without this, packages with explicit
		// creator overrides would deploy under the wrong address and any
		// init code that asserts on the caller would be silently bypassed.
		pkgCreator := creator
		if mod, err := gno.ParseCheckGnoMod(mpkg); err == nil && mod != nil && mod.AddPkg.Creator != "" {
			addr, err := crypto.AddressFromBech32(mod.AddPkg.Creator)
			if err != nil {
				return nil, fmt.Errorf("invalid creator address %q in %q: %w", mod.AddPkg.Creator, pkg.Dir, err)
			}
			pkgCreator = addr
		}

		tx, err := gnoland.LoadPackage(mpkg, pkgCreator, fee, nil)
		if err != nil {
			return nil, fmt.Errorf("load package %q: %w", pkg.Dir, err)
		}
		txs = append(txs, gnoland.TxWithMetadata{Tx: tx})
	}
	return txs, nil
}
