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
type localClient struct {
	service.BaseService

	mtx *sync.Mutex
	abci.Application
	Callback
}

func NewLocalClient(mtx *sync.Mutex, app abci.Application) *localClient {
	if mtx == nil {
		mtx = new(sync.Mutex)
	}
	cli := &localClient{
		mtx:         mtx,
		Application: app,
	}
	cli.BaseService = *service.NewBaseService(nil, "localClient", cli)
	return cli
}

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
