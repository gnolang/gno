package abcicli

import (
	"sync"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// Client defines an interface for an ABCI client.
// All `Async` methods return a `ReqRes` object.
// All `Sync` methods return the appropriate protobuf ResponseXxx struct and an error.
// Note these are client errors, eg. ABCI socket connectivity issues.
// Application-related errors are reflected in response via ABCI error codes and logs.
type Client interface {
	service.Service

	SetResponseCallback(Callback)
	Error() error

	FlushAsync() *ReqRes
	EchoAsync(msg string) *ReqRes
	InfoAsync(abci.RequestInfo) *ReqRes
	SetOptionAsync(abci.RequestSetOption) *ReqRes
	DeliverTxAsync(abci.RequestDeliverTx) *ReqRes
	CheckTxAsync(abci.RequestCheckTx) *ReqRes
	QueryAsync(abci.RequestQuery) *ReqRes
	CommitAsync() *ReqRes
	InitChainAsync(abci.RequestInitChain) *ReqRes
	BeginBlockAsync(abci.RequestBeginBlock) *ReqRes
	EndBlockAsync(abci.RequestEndBlock) *ReqRes

	FlushSync() error
	EchoSync(msg string) (abci.ResponseEcho, error)
	InfoSync(abci.RequestInfo) (abci.ResponseInfo, error)
	SetOptionSync(abci.RequestSetOption) (abci.ResponseSetOption, error)
	DeliverTxSync(abci.RequestDeliverTx) (abci.ResponseDeliverTx, error)
	CheckTxSync(abci.RequestCheckTx) (abci.ResponseCheckTx, error)
	QuerySync(abci.RequestQuery) (abci.ResponseQuery, error)
	CommitSync() (abci.ResponseCommit, error)
	InitChainSync(abci.RequestInitChain) (abci.ResponseInitChain, error)
	BeginBlockSync(abci.RequestBeginBlock) (abci.ResponseBeginBlock, error)
	EndBlockSync(abci.RequestEndBlock) (abci.ResponseEndBlock, error)
}

// ----------------------------------------

type Callback func(abci.Request, abci.Response)

// ----------------------------------------

type ReqRes struct {
	abci.Request
	wg *sync.WaitGroup

	mtx sync.Mutex
	abci.Response
	cb func(abci.Response) // A single callback that may be set.
}

func NewReqRes(req abci.Request) *ReqRes {
	return &ReqRes{
		Request:  req,
		wg:       waitGroup1(),
		Response: nil,

		cb: nil,
	}
}

// Sets the callback for this ReqRes atomically.
// If reqRes is already done, calls cb immediately.
// NOTE: reqRes.cb should not change if reqRes.done.
// NOTE: only one callback is supported.
func (reqRes *ReqRes) SetCallback(cb func(res abci.Response)) {
	reqRes.mtx.Lock()

	if reqRes.Response != nil {
		reqRes.mtx.Unlock()
		cb(reqRes.Response)
		return
	}

	reqRes.cb = cb
	reqRes.mtx.Unlock()
}

func (reqRes *ReqRes) GetCallback() func(abci.Response) {
	reqRes.mtx.Lock()
	defer reqRes.mtx.Unlock()
	return reqRes.cb
}

// Wait will wait until SetResponse() is called.
func (reqRes *ReqRes) Wait() {
	reqRes.wg.Wait()
}

func (reqRes *ReqRes) SetResponse(res abci.Response) {
	reqRes.mtx.Lock()
	if reqRes.Response != nil {
		panic("should not happen")
	}
	reqRes.Response = res
	reqRes.mtx.Unlock()

	reqRes.wg.Done()
}

// NOTE: it should be safe to read reqRes.cb without locks after this.
func (reqRes *ReqRes) Done() {
	// Finally, release the hounds.
	reqRes.wg.Done()
}

func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return
}
