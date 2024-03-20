package dev

import (
	"fmt"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PkgsMap map[string]gnomod.Pkg

func newPkgsMap(paths []string) (PkgsMap, error) {
	pkgs := make(map[string]gnomod.Pkg)
	for _, path := range paths {
		// list all packages from target path
		pkgslist, err := gnomod.ListPkgs(path)
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
			pkgs[pkg.Dir] = pkg
		}
	}

	// Filter out draft packages.
	return pkgs, nil
}

func (pm PkgsMap) toList() gnomod.PkgList {
	list := make([]gnomod.Pkg, 0, len(pm))
	for _, pkg := range pm {
		list = append(list, pkg)
	}
	return list
}

func (pm PkgsMap) Load(creator bft.Address, fee std.Fee, deposit std.Coins) ([]std.Tx, error) {
	pkgs := pm.toList()

	sorted, err := pkgs.Sort()
	if err != nil {
		return nil, fmt.Errorf("unable to sort pkgs: %w", err)
	}

	nonDraft := sorted.GetNonDraftPkgs()
	txs := []std.Tx{}
	for _, pkg := range nonDraft {
		// Open files in directory as MemPackage.
		memPkg := gno.ReadMemPackage(pkg.Dir, pkg.Name)
		if err := memPkg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid package: %w", err)
		}

		// Create transaction
		tx := std.Tx{
			Fee: fee,
			Msgs: []std.Msg{
				vmm.MsgAddPackage{
					Creator: creator,
					Package: memPkg,
					Deposit: deposit,
				},
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs = append(txs, tx)
	}

	return txs, nil
}
