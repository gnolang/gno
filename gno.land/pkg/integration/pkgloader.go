package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// XXX: fully replace with loader

type PkgsLoader struct {
	pkgs    []*packages.Package
	visited map[string]struct{}

	// list of occurrences to patchs with the given value
	// XXX: find a better way
	patchs map[string]string
}

func NewPkgsLoader() *PkgsLoader {
	return &PkgsLoader{
		pkgs:    make([]*packages.Package, 0),
		visited: make(map[string]struct{}),
		patchs:  make(map[string]string),
	}
}

func (pl *PkgsLoader) List() packages.PkgList {
	return pl.pkgs
}

func (pl *PkgsLoader) SetPatch(replace, with string) {
	pl.patchs[replace] = with
}

func (pl *PkgsLoader) LoadPackages(creatorKey crypto.PrivKey, fee std.Fee, deposit std.Coins) ([]gnoland.TxWithMetadata, error) {
	pkgslist, err := pl.List().Sort(true) // sorts packages by their dependencies.
	if err != nil {
		return nil, fmt.Errorf("unable to sort packages: %w", err)
	}

	txs := make([]gnoland.TxWithMetadata, len(pkgslist))
	for i, pkg := range pkgslist {
		tx, err := gnoland.LoadPackage(pkg, creatorKey.PubKey().Address(), fee, deposit)
		if err != nil {
			return nil, fmt.Errorf("unable to load pkg %q: %w", pkg.ImportPath, err)
		}

		// If any replace value is specified, apply them
		if len(pl.patchs) > 0 {
			for _, msg := range tx.Msgs {
				addpkg, ok := msg.(vm.MsgAddPackage)
				if !ok {
					continue
				}

				if addpkg.Package == nil {
					continue
				}

				for _, file := range addpkg.Package.Files {
					for replace, with := range pl.patchs {
						file.Body = strings.ReplaceAll(file.Body, replace, with)
					}
				}
			}
		}

		txs[i] = gnoland.TxWithMetadata{
			Tx: tx,
		}
	}

	if err = gnoland.SignGenesisTxs(txs, creatorKey, "tendermint_test"); err != nil {
		return nil, fmt.Errorf("unable to sign txs: %w", err)
	}

	return txs, nil
}

func (pl *PkgsLoader) LoadAllPackagesFromDir(path string) error {
	// list all packages from target path
	pkgslist, err := packages.Load(nil, filepath.Join(path, "..."))
	if err != nil {
		return fmt.Errorf("listing gno packages: %w", err)
	}

	for _, pkg := range pkgslist {
		if !pl.exist(pkg) {
			pl.add(pkg)
		}
	}

	return nil
}

func (pl *PkgsLoader) LoadPackage(pkgDir string, name string) error {
	examples := filepath.Join(gnoenv.RootDir(), "examples", "...")
	cfg := &packages.LoadConfig{Deps: true, DepsPatterns: []string{examples}}
	pkgs, err := packages.Load(cfg, pkgDir)
	if err != nil {
		return fmt.Errorf("%q: loading: %w", pkgDir, err)
	}

	for _, pkg := range pkgs {
		if pkg.Dir == pkgDir && name != "" {
			pkg.ImportPath = name
		}
		if gnolang.IsStdlib(pkg.ImportPath) {
			continue
		}
		if pkg.Draft {
			continue // Skip draft package
		}
		if pl.exist(pkg) {
			continue
		}
		pl.add(pkg)
	}

	return nil
}

func (pl *PkgsLoader) add(pkg *packages.Package) {
	pl.visited[pkg.ImportPath] = struct{}{}
	pl.pkgs = append(pl.pkgs, pkg)
}

func (pl *PkgsLoader) exist(pkg *packages.Package) (ok bool) {
	_, ok = pl.visited[pkg.ImportPath]
	return
}
