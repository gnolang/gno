package proxy

import (
	"sync"
	"time"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/counter"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

// NewABCIClient returns newly connected client
type ClientCreator interface {
	NewABCIClient() (abcicli.Client, error)
}

//----------------------------------------------------
// local proxy uses a mutex on an in-proc app

type localClientCreator struct {
	mtx     *sync.Mutex
	app     abci.Application
	timeout time.Duration
}

func NewLocalClientCreator(app abci.Application) ClientCreator {
	return NewLocalClientCreatorWithTimeout(app, 0)
}

func NewLocalClientCreatorWithTimeout(app abci.Application, timeout time.Duration) ClientCreator {
	return &localClientCreator{
		mtx:     new(sync.Mutex),
		app:     app,
		timeout: timeout,
	}
}

func (l *localClientCreator) NewABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(l.mtx, l.app, l.timeout), nil
}

//-----------------------------------------------------------------
// DefaultClientCreator

// Returns the local application, or constructs a new one via proxy.
// This function is meant to work with config fields.
func DefaultClientCreator(local abci.Application, proxy string, transport, dbDir string, timeout time.Duration) ClientCreator {
	if local != nil {
		// local applications (ignores other arguments)
		return NewLocalClientCreatorWithTimeout(local, timeout)
	} else {
		switch proxy {
		// default mock applications
		case "mock://counter":
			return NewLocalClientCreatorWithTimeout(counter.NewCounterApplication(false), timeout)
		case "mock://counter_serial":
			return NewLocalClientCreatorWithTimeout(counter.NewCounterApplication(true), timeout)
		case "mock://kvstore":
			return NewLocalClientCreatorWithTimeout(kvstore.NewKVStoreApplication(), timeout)
		case "mock://persistent_kvstore":
			return NewLocalClientCreatorWithTimeout(kvstore.NewPersistentKVStoreApplication(dbDir), timeout)
		case "mock://noop":
			return NewLocalClientCreatorWithTimeout(abci.NewBaseApplication(), timeout)
		default:
			// socket transport applications
			panic("proxy scheme not yet supported: " + proxy)
		}
	}
}
