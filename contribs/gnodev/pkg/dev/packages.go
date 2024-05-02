package dev

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PackagePath struct {
	Path    string
	Creator crypto.Address
	Deposit std.Coins
}

func ResolvePackagePathQuery(bk *address.Book, path string) (PackagePath, error) {
	var ppath PackagePath

	upath, err := url.Parse(path)
	if err != nil {
		return ppath, fmt.Errorf("malformed path/query: %w", err)
	}
	ppath.Path = filepath.Clean(upath.Path)

	// Check for creator option
	creator := upath.Query().Get("creator")
	if creator != "" {
		address, err := crypto.AddressFromBech32(creator)
		if err != nil {
			var ok bool
			address, ok = bk.GetByName(creator)
			if !ok {
				return ppath, fmt.Errorf("invalid name or address for creator %q", creator)
			}
		}

		ppath.Creator = address
	}

	// Check for deposit option
	deposit := upath.Query().Get("deposit")
	if deposit != "" {
		coins, err := std.ParseCoins(deposit)
		if err != nil {
			return ppath, fmt.Errorf(
				"unable to parse deposit amount %q (should be in the form xxxugnot): %w", deposit, err,
			)
		}

		ppath.Deposit = coins
	}

	return ppath, nil
}

type Package struct {
	gnomod.Pkg
	Creator crypto.Address
	Deposit std.Coins
}

type PackagesMap map[string]Package

var (
	ErrEmptyCreatorPackage = errors.New("no creator specified for package")
	ErrEmptyDepositPackage = errors.New("no deposit specified for package")
)

func NewPackagesMap(ppaths []PackagePath) (PackagesMap, error) {
	pkgs := make(map[string]Package)
	for _, ppath := range ppaths {
		if ppath.Creator.IsZero() {
			return nil, fmt.Errorf("unable to load package %q: %w", ppath.Path, ErrEmptyCreatorPackage)
		}

		abspath, err := filepath.Abs(ppath.Path)
		if err != nil {
			return nil, fmt.Errorf("unable to guess absolute path for %q: %w", ppath.Path, err)
		}

		// list all packages from target path
		pkgslist, err := gnomod.ListPkgs(abspath)
		if err != nil {
			return nil, fmt.Errorf("listing gno packages: %w", err)
		}

		for _, pkg := range pkgslist {
			if pkg.Dir == "" {
				continue
			}

			if _, ok := pkgs[pkg.Dir]; ok {
				continue // skip
			}
			pkgs[pkg.Dir] = Package{
				Pkg:     pkg,
				Creator: ppath.Creator,
				Deposit: ppath.Deposit,
			}
		}
	}

	return pkgs, nil
}

func (pm PackagesMap) toList() gnomod.PkgList {
	list := make([]gnomod.Pkg, 0, len(pm))
	for _, pkg := range pm {
		list = append(list, pkg.Pkg)
	}
	return list
}

func (pm PackagesMap) Load(fee std.Fee) ([]std.Tx, error) {
	pkgs := pm.toList()

	sorted, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("unable to sort pkgs: %w", err)
	}

	nonDraft := sorted.GetNonDraftPkgs()
	txs := []std.Tx{}
	for _, modPkg := range nonDraft {
		pkg := pm[modPkg.Dir]
		if pkg.Creator.IsZero() {
			return nil, fmt.Errorf("no creator set for %q", pkg.Dir)
		}

		// Open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(modPkg.Dir, modPkg.Name)
		if err := memPkg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid package: %w", err)
		}

		// Create transaction
		tx := std.Tx{
			Fee: fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: pkg.Creator,
					Deposit: pkg.Deposit,
					Package: memPkg,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}
