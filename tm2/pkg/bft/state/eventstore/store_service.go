package eventstore

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// Service connects the event bus and event store together in order
// to store events coming from event bus
type Service struct {
	service.BaseService

	cancelFn context.CancelFunc

	txEventStore TxEventStore
	evsw         events.EventSwitch
}

// NewEventStoreService returns a new service instance
func NewEventStoreService(idr TxEventStore, evsw events.EventSwitch) *Service {
	is := &Service{txEventStore: idr, evsw: evsw}
	is.BaseService = *service.NewBaseService(nil, "EventStoreService", is)

	return is
}

func (is *Service) OnStart() error {
	// Create a context for the intermediary monitor service
	ctx, cancelFn := context.WithCancel(context.Background())
	is.cancelFn = cancelFn

	// Start the event store
	if err := is.txEventStore.Start(); err != nil {
		return fmt.Errorf("unable to start transaction event store, %w", err)
	}

	// Start the intermediary monitor service
	go is.monitorTxEvents(ctx)

	return nil
}

func (is *Service) OnStop() {
	// Close off any routines
	is.cancelFn()

	// Attempt to gracefully stop the event store
	if err := is.txEventStore.Stop(); err != nil {
		is.Logger.Error(
			fmt.Sprintf("unable to gracefully stop event store, %v", err),
		)
	}
}

// monitorTxEvents acts as an intermediary feed service for the supplied
// event store. It relays transaction events that come from the event stream
func (is *Service) monitorTxEvents(ctx context.Context) {
	// Create a subscription for transaction events
	subCh := events.SubscribeToEvent(is.evsw, "tx-event-store", types.EventTx{})

	for {
		select {
		case <-ctx.Done():
			return
		case evRaw := <-subCh:
			// Cast the event
			ev, ok := evRaw.(types.EventTx)
			if !ok {
				is.Logger.Error("invalid transaction result type cast")

				continue
			}

			// Alert the actual tx event store
			if err := is.txEventStore.Append(ev.Result); err != nil {
				is.Logger.Error("unable to store transaction", "err", err)
			}
		}
	}
}
