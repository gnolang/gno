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

// DeferredCommitter is an Application that can hold its commits in memory
// across blocks and write them when the node says so. During block sync the
// node holds its own writes too, and writes all of them at flush points: its
// block store, then its state, then the application (WriteCommits), so that a
// crash at any instant leaves them consistent. It holds its own writes only
// while the application defers its commits: an application on disk ahead of
// the node's records is the one state a restart cannot recover from.
type DeferredCommitter interface {
	// DeferCommits makes Commit hold its writes in memory (on), or write them
	// as before (off), and reports whether the application defers. It is
	// called only between blocks, with nothing held when turned off.
	DeferCommits(on bool) bool

	// WriteCommits writes every held commit as one all-or-nothing write.
	WriteCommits() error

	// HeldBytes returns the bytes of the held writes, which the node keeps
	// under its budget.
	HeldBytes() int64

	// NeedsFlushPoint reports that the node must write everything after the
	// current block: the halt height is reached, or a flush point was required.
	NeedsFlushPoint() bool
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
