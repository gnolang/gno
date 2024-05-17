package gno

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"sync"

	readme "github.com/gnolang/gno"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gno.me/state"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/db/goleveldb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
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
	Create(ctx context.Context, code string, isPackage, syncable bool) (string, error)
	CreateMemPackage(ctx context.Context, memPackage *std.MemPackage) error
	Call(
		ctx context.Context,
		appName string,
		isPackage bool,
		functionName string,
		args ...string,
	) (res string, events []*state.Event, err error)
	Run(ctx context.Context, code string) (res string, err error)
	ApplyEvent(ctx context.Context, event *state.Event) error
	QueryMemPackages(ctx context.Context) <-chan *std.MemPackage
	QueryMemPackage(ctx context.Context, appName string) *std.MemPackage
}

type VMKeeper struct {
	sync.Mutex
	instance *vm.VMKeeper
	store    types.CommitMultiStore
}

func NewVM() (*VMKeeper, bool) {
	db, err := goleveldb.NewGoLevelDB("gno.me", "./gno.me")
	if err != nil {
		panic("could not ascertain storage: " + err.Error())
	}

	firstStartup := !db.Has([]byte("vminit"))
	if firstStartup {
		db.Set([]byte("vminit"), []byte("true"))
	}

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	acck := auth.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bank.NewBankKeeper(acck)
	vmk := vm.NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, "", 10_000_000)

	vmk.Initialize(ms)
	newVM := &VMKeeper{instance: vmk, store: ms}
	newVM.store.Commit()

	if firstStartup {
		fmt.Println("Installing bundled gno packages...")
		if err := newVM.installExamplePackages(); err != nil {
			panic("could not install example packages: " + err.Error())
		}

		fmt.Println("Initializing event store...")
		if err := newVM.initEventStore(); err != nil {
			panic("could not initialize event store: " + err.Error())
		}
	}

	return newVM, firstStartup
}

func getPackagename(code string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.PackageClauseOnly)
	if err != nil {
		return "", fmt.Errorf("error getting package name: %w", err)
	}

	return file.Name.Name, nil
}

func (v *VMKeeper) installExamplePackages() error {
	pkgs, err := gnomod.ListPkgsWithFS(readme.Examples)
	if err != nil {
		return err
	}

	sortedPkgs, err := pkgs.Sort()
	if err != nil {
		return err
	}

	nonDraftPkgs := sortedPkgs.GetNonDraftPkgs()
	for _, pkg := range nonDraftPkgs {
		memPkg := gnolang.ReadMemPackageFS(readme.Examples, pkg.Dir, pkg.Name)
		if err := memPkg.Validate(); err != nil {
			return err
		}

		msg := vm.MsgAddPackage{
			Package: memPkg,
		}

		if err := v.instance.AddPackage(sdk.Context{}, msg); err != nil {
			return err
		}

		v.store.Commit()
	}

	return nil
}
