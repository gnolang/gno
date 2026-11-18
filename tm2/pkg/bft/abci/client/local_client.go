package abcicli

import (
	"sync"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

var _ Client = (*localClient)(nil)

// NOTE: use defer to unlock mutex because Application might panic (e.g., in
// case of malicious tx or query). It only makes sense for publicly exposed
// methods like CheckTx (/broadcast_tx_* RPC endpoint) or Query (/abci_query
// RPC endpoint), but defers are used everywhere for the sake of consistency.
// The defers are also what makes a non-mutex Locker safe here: a limiter that
// hands out slots gets every slot back even when Application panics.
type localClient struct {
	service.BaseService

	// mtx is held for the duration of every call into Application, so it is
	// what sets this connection's concurrency. A sync.Mutex admits one caller
	// at a time; any other Locker admits whatever it chooses to admit, and the
	// caller supplying it is asserting that everything reachable from the
	// methods this client will serve is safe at that concurrency. See
	// proxy.NewReadOnlyABCIClient for the one connection that supplies
	// something else.
	mtx sync.Locker
	abci.Application
	Callback
}

// NewLocalClient returns a client that calls app in-process, holding mtx for
// the duration of every call. mtx must not be nil.
func NewLocalClient(mtx sync.Locker, app abci.Application) *localClient {
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

// SetResponseCallback sets the callback completeRequest invokes on every Async
// method. The lock below is only as strong as mtx: it excludes the concurrent
// completeRequest reads when mtx is a mutex, and not when mtx admits several
// callers at once. A client built on such a Locker has to reject this method
// rather than rely on the lock here — see proxy.readOnlyClient.
func (app *localClient) SetResponseCallback(cb Callback) {
	app.mtx.Lock()
	app.Callback = cb
	app.mtx.Unlock()
}

// TODO: change abci.Application to include Error()?
func (app *localClient) Error() error {
	return nil
}

func (app *localClient) FlushAsync() *ReqRes {
	// Do nothing
	return newLocalReqRes(abci.RequestFlush{}, nil)
}

func (app *localClient) EchoAsync(msg string) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	return app.completeRequest(
		abci.RequestEcho{Message: msg},
		abci.ResponseEcho{Message: msg},
	)
}

func (app *localClient) InfoAsync(req abci.RequestInfo) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return app.completeRequest(req, res)
}

func (app *localClient) SetOptionAsync(req abci.RequestSetOption) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return app.completeRequest(req, res)
}

func (app *localClient) DeliverTxAsync(req abci.RequestDeliverTx) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(req)
	return app.completeRequest(req, res)
}

func (app *localClient) CheckTxAsync(req abci.RequestCheckTx) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return app.completeRequest(req, res)
}

func (app *localClient) QueryAsync(req abci.RequestQuery) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return app.completeRequest(req, res)
}

func (app *localClient) CommitAsync() *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return app.completeRequest(abci.RequestCommit{}, res)
}

func (app *localClient) InitChainAsync(req abci.RequestInitChain) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return app.completeRequest(req, res)
}

func (app *localClient) BeginBlockAsync(req abci.RequestBeginBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return app.completeRequest(req, res)
}

func (app *localClient) EndBlockAsync(req abci.RequestEndBlock) *ReqRes {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return app.completeRequest(req, res)
}

//-------------------------------------------------------

func (app *localClient) FlushSync() error {
	return nil
}

func (app *localClient) EchoSync(msg string) (abci.ResponseEcho, error) {
	return abci.ResponseEcho{Message: msg}, nil
}

func (app *localClient) InfoSync(req abci.RequestInfo) (abci.ResponseInfo, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Info(req)
	return res, nil
}

func (app *localClient) SetOptionSync(req abci.RequestSetOption) (abci.ResponseSetOption, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.SetOption(req)
	return res, nil
}

func (app *localClient) DeliverTxSync(req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.DeliverTx(req)
	return res, nil
}

func (app *localClient) CheckTxSync(req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.CheckTx(req)
	return res, nil
}

func (app *localClient) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Query(req)
	return res, nil
}

func (app *localClient) CommitSync() (abci.ResponseCommit, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.Commit()
	return res, nil
}

func (app *localClient) InitChainSync(req abci.RequestInitChain) (abci.ResponseInitChain, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.InitChain(req)
	return res, nil
}

func (app *localClient) BeginBlockSync(req abci.RequestBeginBlock) (abci.ResponseBeginBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.BeginBlock(req)
	return res, nil
}

func (app *localClient) EndBlockSync(req abci.RequestEndBlock) (abci.ResponseEndBlock, error) {
	app.mtx.Lock()
	defer app.mtx.Unlock()

	res := app.Application.EndBlock(req)
	return res, nil
}

//-------------------------------------------------------

func (app *localClient) completeRequest(req abci.Request, res abci.Response) *ReqRes {
	app.Callback(req, res)
	return newLocalReqRes(req, res)
}

func newLocalReqRes(req abci.Request, res abci.Response) *ReqRes {
	reqRes := NewReqRes(req)
	reqRes.SetResponse(res)
	return reqRes
}
