package mock

import (
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"

	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	FireEventDelegate      func(events.Event)
	AddListenerDelegate    func(string, events.EventCallback)
	RemoveListenerDelegate func(string)
)

type EventSwitch struct {
	service.BaseService

	FireEventFn      FireEventDelegate
	AddListenerFn    AddListenerDelegate
	RemoveListenerFn RemoveListenerDelegate
}

func (m *EventSwitch) FireEvent(ev events.Event) {
	if m.FireEventFn != nil {
		m.FireEventFn(ev)
	}
}

func (m *EventSwitch) AddListener(
	listenerID string,
	cb events.EventCallback,
) {
	if m.AddListenerFn != nil {
		m.AddListenerFn(listenerID, cb)
	}
}

func (m *EventSwitch) RemoveListener(listenerID string) {
	if m.RemoveListenerFn != nil {
		m.RemoveListenerFn(listenerID)
	}
}

type VMKeeper struct {
	AddPackageFn                func(sdk.Context, vm.MsgAddPackage) error
	CallFn                      func(sdk.Context, vm.MsgCall) (string, error)
	QueryEvalFn                 func(sdk.Context, string, string) (string, error)
	QueryFuncsFn                func(sdk.Context, string) (vm.FunctionSignatures, error)
	QueryPathsFn                func(sdk.Context, string, int) ([]string, error)
	QueryFileFn                 func(sdk.Context, string) (string, error)
	QueryDocFn                  func(sdk.Context, string) (*doc.JSONDocumentation, error)
	QueryStorageFn              func(sdk.Context, string) (string, error)
	RunFn                       func(sdk.Context, vm.MsgRun) (string, error)
	LoadStdlibFn                func(sdk.Context, string)
	LoadStdlibCachedFn          func(sdk.Context, string)
	MakeGnoTransactionStoreFn   func(sdk.Context) sdk.Context
	CommitGnoTransactionStoreFn func(sdk.Context)
}

func (m *VMKeeper) AddPackage(ctx sdk.Context, msg vm.MsgAddPackage) error {
	if m.AddPackageFn != nil {
		return m.AddPackageFn(ctx, msg)
	}

	return nil
}

func (m *VMKeeper) Call(ctx sdk.Context, msg vm.MsgCall) (res string, err error) {
	if m.CallFn != nil {
		return m.CallFn(ctx, msg)
	}

	return "", nil
}

func (m *VMKeeper) QueryEval(ctx sdk.Context, pkgPath, expr string) (res string, err error) {
	if m.QueryEvalFn != nil {
		return m.QueryEvalFn(ctx, pkgPath, expr)
	}

	return "", nil
}

func (m *VMKeeper) QueryFuncs(ctx sdk.Context, pkgPath string) (vm.FunctionSignatures, error) {
	if m.QueryFuncsFn != nil {
		return m.QueryFuncsFn(ctx, pkgPath)
	}

	return vm.FunctionSignatures{}, nil
}

func (m *VMKeeper) QueryPaths(ctx sdk.Context, target string, limit int) ([]string, error) {
	if m.QueryPathsFn != nil {
		return m.QueryPathsFn(ctx, target, limit)
	}

	return nil, nil
}

func (m *VMKeeper) QueryFile(ctx sdk.Context, filepath string) (string, error) {
	if m.QueryFileFn != nil {
		return m.QueryFileFn(ctx, filepath)
	}

	return "", nil
}

func (m *VMKeeper) QueryDoc(ctx sdk.Context, pkgPath string) (*doc.JSONDocumentation, error) {
	if m.QueryDocFn != nil {
		return m.QueryDocFn(ctx, pkgPath)
	}

	return nil, nil
}

func (m *VMKeeper) QueryStorage(ctx sdk.Context, pkgPath string) (string, error) {
	if m.QueryStorageFn != nil {
		return m.QueryStorageFn(ctx, pkgPath)
	}

	return "", nil
}

func (m *VMKeeper) Run(ctx sdk.Context, msg vm.MsgRun) (res string, err error) {
	if m.RunFn != nil {
		return m.RunFn(ctx, msg)
	}

	return "", nil
}

func (m *VMKeeper) LoadStdlib(ctx sdk.Context, stdlibDir string) {
	if m.LoadStdlibFn != nil {
		m.LoadStdlibFn(ctx, stdlibDir)
	}
}

func (m *VMKeeper) LoadStdlibCached(ctx sdk.Context, stdlibDir string) {
	if m.LoadStdlibCachedFn != nil {
		m.LoadStdlibCachedFn(ctx, stdlibDir)
	}
}

func (m *VMKeeper) MakeGnoTransactionStore(ctx sdk.Context) sdk.Context {
	if m.MakeGnoTransactionStoreFn != nil {
		return m.MakeGnoTransactionStoreFn(ctx)
	}
	return ctx
}

func (m *VMKeeper) CommitGnoTransactionStore(ctx sdk.Context) {
	if m.CommitGnoTransactionStoreFn != nil {
		m.CommitGnoTransactionStoreFn(ctx)
	}
}

func (m *VMKeeper) InitGenesis(ctx sdk.Context, gs vm.GenesisState) {}

type BankKeeper struct{}

func (m *BankKeeper) InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error {
	return nil
}

func (m *BankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *BankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *BankKeeper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	return nil, nil
}

func (m *BankKeeper) AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	return nil, nil
}

func (m *BankKeeper) InitGenesis(ctx sdk.Context, data bank.GenesisState)     {}
func (m *BankKeeper) GetParams(ctx sdk.Context) bank.Params                   { return bank.Params{} }
func (m *BankKeeper) GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins { return nil }
func (m *BankKeeper) SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *BankKeeper) HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	return true
}

type AuthKeeper struct{}

func (m *AuthKeeper) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) std.Account {
	return nil
}
func (m *AuthKeeper) GetAccount(ctx sdk.Context, addr crypto.Address) std.Account     { return nil }
func (m *AuthKeeper) GetAllAccounts(ctx sdk.Context) []std.Account                    { return nil }
func (m *AuthKeeper) SetAccount(ctx sdk.Context, acc std.Account)                     {}
func (m *AuthKeeper) IterateAccounts(ctx sdk.Context, process func(std.Account) bool) {}
func (m *AuthKeeper) InitGenesis(ctx sdk.Context, data auth.GenesisState)             {}
func (m *AuthKeeper) GetParams(ctx sdk.Context) auth.Params                           { return auth.Params{} }

type ParamsKeeper struct{}

func (m *ParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string)    {}
func (m *ParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64)      {}
func (m *ParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64)    {}
func (m *ParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool)        {}
func (m *ParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte)     {}
func (m *ParamsKeeper) GetStrings(ctx sdk.Context, key string, ptr *[]string) {}

func (m *ParamsKeeper) SetString(ctx sdk.Context, key string, value string)    {}
func (m *ParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64)      {}
func (m *ParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64)    {}
func (m *ParamsKeeper) SetBool(ctx sdk.Context, key string, value bool)        {}
func (m *ParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte)     {}
func (m *ParamsKeeper) SetStrings(ctx sdk.Context, key string, value []string) {}

func (m *ParamsKeeper) Has(ctx sdk.Context, key string) bool             { return false }
func (m *ParamsKeeper) GetRaw(ctx sdk.Context, key string) []byte        { return nil }
func (m *ParamsKeeper) SetRaw(ctx sdk.Context, key string, value []byte) {}

func (m *ParamsKeeper) GetStruct(ctx sdk.Context, key string, strctPtr any) {}
func (m *ParamsKeeper) SetStruct(ctx sdk.Context, key string, strct any)    {}

func (m *ParamsKeeper) GetAny(ctx sdk.Context, key string) any        { return nil }
func (m *ParamsKeeper) SetAny(ctx sdk.Context, key string, value any) {}

type GasPriceKeeper struct{}

func (m *GasPriceKeeper) LastGasPrice(ctx sdk.Context) std.GasPrice    { return std.GasPrice{} }
func (m *GasPriceKeeper) SetGasPrice(ctx sdk.Context, gp std.GasPrice) {}
func (m *GasPriceKeeper) UpdateGasPrice(ctx sdk.Context)               {}

type (
	LastBlockHeightDelegate func() int64
	LoggerDelegate          func() *slog.Logger
)

type EndBlockerApp struct {
	LastBlockHeightFn LastBlockHeightDelegate
	LoggerFn          LoggerDelegate
}

func (m *EndBlockerApp) LastBlockHeight() int64 {
	if m.LastBlockHeightFn != nil {
		return m.LastBlockHeightFn()
	}

	return 0
}

func (m *EndBlockerApp) Logger() *slog.Logger {
	if m.LoggerFn != nil {
		return m.LoggerFn()
	}

	return log.NewNoopLogger()
}
