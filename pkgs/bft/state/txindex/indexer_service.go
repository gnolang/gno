package txindex

import (
	"github.com/gnolang/gno/pkgs/events"
	"github.com/gnolang/gno/pkgs/service"
)

// IndexerService connects event bus and transaction indexer together in order
// to index transactions coming from event bus.
type IndexerService struct {
	service.BaseService

	idr  TxIndexer
	evsw events.EventSwitch
}

// NewIndexerService returns a new service instance.
func NewIndexerService(idr TxIndexer, evsw events.EventSwitch) *IndexerService {
	is := &IndexerService{idr: idr, evsw: evsw}
	is.BaseService = *service.NewBaseService(nil, "IndexerService", is)
	return is
}

func (is *IndexerService) OnStart() error {
	// TODO
	return nil
}

func (is *IndexerService) OnStop() {
	// TODO
}
