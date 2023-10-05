package gnoland

import (
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

// NewTestingApp returns an in-memory initialized gno.land app.
// It returns a *std.BaseApp which implements abci.Application.
func NewTestingApp() *sdk.BaseApp {
	var (
		db                    = db.NewMemDB()
		skipFailingGenesisTxs = false
		logger                = log.TestingLogger()
		maxCycles             = int64(10000)
	)
	app, err := NewApp(db, skipFailingGenesisTxs, logger, maxCycles)
	if err != nil {
		panic(err)
	}
	return app.(*sdk.BaseApp)
}
