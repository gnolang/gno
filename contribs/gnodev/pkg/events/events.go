package events

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type Type string

const (
	EvtReload         Type = "NODE_RELOAD"
	EvtReset          Type = "NODE_RESET"
	EvtPackagesUpdate Type = "PACKAGES_UPDATE"
	EvtTxResult       Type = "TX_RESULT"
)

type Event interface {
	Type() Type

	assertEvent()
}

// Reload Event

type Reload struct{}

func (Reload) Type() Type   { return EvtReload }
func (Reload) assertEvent() {}

// Reset Event

type Reset struct{}

func (Reset) Type() Type   { return EvtReset }
func (Reset) assertEvent() {}

// PackagesUpdate Event

type PackagesUpdate struct {
	Pkgs []PackageUpdate `json:"packages"`
}

type PackageUpdate struct {
	Package string   `json:"package"`
	Files   []string `json:"files"`
}

func (PackagesUpdate) Type() Type   { return EvtPackagesUpdate }
func (PackagesUpdate) assertEvent() {}

// TxResult Event

type TxResult struct {
	Height   int64                  `json:"height"`
	Index    uint32                 `json:"index"`
	Tx       std.Tx                 `json:"tx"`
	Response abci.ResponseDeliverTx `json:"response"`
}

func (TxResult) Type() Type   { return EvtTxResult }
func (TxResult) assertEvent() {}
