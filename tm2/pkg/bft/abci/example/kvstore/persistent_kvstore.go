package kvstore

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/errors"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const (
	ValidatorUpdatePrefix string = "val:"
	ValidatorKeyPrefix    string = "/val/"
)

const dbBackend = db.PebbleDBBackend

// -----------------------------------------

var _ abci.Application = (*PersistentKVStoreApplication)(nil)

type PersistentKVStoreApplication struct {
	app *KVStoreApplication

	// validator set
	ValSetChanges []abci.ValidatorUpdate

	logger *slog.Logger
}

func NewPersistentKVStoreApplication(dbDir string) *PersistentKVStoreApplication {
	name := "kvstore"
	db, err := db.NewDB(name, dbBackend, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &PersistentKVStoreApplication{
		app:    &KVStoreApplication{state: state},
		logger: log.NewNoopLogger(),
	}
}

func (app *PersistentKVStoreApplication) SetLogger(l *slog.Logger) {
	app.logger = l
}

func (app *PersistentKVStoreApplication) Info(req abci.RequestInfo) abci.ResponseInfo {
	res := app.app.Info(req)
	res.LastBlockHeight = app.app.state.Height
	res.LastBlockAppHash = app.app.state.AppHash
	return res
}

func (app *PersistentKVStoreApplication) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	return app.app.SetOption(req)
}

// tx is either "val:pubkey!power" or "key=value" or just arbitrary bytes
func (app *PersistentKVStoreApplication) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	// if it starts with "val:", update the validator set
	// format is "val:pubkey!power"
	if isValidatorTx(req.Tx) {
		// update validators in the merkle tree
		// and in app.ValSetChanges
		return app.execValidatorTx(req.Tx)
	}

	// otherwise, update the key-value store
	return app.app.DeliverTx(req)
}

func (app *PersistentKVStoreApplication) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return app.app.CheckTx(req)
}

// Commit will panic if InitChain was not called
func (app *PersistentKVStoreApplication) Commit() abci.ResponseCommit {
	return app.app.Commit()
}

// When path=/val and data={validator address}, returns the validator update (abci.ValidatorUpdate) varint encoded.
// For any other path, returns an associated value or nil if missing.
func (app *PersistentKVStoreApplication) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	switch reqQuery.Path {
	case "/val":
		key := []byte(ValidatorUpdatePrefix + string(reqQuery.Data))
		value, err := app.app.state.db.Get(key)
		if err != nil {
			panic(err)
		}

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	default:
		return app.app.Query(reqQuery)
	}
}

// Save the validators in the merkle tree
func (app *PersistentKVStoreApplication) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return abci.ResponseInitChain{}
}

// Track the block hash and header information
func (app *PersistentKVStoreApplication) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	// reset valset changes
	app.ValSetChanges = make([]abci.ValidatorUpdate, 0)

	/* REMOVE
	for _, vio := range req.Violations {
		if _, ok := vio.Evidence.(*tmtypes.DuplicateVoteEvidence); ok {
			for _, val := range vio.Validators {
				// decrease voting power of each by 1
				if val.Power == 0 {
					continue
				}
				app.updateValidator(abci.ValidatorUpdate{
					Address: val.PubKey.Address(),
					PubKey:  val.PubKey,
					Power:   val.Power - 1,
				})
			}
		}
	}
	*/
	return abci.ResponseBeginBlock{}
}

// Update the validator set
func (app *PersistentKVStoreApplication) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{ValidatorUpdates: app.ValSetChanges}
}

// ---------------------------------------------
// update validators

func (app *PersistentKVStoreApplication) Validators() (validators []abci.ValidatorUpdate) {
	itr, err := app.app.state.db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		if isValidatorKey(itr.Key()) {
			validator := new(abci.ValidatorUpdate)
			amino.MustUnmarshal(itr.Value(), validator)
			validators = append(validators, *validator)
		}
	}
	return
}

func makeValidatorKey(val abci.ValidatorUpdate) []byte {
	return fmt.Appendf(nil, "%s%X", ValidatorKeyPrefix, val.PubKey.Address())
}

func isValidatorKey(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorKeyPrefix)
}

func MakeValSetChangeTx(pubkey crypto.PubKey, power int64) []byte {
	pubkeyS := base64.StdEncoding.EncodeToString(pubkey.Bytes())
	return fmt.Appendf(nil, "%s%s!%d", ValidatorUpdatePrefix, pubkeyS, power)
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorUpdatePrefix)
}

// format is "val:pubkey!power"
// pubkey is a base64-encoded 32-byte ed25519 key
func (app *PersistentKVStoreApplication) execValidatorTx(tx []byte) (res abci.ResponseDeliverTx) {
	tx = tx[len(ValidatorUpdatePrefix):]

	// get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "!")
	if len(pubKeyAndPower) != 2 {
		res.Error = errors.EncodingError{}
		res.Log = fmt.Sprintf("Expected 'pubkey!power'. Got %v", pubKeyAndPower)
		return
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey
	bz, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		res.Error = errors.EncodingError{}
		res.Log = fmt.Sprintf("Pubkey (%s) is invalid base64", pubkeyS)
		return
	}
	var pubkey crypto.PubKey
	amino.MustUnmarshal(bz, &pubkey)

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		res.Error = errors.EncodingError{}
		res.Log = fmt.Sprintf("Power (%s) is not an int", powerS)
		return
	}

	// update
	return app.updateValidator(abci.ValidatorUpdate{Address: pubkey.Address(), PubKey: pubkey, Power: power})
}

// add, update, or remove a validator
func (app *PersistentKVStoreApplication) updateValidator(val abci.ValidatorUpdate) (res abci.ResponseDeliverTx) {
	if val.Power == 0 {
		// remove validator
		found, err := app.app.state.db.Has(makeValidatorKey(val))
		if err != nil {
			panic(err)
		}
		if !found {
			res.Error = errors.UnauthorizedError{}
			res.Log = fmt.Sprintf("Cannot remove non-existent validator %s", val.PubKey.String())
			return res
		}
		app.app.state.db.Delete(makeValidatorKey(val))
	} else {
		// add or update validator
		bz := amino.MustMarshal(val)
		app.app.state.db.Set(makeValidatorKey(val), bz)
	}

	// we only update the changes array if we successfully updated the tree
	app.ValSetChanges = append(app.ValSetChanges, val)

	return abci.ResponseDeliverTx{}
}

func (app *PersistentKVStoreApplication) Close() error {
	return app.app.Close()
}
