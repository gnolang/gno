package proxy

import (
	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

//----------------------------------------------------------------------------------------
// Enforce which abci msgs can be sent on a connection at the type level

type AppConnConsensus interface {
	SetResponseCallback(abcicli.Callback)
	Error() error

	InitChainSync(abci.RequestInitChain) (abci.ResponseInitChain, error)

	BeginBlockSync(abci.RequestBeginBlock) (abci.ResponseBeginBlock, error)
	DeliverTxAsync(abci.RequestDeliverTx) *abcicli.ReqRes
	EndBlockSync(abci.RequestEndBlock) (abci.ResponseEndBlock, error)
	CommitSync() (abci.ResponseCommit, error)
}

type AppConnMempool interface {
	SetResponseCallback(abcicli.Callback)
	Error() error

	CheckTxAsync(abci.RequestCheckTx) *abcicli.ReqRes

	FlushAsync() *abcicli.ReqRes
	FlushSync() error
}

type AppConnQuery interface {
	Error() error

	EchoSync(string) (abci.ResponseEcho, error)
	InfoSync(abci.RequestInfo) (abci.ResponseInfo, error)
	QuerySync(abci.RequestQuery) (abci.ResponseQuery, error)

	//	SetOptionSync(key string, value string) (res abci.Result)
}

//-----------------------------------------------------------------------------------------
// Implements AppConnConsensus (subset of abcicli.Client)

type appConnConsensus struct {
	appConn abcicli.Client
}

func NewAppConnConsensus(appConn abcicli.Client) *appConnConsensus {
	return &appConnConsensus{
		appConn: appConn,
	}
}

func (app *appConnConsensus) SetResponseCallback(cb abcicli.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *appConnConsensus) Error() error {
	return app.appConn.Error()
}

func (app *appConnConsensus) InitChainSync(req abci.RequestInitChain) (abci.ResponseInitChain, error) {
	return app.appConn.InitChainSync(req)
}

func (app *appConnConsensus) BeginBlockSync(req abci.RequestBeginBlock) (abci.ResponseBeginBlock, error) {
	return app.appConn.BeginBlockSync(req)
}

func (app *appConnConsensus) DeliverTxAsync(req abci.RequestDeliverTx) *abcicli.ReqRes {
	return app.appConn.DeliverTxAsync(req)
}

func (app *appConnConsensus) EndBlockSync(req abci.RequestEndBlock) (abci.ResponseEndBlock, error) {
	return app.appConn.EndBlockSync(req)
}

func (app *appConnConsensus) CommitSync() (abci.ResponseCommit, error) {
	return app.appConn.CommitSync()
}

//------------------------------------------------
// Implements AppConnMempool (subset of abcicli.Client)

type appConnMempool struct {
	appConn abcicli.Client
}

func NewAppConnMempool(appConn abcicli.Client) *appConnMempool {
	return &appConnMempool{
		appConn: appConn,
	}
}

func (app *appConnMempool) SetResponseCallback(cb abcicli.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *appConnMempool) Error() error {
	return app.appConn.Error()
}

func (app *appConnMempool) FlushAsync() *abcicli.ReqRes {
	return app.appConn.FlushAsync()
}

func (app *appConnMempool) FlushSync() error {
	return app.appConn.FlushSync()
}

func (app *appConnMempool) CheckTxAsync(req abci.RequestCheckTx) *abcicli.ReqRes {
	return app.appConn.CheckTxAsync(req)
}

//------------------------------------------------
// Implements AppConnQuery (subset of abcicli.Client)

type appConnQuery struct {
	appConn abcicli.Client
}

func NewAppConnQuery(appConn abcicli.Client) *appConnQuery {
	return &appConnQuery{
		appConn: appConn,
	}
}

func (app *appConnQuery) Error() error {
	return app.appConn.Error()
}

func (app *appConnQuery) EchoSync(msg string) (abci.ResponseEcho, error) {
	return app.appConn.EchoSync(msg)
}

func (app *appConnQuery) InfoSync(req abci.RequestInfo) (abci.ResponseInfo, error) {
	return app.appConn.InfoSync(req)
}

func (app *appConnQuery) QuerySync(reqQuery abci.RequestQuery) (abci.ResponseQuery, error) {
	return app.appConn.QuerySync(reqQuery)
}
