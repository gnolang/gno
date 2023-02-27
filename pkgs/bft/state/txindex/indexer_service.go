package txindex

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/events"
	"github.com/gnolang/gno/pkgs/service"
)

// IndexerService connects event bus and transaction indexer together in order
// to index transactions coming from event bus.
type IndexerService struct {
	service.BaseService

	cancelFn context.CancelFunc

	indexer  TxIndexer
	evSwitch events.EventSwitch
}

// NewIndexerService returns a new service instance.
func NewIndexerService(idr TxIndexer, evsw events.EventSwitch) *IndexerService {
	is := &IndexerService{indexer: idr, evSwitch: evsw}
	is.BaseService = *service.NewBaseService(nil, "IndexerService", is)

	return is
}

func (is *IndexerService) OnStart() error {
	// Create a context for the intermediary monitor service
	ctx, cancelFn := context.WithCancel(context.Background())
	is.cancelFn = cancelFn

	// Start the indexer
	if err := is.indexer.Start(); err != nil {
		return fmt.Errorf("unable to start transaction indexer, %w", err)
	}

	// Start the intermediary monitor service
	go is.monitorTxEvents(ctx)

	return nil
}

func (is *IndexerService) OnStop() {
	// Close off any routines
	is.cancelFn()

	// Attempt to gracefully stop the transaction indexer
	if err := is.indexer.Close(); err != nil {
		is.Logger.Error(
			fmt.Sprintf("unable to gracefully stop transaction indexer, %v", err),
		)
	}
}

// monitorTxEvents acts as an intermediary feed service for the supplied
// transaction indexer. It relays transaction events that come from the event stream
func (is *IndexerService) monitorTxEvents(ctx context.Context) {
	// Create a subscription for transaction events
	subCh := events.SubscribeToEvent(is.evSwitch, "tx-indexer", types.TxResult{})

	for {
		select {
		case <-ctx.Done():
			return
		case evRaw := <-subCh:
			// Cast the event
			ev, ok := evRaw.(*types.TxResult)
			if !ok {
				is.Logger.Error("invalid transaction result type cast")

				continue
			}

			// Alert the actual indexer
			if err := is.indexer.Index(ev); err != nil {
				is.Logger.Error(
					fmt.Sprintf("unable to index transaction, %v", err),
				)
			}
		}
	}
}
