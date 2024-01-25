package events

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type EventType string

const (
	EvtReload         EventType = "EVENT_NODE_RELOAD"
	EvtReset          EventType = "EVENT_NODE_RESET"
	EvtPackagesUpdate EventType = "EVENT_PACKAGES_UPDATE"
	EvtTxResult       EventType = "EVENT_TX_RESULT"
)

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}

// Event Reload

type EventReload struct{}

func NewEventReload() *Event {
	return &Event{
		Type: EvtReload,
		Data: &EventReload{},
	}
}

// Event Reset

type EventReset struct{}

func NewEventReset() *Event {
	return &Event{
		Type: EvtReset,
		Data: &EventReset{},
	}
}

// Event Packages Update

type PackageUpdate struct {
	Package string   `json:"package"`
	Files   []string `json:"files"`
}

type PackagesUpdateEvent struct {
	Pkgs []PackageUpdate `json:"packages"`
}

func NewPackagesUpdateEvent(pkgs []PackageUpdate) *Event {
	return &Event{
		Type: EvtPackagesUpdate,
		Data: &PackagesUpdateEvent{
			Pkgs: pkgs,
		},
	}
}

// Event Tx is an alias to TxResult

type EventTxResult struct {
	Height   int64                  `json:"height"`
	Index    uint32                 `json:"index"`
	Tx       std.Tx                 `json:"tx"`
	Response abci.ResponseDeliverTx `json:"response"`
}

func NewTxEventResult(result types.TxResult) (*Event, error) {
	evt := &EventTxResult{
		Height:   result.Height,
		Index:    result.Index,
		Response: result.Response,
	}
	if err := amino.Unmarshal(result.Tx, &evt.Tx); err != nil {
		return nil, fmt.Errorf("unable unmarshal tx: %w", err)
	}

	return &Event{
		Type: EvtTxResult,
		Data: evt,
	}, nil
}
