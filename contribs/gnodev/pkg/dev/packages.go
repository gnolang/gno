package dev

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
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
	*packages.Package
	Creator crypto.Address
	Deposit std.Coins
}

type PackagesMap map[string]Package

var (
	ErrEmptyCreatorPackage = errors.New("no creator specified for package")
	ErrEmptyDepositPackage = errors.New("no deposit specified for package")
)

func NewPackagesMap(cfg *packages.LoadConfig, ppaths []PackagePath) (PackagesMap, error) {
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
		pkgslist, err := packages.Load(cfg, filepath.Join(abspath, "..."))
		if err != nil {
			return nil, fmt.Errorf("listing gno packages: %w", err)
		}

		for _, pkg := range pkgslist {
			if pkg.Dir == "" {
				continue
			}

			if pkg.ImportPath != "" && gnolang.IsStdlib(pkg.ImportPath) {
				continue
			}

			if _, ok := pkgs[pkg.Dir]; ok {
				continue // skip
			}
			pkgs[pkg.Dir] = Package{
				Package: pkg,
				Creator: ppath.Creator,
				Deposit: ppath.Deposit,
			}
		}
	}

	return pkgs, nil
}

func (pm PackagesMap) toList() packages.PkgList {
	list := make([]*packages.Package, 0, len(pm))
	for _, pkg := range pm {
		list = append(list, pkg.Package)
	}
	return list
}

func (pm PackagesMap) Load(fee std.Fee, start time.Time) ([]gnoland.TxWithMetadata, error) {
	pkgs := pm.toList()

	sorted, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("unable to sort pkgs: %w", err)
	}

	nonDraft := sorted.GetNonDraftPkgs()

	metatxs := make([]gnoland.TxWithMetadata, 0, len(nonDraft))
	for _, modPkg := range nonDraft {
		pkg := pm[modPkg.Dir]
		if pkg.Creator.IsZero() {
			return nil, fmt.Errorf("no creator set for %q", pkg.Dir)
		}

		// Open files in directory as MemPackage.
		memPkg := gno.MustReadMemPackage(modPkg.Dir, modPkg.ImportPath)
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
		metatx := gnoland.TxWithMetadata{
			Tx: tx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp: start.Unix(),
			},
		}

		metatxs = append(metatxs, metatx)
	}

	return metatxs, nil
}
