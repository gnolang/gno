package sdk

import abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"

// InitChainer initializes application state at genesis
type InitChainer func(ctx Context, req abci.RequestInitChain) abci.ResponseInitChain

// BeginBlocker runs code before the transactions in a block
//
// Note: applications which set create_empty_blocks=false will not have regular block timing and should use
// e.g. BFT timestamps rather than block height for any periodic BeginBlock logic
type BeginBlocker func(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

// EndBlocker runs code after the transactions in a block and return updates to the validator set
//
// Note: applications which set create_empty_blocks=false will not have regular block timing and should use
// e.g. BFT timestamps rather than block height for any periodic EndBlock logic
type EndBlocker func(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock

// BeginTxHook is a BaseApp-specific hook, called to modify the context with any
// additional application-specific information, before running the messages in a
// transaction.
type BeginTxHook func(ctx Context) Context

// EndTxHook is a BaseApp-specific hook, called after all the messages in a
// transaction have terminated, and only when the tx SUCCEEDED — a failed tx
// reverts all of its state and runs no settlement. Returning a non-nil error
// fails the tx: its writes are reverted and the error surfaces as a typed ABCI
// error instead of a panic.
//
// It runs in BOTH RunTxModeDeliver and RunTxModeCheckExecute. Running it at
// mempool admission is what keeps a tx that cannot possibly settle (e.g. an
// insolvent sponsor, or a SponsorStorage tx that grew storage without calling
// PayStorage) out of the mempool: such a tx is deterministically doomed, and
// admitting it would let anyone burn block gas for free, since a failed
// sponsored tx charges nobody. Under CheckExecute every write the hook makes
// lands in a cache that is discarded, so settlement there is a dry run.
//
// Because of that, the hook MUST NOT write outside the cache-wrapped store —
// in particular it must not commit an app-side transaction store. Do that in
// CommitTxHook, which runs only when the tx is actually being committed.
type EndTxHook func(ctx Context, result Result) error

// CommitTxHook is a BaseApp-specific hook, called only for a successful
// RunTxModeDeliver tx, after EndTxHook and immediately before the tx's writes
// are flushed. It is the place to commit app-side transaction stores that live
// outside the cache-wrapped MultiStore and therefore would NOT be rolled back
// by the CheckExecute dry run.
type CommitTxHook func(ctx Context)
