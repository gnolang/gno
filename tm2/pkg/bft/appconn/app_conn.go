// Package appconn manages the connection of Tendermint to the application layer.
package appconn

import (
	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

//----------------------------------------------------------------------------------------
// Enforce which abci msgs can be sent on a connection at the type level

type Consensus interface {
	SetResponseCallback(abcicli.Callback)
	Error() error

	InitChainSync(abci.RequestInitChain) (abci.ResponseInitChain, error)

	BeginBlockSync(abci.RequestBeginBlock) (abci.ResponseBeginBlock, error)
	DeliverTxAsync(abci.RequestDeliverTx) *abcicli.ReqRes
	EndBlockSync(abci.RequestEndBlock) (abci.ResponseEndBlock, error)
	CommitSync() (abci.ResponseCommit, error)
}

type Mempool interface {
	SetResponseCallback(abcicli.Callback)
	Error() error

	CheckTxAsync(abci.RequestCheckTx) *abcicli.ReqRes

	FlushAsync() *abcicli.ReqRes
	FlushSync() error

	QuerySync(abci.RequestQuery) (abci.ResponseQuery, error)
}

type Query interface {
	Error() error

	EchoSync(string) (abci.ResponseEcho, error)
	InfoSync(abci.RequestInfo) (abci.ResponseInfo, error)
	QuerySync(abci.RequestQuery) (abci.ResponseQuery, error)

	//	SetOptionSync(key string, value string) (res abci.Result)
}

//-----------------------------------------------------------------------------------------
// Implements Consensus (subset of abcicli.Client)

type consensus struct {
	appConn abcicli.Client
}

func NewConsensus(appConn abcicli.Client) *consensus {
	return &consensus{
		appConn: appConn,
	}
}

func (app *consensus) SetResponseCallback(cb abcicli.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *consensus) Error() error {
	return app.appConn.Error()
}

func (app *consensus) InitChainSync(req abci.RequestInitChain) (abci.ResponseInitChain, error) {
	return app.appConn.InitChainSync(req)
}

func (app *consensus) BeginBlockSync(req abci.RequestBeginBlock) (abci.ResponseBeginBlock, error) {
	return app.appConn.BeginBlockSync(req)
}

func (app *consensus) DeliverTxAsync(req abci.RequestDeliverTx) *abcicli.ReqRes {
	return app.appConn.DeliverTxAsync(req)
}

func (app *consensus) EndBlockSync(req abci.RequestEndBlock) (abci.ResponseEndBlock, error) {
	return app.appConn.EndBlockSync(req)
}

func (app *consensus) CommitSync() (abci.ResponseCommit, error) {
	return app.appConn.CommitSync()
}

//------------------------------------------------
// Implements Mempool (subset of abcicli.Client)

type mempool struct {
	appConn abcicli.Client
}

func NewMempool(appConn abcicli.Client) *mempool {
	return &mempool{
		appConn: appConn,
	}
}

func (app *mempool) SetResponseCallback(cb abcicli.Callback) {
	app.appConn.SetResponseCallback(cb)
}

func (app *mempool) Error() error {
	return app.appConn.Error()
}

func (app *mempool) FlushAsync() *abcicli.ReqRes {
	return app.appConn.FlushAsync()
}

func (app *mempool) FlushSync() error {
	return app.appConn.FlushSync()
}

func (app *mempool) CheckTxAsync(req abci.RequestCheckTx) *abcicli.ReqRes {
	return app.appConn.CheckTxAsync(req)
}

//------------------------------------------------
// Implements Query (subset of abcicli.Client)

type query struct {
	appConn abcicli.Client
}

func NewQuery(appConn abcicli.Client) *query {
	return &query{
		appConn: appConn,
	}
}

func (app *query) Error() error {
	return app.appConn.Error()
}

func (app *query) EchoSync(msg string) (abci.ResponseEcho, error) {
	return app.appConn.EchoSync(msg)
}

func (app *query) InfoSync(req abci.RequestInfo) (abci.ResponseInfo, error) {
	return app.appConn.InfoSync(req)
}

func (app *query) QuerySync(reqQuery abci.RequestQuery) (abci.ResponseQuery, error) {
	return app.appConn.QuerySync(reqQuery)
}

func (app *mempool) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	return app.appConn.QuerySync(req)
}
