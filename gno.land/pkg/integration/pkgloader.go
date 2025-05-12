package integration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PkgsLoader struct {
	pkgs    []gnomod.Pkg
	visited map[string]struct{}

	// list of occurrences to patchs with the given value
	// XXX: find a better way
	patchs map[string]string
}

func NewPkgsLoader() *PkgsLoader {
	return &PkgsLoader{
		pkgs:    make([]gnomod.Pkg, 0),
		visited: make(map[string]struct{}),
		patchs:  make(map[string]string),
	}
}

func (pl *PkgsLoader) List() gnomod.PkgList {
	return pl.pkgs
}

func (pl *PkgsLoader) SetPatch(replace, with string) {
	pl.patchs[replace] = with
}

func (pl *PkgsLoader) LoadPackages(creatorKey crypto.PrivKey, fee std.Fee, deposit std.Coins) ([]gnoland.TxWithMetadata, error) {
	pkgslist, err := pl.List().Sort() // sorts packages by their dependencies.
	if err != nil {
		return nil, fmt.Errorf("unable to sort packages: %w", err)
	}

	txs := make([]gnoland.TxWithMetadata, len(pkgslist))
	for i, pkg := range pkgslist {
		tx, err := gnoland.LoadPackage(pkg, creatorKey.PubKey().Address(), fee, deposit)
		if err != nil {
			return nil, fmt.Errorf("unable to load pkg %q: %w", pkg.Name, err)
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
	pkgslist, err := gnomod.ListPkgs(path)
	if err != nil {
		return fmt.Errorf("listing gno packages from gnomod: %w", err)
	}

	for _, pkg := range pkgslist {
		if !pl.exist(pkg) {
			pl.add(pkg)
		}
	}

	return nil
}

func (pl *PkgsLoader) LoadPackage(modroot string, path, name string) error {
	// Initialize a queue with the root package
	queue := []gnomod.Pkg{{Dir: path, Name: name}}

	for len(queue) > 0 {
		// Dequeue the first package
		currentPkg := queue[0]
		queue = queue[1:]

		if currentPkg.Dir == "" {
			return fmt.Errorf("no path specified for package")
		}

		if currentPkg.Name == "" {
			// Load `gno.mod` information
			gnoModPath := filepath.Join(currentPkg.Dir, "gno.mod")
			gm, err := gnomod.ParseGnoMod(gnoModPath)
			if err != nil {
				return fmt.Errorf("unable to load %q: %w", gnoModPath, err)
			}
			gm.Sanitize()

			// Override package info with mod infos
			currentPkg.Name = gm.Module.Mod.Path
			currentPkg.Draft = gm.Draft

			pkg, err := gnolang.ReadMemPackage(currentPkg.Dir, currentPkg.Name)
			if err != nil {
				return fmt.Errorf("unable to read package at %q: %w", currentPkg.Dir, err)
			}
			importsMap, err := packages.Imports(pkg, nil)
			if err != nil {
				return fmt.Errorf("unable to load package imports in %q: %w", currentPkg.Dir, err)
			}
			imports := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest)
			for _, imp := range imports {
				if imp.PkgPath == currentPkg.Name || gnolang.IsStdlib(imp.PkgPath) {
					continue
				}
				currentPkg.Imports = append(currentPkg.Imports, imp.PkgPath)
			}
		}

		if currentPkg.Draft {
			continue // Skip draft package
		}

		if pl.exist(currentPkg) {
			continue
		}
		pl.add(currentPkg)

		// Add requirements to the queue
		for _, pkgPath := range currentPkg.Imports {
			fullPath := filepath.Join(modroot, pkgPath)
			queue = append(queue, gnomod.Pkg{Dir: fullPath})
		}
	}

	return nil
}

func (pl *PkgsLoader) add(pkg gnomod.Pkg) {
	pl.visited[pkg.Name] = struct{}{}
	pl.pkgs = append(pl.pkgs, pkg)
}

func (pl *PkgsLoader) exist(pkg gnomod.Pkg) (ok bool) {
	_, ok = pl.visited[pkg.Name]
	return
}
