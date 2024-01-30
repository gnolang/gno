package appconn

import (
	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/service"
)

//-----------------------------

// Tendermint's interface to the application consists of multiple connections
type AppConns interface {
	service.Service

	Mempool() Mempool
	Consensus() Consensus
	Query() Query
}

// NewABCIClient returns newly connected client
type ClientCreator interface {
	NewABCIClient() (abcicli.Client, error)
}

func NewAppConns(clientCreator ClientCreator) AppConns {
	return NewMulti(clientCreator)
}

//-----------------------------
// multi implements AppConns

// a multi is made of a few appConns (mempool, consensus, query)
// and manages their underlying abci clients
// TODO: on app restart, clients must reboot together
type multi struct {
	service.BaseService

	mempoolConn   *mempool
	consensusConn *consensus
	queryConn     *query

	clientCreator ClientCreator
}

// Make all necessary abci connections to the application
func NewMulti(clientCreator ClientCreator) *multi {
	multi := &multi{
		clientCreator: clientCreator,
	}
	multi.BaseService = *service.NewBaseService(nil, "multi", multi)
	return multi
}

// Returns the mempool connection
func (app *multi) Mempool() Mempool {
	return app.mempoolConn
}

// Returns the consensus Connection
func (app *multi) Consensus() Consensus {
	return app.consensusConn
}

// Returns the query Connection
func (app *multi) Query() Query {
	return app.queryConn
}

func (app *multi) OnStart() error {
	// query connection
	querycli, err := app.clientCreator.NewABCIClient()
	if err != nil {
		return errors.Wrap(err, "Error creating ABCI client (query connection)")
	}
	querycli.SetLogger(app.Logger.With("module", "abci-client", "connection", "query"))
	if err := querycli.Start(); err != nil {
		return errors.Wrap(err, "Error starting ABCI client (query connection)")
	}
	app.queryConn = NewQuery(querycli)

	// mempool connection
	memcli, err := app.clientCreator.NewABCIClient()
	if err != nil {
		return errors.Wrap(err, "Error creating ABCI client (mempool connection)")
	}
	memcli.SetLogger(app.Logger.With("module", "abci-client", "connection", "mempool"))
	if err := memcli.Start(); err != nil {
		return errors.Wrap(err, "Error starting ABCI client (mempool connection)")
	}
	app.mempoolConn = NewMempool(memcli)

	// consensus connection
	concli, err := app.clientCreator.NewABCIClient()
	if err != nil {
		return errors.Wrap(err, "Error creating ABCI client (consensus connection)")
	}
	concli.SetLogger(app.Logger.With("module", "abci-client", "connection", "consensus"))
	if err := concli.Start(); err != nil {
		return errors.Wrap(err, "Error starting ABCI client (consensus connection)")
	}
	app.consensusConn = NewConsensus(concli)

	return nil
}
