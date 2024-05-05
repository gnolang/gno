package gno

import (
	"context"
	"path/filepath"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/db/boltdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type (
	MsgAddPackage vm.MsgAddPackage
	MsgCall       vm.MsgCall
)

type VM interface {
	AddPackage(ctx context.Context, msg MsgAddPackage) error
	Call(ctx context.Context, msg MsgCall) (res string, err error)
}

type VMKeeper struct {
	instance *vm.VMKeeper
}

func (v VMKeeper) AddPackage(ctx context.Context, msg MsgAddPackage) error {
	return v.instance.AddPackage(sdk.Context{}.WithContext(ctx), vm.MsgAddPackage(msg))
}

func (v VMKeeper) Call(ctx context.Context, msg MsgCall) (res string, err error) {
	return v.instance.Call(sdk.Context{}.WithContext(ctx), vm.MsgCall(msg))
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

	vmk.Initialize(ms.MultiCacheWrap())
	return VMKeeper{instance: vmk}

	// 	addPkg := vm.MsgAddPackage{
	// 		Package: &std.MemPackage{
	// 			Name: "firstpkg",
	// 			Path: "gno.land/r/firstpkg",
	// 			Files: []*std.MemFile{
	// 				{
	// 					Name: "print.gno",
	// 					Body: `
	// package firstpkg

	// var value string

	// func Print() string {
	// 	return "Amazing!"
	// }

	// func SetValue(s string) {
	// 	value = s
	// }

	// func GetValue() string {
	// 	return value
	// }
	// 					`,
	// 				},
	// 			},
	// 		},
	// 	}
	// 	if err := vmk.AddPackage(sdk.Context{}, addPkg); err != nil {
	// 		panic(err)
	// 	}

	// 	msgCall := vm.MsgCall{
	// 		PkgPath: "gno.land/r/firstpkg",
	// 		Func:    "GetValue",
	// 	}
	// 	res, err := vmk.Call(sdk.Context{}, msgCall)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println(res)

	// 	msgCall = vm.MsgCall{
	// 		PkgPath: "gno.land/r/firstpkg",
	// 		Func:    "SetValue",
	// 		Args:    []string{"Hello, World!"},
	// 	}
	// 	res, err = vmk.Call(sdk.Context{}, msgCall)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println(res)

	// 	msgCall = vm.MsgCall{
	// 		PkgPath: "gno.land/r/firstpkg",
	// 		Func:    "GetValue",
	// 	}
	// 	res, err = vmk.Call(sdk.Context{}, msgCall)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println(res)

	// db.Print()
}

func NewMsgAddPackage(name, code string) MsgAddPackage {
	return MsgAddPackage{
		Package: &std.MemPackage{
			Name: name,
			Path: "gno.land/r/" + name,
			Files: []*std.MemFile{
				{
					Name: name + ".gno",
					Body: code,
				},
			},
		},
	}
}

func NewMsgCall(name, funcName string, args []string) MsgCall {
	return MsgCall{
		PkgPath: "gno.land/r/" + name,
		Func:    funcName,
		Args:    args,
	}
}
