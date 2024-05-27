package event

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/gno.me/state"
)

type Processor struct {
	eventApplier Applier
	eventCh      chan *state.Event
	done         chan struct{}
}

func NewProcessor(applier Applier, eventCh chan *state.Event, done chan struct{}) *Processor {
	return &Processor{
		eventApplier: applier,
		eventCh:      eventCh,
		done:         done,
	}
}

func (p Processor) Process() {
	for {
		select {
		case event, ok := <-p.eventCh:
			if !ok {
				close(p.done)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			if err := p.eventApplier.ApplyEvent(ctx, event); err != nil {
				fmt.Println("error applying event:", err)
			}
			cancel()
		case <-p.done:
			return
		}
	}
}
