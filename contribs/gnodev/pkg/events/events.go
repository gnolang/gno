package events

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type Type string

const (
	EvtReload         Type = "NODE_RELOAD"
	EvtReset          Type = "NODE_RESET"
	EvtPackagesUpdate Type = "PACKAGES_UPDATE"
	EvtTxResult       Type = "TX_RESULT"
)

type Event struct {
	Type Type `json:"type"`
	Data any  `json:"data"`
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

type EventPackagesUpdate struct {
	Pkgs []PackageUpdate `json:"packages"`
}

func NewEventPackagesUpdate(pkgs []PackageUpdate) *Event {
	return &Event{
		Type: EvtPackagesUpdate,
		Data: &EventPackagesUpdate{
			Pkgs: pkgs,
		},
	}
}

type EventTxResult struct {
	Height   int64                  `json:"height"`
	Index    uint32                 `json:"index"`
	Tx       std.Tx                 `json:"tx"`
	Response abci.ResponseDeliverTx `json:"response"`
}

func NewEventTxResult(result types.TxResult) (*Event, error) {
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
