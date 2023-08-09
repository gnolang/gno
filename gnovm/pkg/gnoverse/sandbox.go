package gnoverse

import (
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type Sandbox struct {
	db         db.DB
	account    auth.AccountKeeper
	bank       bank.BankKeeper
	vm         vm.VMKeeper
	stdlibsDir string
	maxCycles  int64
}

type SandboxOpts struct {
	DB db.DB
}

func (opts SandboxOpts) Validate() error {
	if opts.DB == nil {
		return fmt.Errorf("missing DB")
	}
	return nil
}

func NewSandbox(opts SandboxOpts) (Sandbox, error) {
	if err := opts.Validate(); err != nil {
		return Sandbox{}, fmt.Errorf("invalid opts: %w", err)
	}

	baseKey := store.NewStoreKey("base")
	iavlKey := store.NewStoreKey("main")
	acctKpr := auth.NewAccountKeeper(iavlKey, ProtoGnoAccount)
	bankKpr := bank.NewBankKeeper(acctKpr)
	stdlibsDir := filepath.Join("..", "..", "stdlibs")
	vm := vm.NewVMKeeper(
		baseKey,
		iavlKey,
		acctKpr,
		bankKpr,
		stdlibsDir,
		maxCycles,
	)
	box := Sandbox{
		db: opts.DB,
		vm: vm,
	}
	return box, nil
}
