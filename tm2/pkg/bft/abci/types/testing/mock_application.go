package testing

import abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"

type MockApplication struct {
	abci.BaseApplication
	InfoFn       func(abci.RequestInfo) abci.ResponseInfo
	SetOptionFn  func(abci.RequestSetOption) abci.ResponseSetOption
	QueryFn      func(abci.RequestQuery) abci.ResponseQuery
	CheckTxFn    func(abci.RequestCheckTx) abci.ResponseCheckTx
	InitChainFn  func(abci.RequestInitChain) abci.ResponseInitChain
	BeginBlockFn func(abci.RequestBeginBlock) abci.ResponseBeginBlock
	DeliverTxFn  func(abci.RequestDeliverTx) abci.ResponseDeliverTx
	EndBlockFn   func(abci.RequestEndBlock) abci.ResponseEndBlock
	CommitFn     func() abci.ResponseCommit
	CloseFn      func() error
}

// Info/Query Connection
func (app *MockApplication) Info(req abci.RequestInfo) abci.ResponseInfo {
	if app.InfoFn != nil {
		return app.InfoFn(req)
	}
	return abci.ResponseInfo{}
}

func (app *MockApplication) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	if app.SetOptionFn != nil {
		return app.SetOptionFn(req)
	}
	return abci.ResponseSetOption{}
}

func (app *MockApplication) Query(req abci.RequestQuery) abci.ResponseQuery {
	if app.QueryFn != nil {
		return app.QueryFn(req)
	}
	return abci.ResponseQuery{}
}

// Mempool Connection
func (app *MockApplication) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	if app.CheckTxFn != nil {
		return app.CheckTxFn(req)
	}
	return abci.ResponseCheckTx{}
}

// Consensus Connection
func (app *MockApplication) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	if app.InitChainFn != nil {
		return app.InitChainFn(req)
	}
	return abci.ResponseInitChain{}
}

func (app *MockApplication) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	if app.BeginBlockFn != nil {
		return app.BeginBlockFn(req)
	}
	return abci.ResponseBeginBlock{}
}

func (app *MockApplication) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	if app.DeliverTxFn != nil {
		return app.DeliverTxFn(req)
	}
	return abci.ResponseDeliverTx{}
}

func (app *MockApplication) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	if app.EndBlockFn != nil {
		return app.EndBlockFn(req)
	}
	return abci.ResponseEndBlock{}
}

func (app *MockApplication) Commit() abci.ResponseCommit {
	if app.CommitFn != nil {
		return app.CommitFn()
	}
	return abci.ResponseCommit{}
}

// Cleanup
func (app *MockApplication) Close() error {
	if app.CloseFn != nil {
		return app.CloseFn()
	}
	return nil
}
