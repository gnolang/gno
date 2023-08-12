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
	// Mandatory fields.
	// ---

	DB db.DB

	// Common configuration fields with good defaults.
	// ---

	StdlibsDir string
	MaxCycles  int64

	// State represents the runtime components. They are initialized by this
	// library but exported to make them available for debugging.
	State struct {
		initialized bool
		Account     auth.AccountKeeper
		Bank        bank.BankKeeper
		VM          *vm.VMKeeper
	}
}

const (
	defaultMaxCycles int64 = 10000
)

func (s Sandbox) applyDefaultsAndValidate() error {
	return nil
}

// Initialize the Sandbox.
func (s *Sandbox) Init() error {
	if s.State.initialized {
		return fmt.Errorf("already initialized")
	}
	// apply defaults and validate
	if s.DB == nil {
		return fmt.Errorf("missing DB")
	}
	if s.MaxCycles == 0 {
		s.MaxCycles = defaultMaxCycles
	}
	if s.StdlibsDir == "" {
		s.StdlibsDir = filepath.Join("..", "gnovm", "stdlibs") // TODO: smarter
	}

	// initialize components
	baseKey := store.NewStoreKey("base")
	iavlKey := store.NewStoreKey("main")
	acctKpr := auth.NewAccountKeeper(iavlKey, protoAccount)
	bankKpr := bank.NewBankKeeper(acctKpr)
	vm := vm.NewVMKeeper(
		baseKey,
		iavlKey,
		acctKpr,
		bankKpr,
		s.StdlibsDir,
		s.MaxCycles,
	)

	// configure state
	s.State.Account = acctKpr
	s.State.Bank = bankKpr
	s.State.VM = vm
	return nil
}

// TODO: func (s *Sandbox) HandleTx()
// TODO: func (s *Sandbox) HandleQuery()
