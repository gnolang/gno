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
// transaction have terminated. It is invoked once per DeliverTx:
//   - committed == true  on the success path (before the tx's writes are
//     flushed): the hook may perform end-of-tx settlement AND commit any
//     app-side transaction store.
//   - committed == false on the failure path (after msg writes have been
//     reverted): the hook may still settle obligations that survive failure
//     (e.g. charge a gas sponsor for gas consumed), but must NOT commit
//     state tied to the reverted messages.
//
// Returning a non-nil error on the success path fails the tx (its writes are
// reverted and the error surfaces as a typed ABCI error instead of a panic).
// An error returned on the failure path is logged; the tx already failed.
type EndTxHook func(ctx Context, result Result, committed bool) error
