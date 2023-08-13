package gnoland

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/log"
)

func NewTestingApp() abci.Application {
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
	return app
}
