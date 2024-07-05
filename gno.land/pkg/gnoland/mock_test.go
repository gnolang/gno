package gnoland

import (
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/service"
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

type (
	addPackageDelegate func(sdk.Context, vm.MsgAddPackage) error
	callDelegate       func(sdk.Context, vm.MsgCall) (string, error)
	queryEvalDelegate  func(sdk.Context, string, string) (string, error)
	runDelegate        func(sdk.Context, vm.MsgRun) (string, error)
)

type mockVMKeeper struct {
	addPackageFn addPackageDelegate
	callFn       callDelegate
	queryFn      queryEvalDelegate
	runFn        runDelegate
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

func (m *mockVMKeeper) QueryEval(ctx sdk.Context, pkgPath, expr string) (res string, err error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, pkgPath, expr)
	}

	return "", nil
}

func (m *mockVMKeeper) Run(ctx sdk.Context, msg vm.MsgRun) (res string, err error) {
	if m.runFn != nil {
		return m.runFn(ctx, msg)
	}

	return "", nil
}

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
