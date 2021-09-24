package vm

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, creator crypto.Address, pkgPath string, files []NamedFile) error
	Eval(ctx sdk.Context, caller crypto.Address, pkgPath string, expr string) (string, error)
}

var _ VMKeeperI = VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	key  store.StoreKey
	acck auth.AccountKeeper
	bank bank.BankKeeper

	// TODO: remove these and fully implement persistence.
	// For now, the whole chain must be re-run with each reboot.
	fs    *dbm.FSDB // XXX hack -- not immutable store.
	store gno.Store // XXX hack -- in mem only.
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(key store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper) VMKeeper {
	fs := dbm.NewFSDB("_testdata")  // XXX hack
	store := gno.NewCacheStore(nil) // XXX hack
	return VMKeeper{
		key:   key,
		acck:  acck,
		bank:  bank,
		fs:    fs,
		store: store,
	}
}

// AddPackage adds a package with given fileset.
func (vm VMKeeper) AddPackage(ctx sdk.Context, creator crypto.Address, pkgPath string, files []NamedFile) error {
	// Validate arguments.
	if creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	creatorAcc := vm.acck.GetAccount(ctx, creator)
	if creatorAcc == nil {
		return std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", creator))
	}
	if pkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}
	if pv := vm.store.GetPackage(pkgPath); pv != nil {
		// XXX hack, not immutable store.  For
		// re-running txs from block 1, do nothing.
		// In the future, this would return an error.
	} else {
		// TODO check to ensure that creator can pay.
		// TODO deduct price from creator.
		// Add files to global. NOTE: hack
		for _, file := range files {
			name := file.Name
			body := file.Body
			fpath := path.Join(pkgPath, name)
			vm.fs.Set([]byte(fpath), []byte(body))
		}
		// Parse and run the files, construct *PV.
		pkgName := gno.Name("")
		fnodes := []*gno.FileNode{}
		for i, file := range files {
			if filepath.Ext(file.Name) != ".go" {
				continue
			}
			fnode := gno.MustParseFile(file.Name, file.Body)
			if i == 0 {
				pkgName = fnode.PkgName
			} else if fnode.PkgName != pkgName {
				panic(fmt.Sprintf(
					"expected package name %q but got %v",
					pkgName,
					fnode.PkgName))
			}
			fnodes = append(fnodes, fnode)
		}
		pkg := gno.NewPackageNode(pkgName, pkgPath, nil)
		rlm := gno.NewRealm(pkgPath)
		pv := pkg.NewPackage(rlm)
		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				Package: pv,
				Output:  nil, // XXX
				Store:   vm.store,
			})
		m2.RunFiles(fnodes...)
		// Set package to store.
		vm.store.SetPackage(pv)
		return nil
	}
	return nil
}

// Eval evaluates gno expression.
func (vm VMKeeper) Eval(ctx sdk.Context, caller crypto.Address, pkgPath string, expr string) (res string, err error) {
	// Get Package.
	pv := vm.store.GetPackage(pkgPath)
	if pv == nil {
		err = ErrInvalidPkgPath("package not found")
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			Package: pv,
			Output:  nil,
			Store:   vm.store,
		})
	rtv := m.Eval(xx)
	res = rtv.String()
	return res, nil
	// TODO pay for gas? TODO see context?
	/*
		_, err := vm.SubtractCoins(ctx, caller, amt)
		if err != nil {
			return err
		}
	*/
}
