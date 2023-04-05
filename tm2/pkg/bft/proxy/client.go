package proxy

import (
	"sync"

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
	mtx *sync.Mutex
	app abci.Application
}

func NewLocalClientCreator(app abci.Application) ClientCreator {
	return &localClientCreator{
		mtx: new(sync.Mutex),
		app: app,
	}
}

func (l *localClientCreator) NewABCIClient() (abcicli.Client, error) {
	return abcicli.NewLocalClient(l.mtx, l.app), nil
}

//-----------------------------------------------------------------
// DefaultClientCreator

// Returns the local application, or constructs a new one via proxy.
// This function is meant to work with config fields.
func DefaultClientCreator(local abci.Application, proxy string, transport, dbDir string) ClientCreator {
	if local != nil {
		// local applications (ignores other arguments)
		return NewLocalClientCreator(local)
	} else {
		switch proxy {
		// default mock applications
		case "mock://counter":
			return NewLocalClientCreator(counter.NewCounterApplication(false))
		case "mock://counter_serial":
			return NewLocalClientCreator(counter.NewCounterApplication(true))
		case "mock://kvstore":
			return NewLocalClientCreator(kvstore.NewKVStoreApplication())
		case "mock://persistent_kvstore":
			return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
		case "mock://noop":
			return NewLocalClientCreator(abci.NewBaseApplication())
		default:
			// socket transport applications
			panic("proxy scheme not yet supported: " + proxy)
		}
	}
}
