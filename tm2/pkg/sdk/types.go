package sdk

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Router provides handlers for each transaction type.
type Router interface {
	AddRoute(r string, h Handler) Router
	Route(path string) Handler
}

// A Handler handles processing messages and answering queries
// for a particular application concern.
type Handler interface {
	// Process defines the core of the state transition function of an application.
	Process(ctx Context, msg Msg) Result
	// Query allows the state to be queried.
	Query(ctx Context, req abci.RequestQuery) abci.ResponseQuery
}

// PayGasInfo tracks whether a realm has called PayGas in the current transaction.
type PayGasInfo struct {
	RealmPkgPath string         // pkg path of the realm that called PayGas
	RealmAddr    crypto.Address // derived address of the realm
	MaxFee       int64          // gas fee cap in ugnot (0 = PayGas not called)
}

type PayStorageInfo struct {
	RealmPkgPath     string           // pkg path of the realm that called PayStorage
	RealmAddr        crypto.Address   // derived address of the realm
	MaxDeposit       int64            // storage deposit cap in ugnot (0 = PayStorage not called)
	AccumulatedDiffs map[string]int64 // tx-level storage diff accumulator (when SponsorStorage=true)
}

// Result is the union of ResponseDeliverTx and ResponseCheckTx plus events.
type Result struct {
	abci.ResponseBase
	GasWanted  int64
	GasUsed    int64
	PayGasInfo *PayGasInfo // set when a realm calls PayGas during execution
}

// AnteHandler authenticates transactions, before their internal messages are handled.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, result Result, abort bool)

// Exports from std.
type Msg = std.Msg

type (
	Tx       = std.Tx
	Coin     = std.Coin
	Coins    = std.Coins
	GasPrice = std.GasPrice
)

var (
	ParseGasPrice  = std.ParseGasPrice
	ParseGasPrices = std.ParseGasPrices
)

//----------------------------------------

// Enum mode for app.runTx
type RunTxMode uint8

const (
	// Check a transaction
	RunTxModeCheck RunTxMode = iota
	// Simulate a transaction
	RunTxModeSimulate RunTxMode = iota
	// Deliver a transaction
	RunTxModeDeliver RunTxMode = iota
)
