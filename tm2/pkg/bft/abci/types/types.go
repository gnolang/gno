package abci

import (
	"encoding/json"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
)

// ----------------------------------------
// Request types

type Request interface {
	AssertRequest()
}

type RequestBase struct{}

func (RequestBase) AssertRequest() {}

type RequestEcho struct {
	RequestBase
	Message string
}

type RequestFlush struct {
	RequestBase
}

type RequestInfo struct {
	RequestBase
}

// nondeterministic
type RequestSetOption struct {
	RequestBase
	Key   string
	Value string
}

type RequestInitChain struct {
	RequestBase
	Time            time.Time
	ChainID         string
	ConsensusParams *ConsensusParams
	Validators      []ValidatorUpdate
	AppState        interface{}
}

type RequestQuery struct {
	RequestBase
	Data   []byte
	Path   string
	Height int64
	Prove  bool
}

type RequestBeginBlock struct {
	RequestBase
	Hash           []byte
	Header         Header
	LastCommitInfo *LastCommitInfo
	// Violations     []Violation
}

type CheckTxType int

const (
	CheckTxTypeNew     CheckTxType = 0
	CheckTxTypeRecheck             = iota
)

type RequestCheckTx struct {
	RequestBase
	Tx   []byte
	Type CheckTxType
}

type RequestDeliverTx struct {
	RequestBase
	Tx []byte
}

type RequestEndBlock struct {
	RequestBase
	Height int64
}

type RequestCommit struct {
	RequestBase
}

// ----------------------------------------
// Response types

type Response interface {
	AssertResponse()
}

type ResponseBase struct {
	Error  Error
	Data   []byte
	Events []Event

	Log  string // nondeterministic
	Info string // nondeterministic
}

func (ResponseBase) AssertResponse() {}

func (r ResponseBase) IsOK() bool {
	return r.Error == nil
}

func (r ResponseBase) IsErr() bool {
	return r.Error != nil
}

func (r ResponseBase) EncodeEvents() []byte {
	if len(r.Events) == 0 {
		return []byte("[]")
	}
	res, err := json.Marshal(r.Events)
	if err != nil {
		panic(err)
	}
	return res
}

// nondeterministic
type ResponseException struct {
	ResponseBase
}

type ResponseEcho struct {
	ResponseBase
	Message string
}

type ResponseFlush struct {
	ResponseBase
}

type ResponseInfo struct {
	ResponseBase
	ABCIVersion      string
	AppVersion       string
	LastBlockHeight  int64
	LastBlockAppHash []byte
}

// nondeterministic
type ResponseSetOption struct {
	ResponseBase
}

type ResponseInitChain struct {
	ResponseBase
	ConsensusParams *ConsensusParams
	Validators      []ValidatorUpdate
}

type ResponseQuery struct {
	ResponseBase
	Key    []byte
	Value  []byte
	Proof  *merkle.Proof
	Height int64
}

type ResponseBeginBlock struct {
	ResponseBase
}

type ResponseCheckTx struct {
	ResponseBase
	GasWanted int64 // nondeterministic
	GasUsed   int64
}

type ResponseDeliverTx struct {
	ResponseBase
	GasWanted int64
	GasUsed   int64
}

type ResponseEndBlock struct {
	ResponseBase
	ValidatorUpdates []ValidatorUpdate
	ConsensusParams  *ConsensusParams
	Events           []Event
}

type ResponseCommit struct {
	ResponseBase
}

// ----------------------------------------
// Interface types

type Error interface {
	AssertABCIError()
	Error() string
}

type Event interface {
	AssertABCIEvent()
}

type Header interface {
	GetChainID() string
	GetHeight() int64
	GetTime() time.Time
	AssertABCIHeader()
}

// ----------------------------------------
// Error types

type StringError string

func (StringError) AssertABCIError() {}

func (err StringError) Error() string {
	return string(err)
}

// ----------------------------------------
// Event types

type EventString string

func (EventString) AssertABCIEvent() {}

func (err EventString) Event() string {
	return string(err)
}

// ----------------------------------------
// Misc

// Parameters that need to be negotiated between the app and consensus.
type ConsensusParams struct {
	Block     *BlockParams
	Validator *ValidatorParams
}

type BlockParams struct {
	MaxTxBytes    int64 // must be > 0
	MaxDataBytes  int64 // must be > 0
	MaxBlockBytes int64 // must be > 0
	MaxGas        int64 // must be >= -1
	TimeIotaMS    int64 // must be > 0
}

type ValidatorParams struct {
	PubKeyTypeURLs []string
}

type ValidatorUpdate struct {
	Address crypto.Address
	PubKey  crypto.PubKey
	Power   int64
}

type LastCommitInfo struct {
	Round int32
	Votes []VoteInfo
}

// unstable
type VoteInfo struct {
	Address         crypto.Address
	Power           int64
	SignedLastBlock bool
}

/*
// unstable
type Validator struct {
	Address crypto.Address
	PubKey  crypto.PubKey
	Power   int64
}

// unstable
type Violation struct {
	Evidence
	Validators       []Validator
	Height           int64
	Time             time.Time
	TotalVotingPower int64
}
*/
