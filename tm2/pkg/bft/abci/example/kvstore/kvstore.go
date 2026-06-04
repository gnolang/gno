package kvstore

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	abciver "github.com/gnolang/gno/tm2/pkg/bft/abci/version"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

var (
	stateKey        = []byte("stateKey")
	kvPairPrefixKey = []byte("kvPairKey:")
	AppVersion      = "v0.0.0"
)

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

// ---------------------------------------------------

var _ abci.Application = (*KVStoreApplication)(nil)

type KVStoreApplication struct {
	abci.BaseApplication

	state State
}

func NewKVStoreApplication() *KVStoreApplication {
	state := loadState(memdb.NewMemDB())
	return &KVStoreApplication{state: state}
}

func (app *KVStoreApplication) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{
		ResponseBase: abci.ResponseBase{
			Data: fmt.Appendf(nil, "{\"size\":%v}", app.state.Size),
		},
		ABCIVersion: abciver.Version,
		AppVersion:  AppVersion,
	}
}

// tx is either "key=value" or just arbitrary bytes
func (app *KVStoreApplication) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {
	var key, value []byte
	parts := bytes.Split(req.Tx, []byte("="))
	if len(parts) == 2 {
		key, value = parts[0], parts[1]
	} else {
		key, value = req.Tx, req.Tx
	}

	app.state.db.Set(prefixKey(key), value)
	app.state.Size += 1

	events := []abci.Event{abci.EventString(`{"creator":"Cosmoshi Netowoko"}`)}

	res.Events = events
	return res
}

func (app *KVStoreApplication) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return abci.ResponseCheckTx{GasWanted: 1}
}

func (app *KVStoreApplication) Commit() (res abci.ResponseCommit) {
	// Using a memdb - just return the big endian size of the db
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height += 1
	saveState(app.state)

	res.Data = appHash
	return res
}

// Returns an associated value or nil if missing.
func (app *KVStoreApplication) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if reqQuery.Prove {
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}
		// resQuery.Index = -1 // TODO make Proof return index
		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	} else {
		resQuery.Key = reqQuery.Data
		value, err := app.state.db.Get(prefixKey(reqQuery.Data))
		if err != nil {
			panic(err)
		}
		resQuery.Value = value
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
}

func (app *KVStoreApplication) Close() error {
	return app.state.db.Close()
}
