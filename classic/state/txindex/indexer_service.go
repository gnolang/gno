package txindex

import (
	cmn "github.com/tendermint/classic/libs/common"
)

// IndexerService connects event bus and transaction indexer together in order
// to index transactions coming from event bus.
type IndexerService struct {
	cmn.BaseService

	idr TxIndexer
}

// NewIndexerService returns a new service instance.
func NewIndexerService(idr TxIndexer) *IndexerService {
	is := &IndexerService{idr: idr}
	is.BaseService = *cmn.NewBaseService(nil, "IndexerService", is)
	return is
}

func (is *IndexerService) OnStart() error {
	// TODO
	return nil
}

func (is *IndexerService) OnStop() {
	// TODO
}
