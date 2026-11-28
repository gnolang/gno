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

// TxProfiler runs a tx through Simulate with gas profiling enabled and returns a
// pprof profile of its gas usage plus a status log (e.g. whether the tx
// completed or the profile is partial because the tx failed/ran out of gas).
// A non-nil err means no profile could be produced. Optional and dev-only: nil
// disables the .app/profiletx query. Registered by the application, not tm2, so
// tm2 stays free of any profiler dependency.
type TxProfiler func(txBytes []byte) (profile []byte, log string, err error)

// EndTxHook is a BaseApp-specific hook, called after all the messages in a
// transaction have terminated.
type EndTxHook func(ctx Context, result Result)
