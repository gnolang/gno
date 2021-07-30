package types

import (
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/events"
)

// TMEvent implements events.Event.
type TMEvent interface {
	events.Event
}

func (_ EventNewBlock) AssertEvent()            {}
func (_ EventNewBlockHeader) AssertEvent()      {}
func (_ EventTx) AssertEvent()                  {}
func (_ EventVote) AssertEvent()                {}
func (_ EventString) AssertEvent()              {}
func (_ EventValidatorSetUpdates) AssertEvent() {}

// Most event messages are basic types (a block, a transaction)
// but some (an input to a call tx or a receive) are more exotic

type EventNewBlock struct {
	Block *Block `json:"block"`

	ResultBeginBlock abci.ResponseBeginBlock `json:"result_begin_block"`
	ResultEndBlock   abci.ResponseEndBlock   `json:"result_end_block"`
}

// light weight event for benchmarking
type EventNewBlockHeader struct {
	Header Header `json:"header"`

	ResultBeginBlock abci.ResponseBeginBlock `json:"result_begin_block"`
	ResultEndBlock   abci.ResponseEndBlock   `json:"result_end_block"`
}

// All txs fire EventTx
type EventTx struct {
	Result TxResult `json:"result"`
}

type EventVote struct {
	Vote *Vote `json:"vote"`
}

type EventString string

type EventValidatorSetUpdates struct {
	ValidatorUpdates []abci.ValidatorUpdate `json:"validator_updates"`
}
