package gno

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/db/boltdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	AppPrefix string = "gno.land/r/"
	PkgPrefix string = "gno.land/p/"

	gnoFileSuffix string = ".gno"
)

type VM interface {
	Create(ctx context.Context, code string, isPackage bool) error
	Call(ctx context.Context, appName string, isPackage bool, functionName string, args ...string) (res string, err error)
	Run(ctx context.Context, code string) (res string, err error)
}

type VMKeeper struct {
	instance *vm.VMKeeper
	store    types.CommitMultiStore
}

func NewVM() VMKeeper {
	// db := memdb.NewMemDB()
	// DMB: make this actually persist to disk
	db, err := boltdb.New("gno.me", "./gno.me")
	if err != nil {
		panic("could not ascertain storage: " + err.Error())
	}

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	acck := auth.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bank.NewBankKeeper(acck)
	stdlibsDir := filepath.Join("..", "..", "gnovm", "stdlibs")
	vmk := vm.NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, stdlibsDir, 10_000_000)

	vmk.Initialize(ms)
	return VMKeeper{instance: vmk, store: ms}
}

func getPackagename(code string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.PackageClauseOnly)
	if err != nil {
		return "", fmt.Errorf("error getting package name: %w", err)
	}

	return file.Name.Name, nil
}
