package abci

// Application is an interface that enables any finite, deterministic state
// machine to be driven by a blockchain-based replication engine via the ABCI.
// All methods take a RequestXxx argument and return a ResponseXxx argument,
// except CheckTx/DeliverTx, which take `tx []byte`, and `Commit`, which takes
// nothing.
type Application interface {
	// Info/Query Connection
	Info(RequestInfo) ResponseInfo                // Return application info
	SetOption(RequestSetOption) ResponseSetOption // Set application option
	Query(RequestQuery) ResponseQuery             // Query for state

	// Mempool Connection
	CheckTx(RequestCheckTx) ResponseCheckTx // Validate a tx for the mempool

	// Consensus Connection
	InitChain(RequestInitChain) ResponseInitChain    // Initialize blockchain with validators and other info from TendermintCore
	BeginBlock(RequestBeginBlock) ResponseBeginBlock // Signals the beginning of a block
	DeliverTx(RequestDeliverTx) ResponseDeliverTx    // Deliver a tx for full processing
	EndBlock(RequestEndBlock) ResponseEndBlock       // Signals the end of a block, returns changes to the validator set
	Commit() ResponseCommit                          // Commit the state and return the application Merkle root hash

	// Cleanup
	Close() error
}

//-------------------------------------------------------
// BaseApplication is a base form of Application

var _ Application = (*BaseApplication)(nil)

type BaseApplication struct{}

func NewBaseApplication() *BaseApplication {
	return &BaseApplication{}
}

func (BaseApplication) Info(req RequestInfo) ResponseInfo {
	return ResponseInfo{}
}

func (BaseApplication) SetOption(req RequestSetOption) ResponseSetOption {
	return ResponseSetOption{}
}

func (BaseApplication) DeliverTx(req RequestDeliverTx) ResponseDeliverTx {
	return ResponseDeliverTx{}
}

func (BaseApplication) CheckTx(req RequestCheckTx) ResponseCheckTx {
	return ResponseCheckTx{}
}

func (BaseApplication) Commit() ResponseCommit {
	return ResponseCommit{}
}

func (BaseApplication) Query(req RequestQuery) ResponseQuery {
	return ResponseQuery{}
}

func (BaseApplication) InitChain(req RequestInitChain) ResponseInitChain {
	return ResponseInitChain{}
}

func (BaseApplication) BeginBlock(req RequestBeginBlock) ResponseBeginBlock {
	return ResponseBeginBlock{}
}

func (BaseApplication) EndBlock(req RequestEndBlock) ResponseEndBlock {
	return ResponseEndBlock{}
}

func (BaseApplication) Close() error {
	return nil
}
