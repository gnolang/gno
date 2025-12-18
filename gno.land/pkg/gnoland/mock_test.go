package gnoland

import (
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
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
	fireEventDelegate      func(events.Event)
	addListenerDelegate    func(string, events.EventCallback)
	removeListenerDelegate func(string)
)

type mockEventSwitch struct {
	service.BaseService

	fireEventFn      fireEventDelegate
	addListenerFn    addListenerDelegate
	removeListenerFn removeListenerDelegate
}

func (m *mockEventSwitch) FireEvent(ev events.Event) {
	if m.fireEventFn != nil {
		m.fireEventFn(ev)
	}
}

func (m *mockEventSwitch) AddListener(
	listenerID string,
	cb events.EventCallback,
) {
	if m.addListenerFn != nil {
		m.addListenerFn(listenerID, cb)
	}
}

func (m *mockEventSwitch) RemoveListener(listenerID string) {
	if m.removeListenerFn != nil {
		m.removeListenerFn(listenerID)
	}
}

type mockVMKeeper struct {
	addPackageFn                func(sdk.Context, vm.MsgAddPackage) error
	callFn                      func(sdk.Context, vm.MsgCall) (string, error)
	queryFn                     func(sdk.Context, string, string, vm.QueryFormat) (string, error)
	runFn                       func(sdk.Context, vm.MsgRun) (string, error)
	loadStdlibFn                func(sdk.Context, string)
	loadStdlibCachedFn          func(sdk.Context, string)
	makeGnoTransactionStoreFn   func(ctx sdk.Context) sdk.Context
	commitGnoTransactionStoreFn func(ctx sdk.Context)
}

func (m *mockVMKeeper) AddPackage(ctx sdk.Context, msg vm.MsgAddPackage) error {
	if m.addPackageFn != nil {
		return m.addPackageFn(ctx, msg)
	}

	return nil
}

func (m *mockVMKeeper) Call(ctx sdk.Context, msg vm.MsgCall) (res string, err error) {
	if m.callFn != nil {
		return m.callFn(ctx, msg)
	}

	return "", nil
}

func (m *mockVMKeeper) QueryEval(ctx sdk.Context, pkgPath, expr string, format vm.QueryFormat) (res string, err error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, pkgPath, expr, format)
	}

	return "", nil
}

func (m *mockVMKeeper) Run(ctx sdk.Context, msg vm.MsgRun) (res string, err error) {
	if m.runFn != nil {
		return m.runFn(ctx, msg)
	}

	return "", nil
}

func (m *mockVMKeeper) LoadStdlib(ctx sdk.Context, stdlibDir string) {
	if m.loadStdlibFn != nil {
		m.loadStdlibFn(ctx, stdlibDir)
	}
}

func (m *mockVMKeeper) LoadStdlibCached(ctx sdk.Context, stdlibDir string) {
	if m.loadStdlibCachedFn != nil {
		m.loadStdlibCachedFn(ctx, stdlibDir)
	}
}

func (m *mockVMKeeper) MakeGnoTransactionStore(ctx sdk.Context) sdk.Context {
	if m.makeGnoTransactionStoreFn != nil {
		return m.makeGnoTransactionStoreFn(ctx)
	}
	return ctx
}

func (m *mockVMKeeper) CommitGnoTransactionStore(ctx sdk.Context) {
	if m.commitGnoTransactionStoreFn != nil {
		m.commitGnoTransactionStoreFn(ctx)
	}
}

func (m *mockVMKeeper) InitGenesis(ctx sdk.Context, gs vm.GenesisState) {}

type mockBankKeeper struct{}

func (m *mockBankKeeper) InputOutputCoins(ctx sdk.Context, inputs []bank.Input, outputs []bank.Output) error {
	return nil
}

func (m *mockBankKeeper) SendCoins(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsUnrestricted(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *mockBankKeeper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	return nil, nil
}

func (m *mockBankKeeper) AddCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) (std.Coins, error) {
	return nil, nil
}

func (m *mockBankKeeper) InitGenesis(ctx sdk.Context, data bank.GenesisState)     {}
func (m *mockBankKeeper) GetParams(ctx sdk.Context) bank.Params                   { return bank.Params{} }
func (m *mockBankKeeper) GetCoins(ctx sdk.Context, addr crypto.Address) std.Coins { return nil }
func (m *mockBankKeeper) SetCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) error {
	return nil
}

func (m *mockBankKeeper) HasCoins(ctx sdk.Context, addr crypto.Address, amt std.Coins) bool {
	return true
}

type mockAuthKeeper struct{}

func (m *mockAuthKeeper) NewAccountWithAddress(ctx sdk.Context, addr crypto.Address) std.Account {
	return nil
}
func (m *mockAuthKeeper) GetAccount(ctx sdk.Context, addr crypto.Address) std.Account     { return nil }
func (m *mockAuthKeeper) GetAllAccounts(ctx sdk.Context) []std.Account                    { return nil }
func (m *mockAuthKeeper) SetAccount(ctx sdk.Context, acc std.Account)                     {}
func (m *mockAuthKeeper) IterateAccounts(ctx sdk.Context, process func(std.Account) bool) {}
func (m *mockAuthKeeper) InitGenesis(ctx sdk.Context, data auth.GenesisState)             {}
func (m *mockAuthKeeper) GetParams(ctx sdk.Context) auth.Params                           { return auth.Params{} }

type mockParamsKeeper struct{}

func (m *mockParamsKeeper) GetString(ctx sdk.Context, key string, ptr *string)    {}
func (m *mockParamsKeeper) GetInt64(ctx sdk.Context, key string, ptr *int64)      {}
func (m *mockParamsKeeper) GetUint64(ctx sdk.Context, key string, ptr *uint64)    {}
func (m *mockParamsKeeper) GetBool(ctx sdk.Context, key string, ptr *bool)        {}
func (m *mockParamsKeeper) GetBytes(ctx sdk.Context, key string, ptr *[]byte)     {}
func (m *mockParamsKeeper) GetStrings(ctx sdk.Context, key string, ptr *[]string) {}

func (m *mockParamsKeeper) SetString(ctx sdk.Context, key string, value string)    {}
func (m *mockParamsKeeper) SetInt64(ctx sdk.Context, key string, value int64)      {}
func (m *mockParamsKeeper) SetUint64(ctx sdk.Context, key string, value uint64)    {}
func (m *mockParamsKeeper) SetBool(ctx sdk.Context, key string, value bool)        {}
func (m *mockParamsKeeper) SetBytes(ctx sdk.Context, key string, value []byte)     {}
func (m *mockParamsKeeper) SetStrings(ctx sdk.Context, key string, value []string) {}

func (m *mockParamsKeeper) Has(ctx sdk.Context, key string) bool                { return false }
func (m *mockParamsKeeper) GetStruct(ctx sdk.Context, key string, strctPtr any) {}
func (m *mockParamsKeeper) SetStruct(ctx sdk.Context, key string, strct any)    {}

func (m *mockParamsKeeper) GetAny(ctx sdk.Context, key string) any        { return nil }
func (m *mockParamsKeeper) SetAny(ctx sdk.Context, key string, value any) {}

type mockGasPriceKeeper struct{}

func (m *mockGasPriceKeeper) LastGasPrice(ctx sdk.Context) std.GasPrice    { return std.GasPrice{} }
func (m *mockGasPriceKeeper) SetGasPrice(ctx sdk.Context, gp std.GasPrice) {}
func (m *mockGasPriceKeeper) UpdateGasPrice(ctx sdk.Context)               {}

type (
	lastBlockHeightDelegate func() int64
	loggerDelegate          func() *slog.Logger
)

type mockEndBlockerApp struct {
	lastBlockHeightFn lastBlockHeightDelegate
	loggerFn          loggerDelegate
}

func (m *mockEndBlockerApp) LastBlockHeight() int64 {
	if m.lastBlockHeightFn != nil {
		return m.lastBlockHeightFn()
	}

	return 0
}

func (m *mockEndBlockerApp) Logger() *slog.Logger {
	if m.loggerFn != nil {
		return m.loggerFn()
	}

	return log.NewNoopLogger()
}
