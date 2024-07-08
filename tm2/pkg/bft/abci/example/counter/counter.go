package counter

import (
	"encoding/binary"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/errors"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

type CounterApplication struct {
	abci.BaseApplication

	hashCount int
	txCount   int
	serial    bool
}

func NewCounterApplication(serial bool) *CounterApplication {
	return &CounterApplication{serial: serial}
}

func (app *CounterApplication) Info(req abci.RequestInfo) abci.ResponseInfo {
	return abci.ResponseInfo{ResponseBase: abci.ResponseBase{
		Data: []byte(fmt.Sprintf("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)),
	}}
}

func (app *CounterApplication) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	key, value := req.Key, req.Value
	if key == "serial" && value == "on" {
		app.serial = true
	} else {
		/*
			TODO Panic and have the ABCI server pass an exception.
			The client can call SetOptionSync() and get an `error`.
			return abci.ResponseSetOption{
				Error: fmt.Sprintf("Unknown key (%s) or value (%s)", key, value),
			}
		*/
		return abci.ResponseSetOption{}
	}

	return abci.ResponseSetOption{}
}

func (app *CounterApplication) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{
					Error: errors.EncodingError{},
					Log:   fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx)),
				},
			}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue != uint64(app.txCount) {
			return abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{
					Error: errors.BadNonceError{},
					Log:   fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.txCount, txValue),
				},
			}
		}
	}
	app.txCount++
	return abci.ResponseDeliverTx{}
}

func (app *CounterApplication) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	if app.serial {
		if len(req.Tx) > 8 {
			return abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Error: errors.EncodingError{},
					Log:   fmt.Sprintf("Max tx size is 8 bytes, got %d", len(req.Tx)),
				},
			}
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(req.Tx):], req.Tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Error: errors.BadNonceError{},
					Log:   fmt.Sprintf("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue),
				},
			}
		}
	}
	return abci.ResponseCheckTx{}
}

func (app *CounterApplication) Commit() (resp abci.ResponseCommit) {
	app.hashCount++
	if app.txCount == 0 {
		return abci.ResponseCommit{}
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return abci.ResponseCommit{ResponseBase: abci.ResponseBase{Data: hash}}
}

func (app *CounterApplication) Query(reqQuery abci.RequestQuery) abci.ResponseQuery {
	switch reqQuery.Path {
	case "hash":
		return abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.hashCount))}
	case "tx":
		return abci.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.txCount))}
	default:
		return abci.ResponseQuery{ResponseBase: abci.ResponseBase{Log: fmt.Sprintf("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}}
	}
}
