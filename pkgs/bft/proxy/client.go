package proxy

import (
	"sync"

	abcicli "github.com/gnolang/gno/pkgs/bft/abci/client"
	"github.com/gnolang/gno/pkgs/bft/abci/example/counter"
	"github.com/gnolang/gno/pkgs/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
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
// default

func DefaultClientCreator(addr, transport, dbDir string) ClientCreator {
	switch addr {
	case "counter":
		return NewLocalClientCreator(counter.NewCounterApplication(false))
	case "counter_serial":
		return NewLocalClientCreator(counter.NewCounterApplication(true))
	case "kvstore":
		return NewLocalClientCreator(kvstore.NewKVStoreApplication())
	case "persistent_kvstore":
		return NewLocalClientCreator(kvstore.NewPersistentKVStoreApplication(dbDir))
	case "noop":
		return NewLocalClientCreator(abci.NewBaseApplication())
	default:
		panic("unknown client " + addr)
	}
}
