package abci

type MockApplication struct {
	BaseApplication
	InfoFn       func(RequestInfo) ResponseInfo
	SetOptionFn  func(RequestSetOption) ResponseSetOption
	QueryFn      func(RequestQuery) ResponseQuery
	CheckTxFn    func(RequestCheckTx) ResponseCheckTx
	InitChainFn  func(RequestInitChain) ResponseInitChain
	BeginBlockFn func(RequestBeginBlock) ResponseBeginBlock
	DeliverTxFn  func(RequestDeliverTx) ResponseDeliverTx
	EndBlockFn   func(RequestEndBlock) ResponseEndBlock
	CommitFn     func() ResponseCommit
	CloseFn      func() error
}

// Info/Query Connection
func (app *MockApplication) Info(req RequestInfo) ResponseInfo {
	if app.InfoFn != nil {
		return app.InfoFn(req)
	}
	return ResponseInfo{}
}

func (app *MockApplication) SetOption(req RequestSetOption) ResponseSetOption {
	if app.SetOptionFn != nil {
		return app.SetOptionFn(req)
	}
	return ResponseSetOption{}
}

func (app *MockApplication) Query(req RequestQuery) ResponseQuery {
	if app.QueryFn != nil {
		return app.QueryFn(req)
	}
	return ResponseQuery{}
}

// Mempool Connection
func (app *MockApplication) CheckTx(req RequestCheckTx) ResponseCheckTx {
	if app.CheckTxFn != nil {
		return app.CheckTxFn(req)
	}
	return ResponseCheckTx{}
}

// Consensus Connection
func (app *MockApplication) InitChain(req RequestInitChain) ResponseInitChain {
	if app.InitChainFn != nil {
		return app.InitChainFn(req)
	}
	return ResponseInitChain{}
}

func (app *MockApplication) BeginBlock(req RequestBeginBlock) ResponseBeginBlock {
	if app.BeginBlockFn != nil {
		return app.BeginBlockFn(req)
	}
	return ResponseBeginBlock{}
}

func (app *MockApplication) DeliverTx(req RequestDeliverTx) ResponseDeliverTx {
	if app.DeliverTxFn != nil {
		return app.DeliverTxFn(req)
	}
	return ResponseDeliverTx{}
}

func (app *MockApplication) EndBlock(req RequestEndBlock) ResponseEndBlock {
	if app.EndBlockFn != nil {
		return app.EndBlockFn(req)
	}
	return ResponseEndBlock{}
}

func (app *MockApplication) Commit() ResponseCommit {
	if app.CommitFn != nil {
		return app.CommitFn()
	}
	return ResponseCommit{}
}

// Cleanup
func (app *MockApplication) Close() error {
	if app.CloseFn != nil {
		return app.CloseFn()
	}
	return nil
}
