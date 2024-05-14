//go:build !ledger_suite
// +build !ledger_suite

package ledger

import (
	ledger_go "github.com/cosmos/ledger-cosmos-go"
)

// discoverLedger defines a function to be invoked at runtime for discovering
// a connected Ledger device.
var discoverLedger discoverLedgerFn = func() (LedgerSECP256K1, error) {
	device, err := ledger_go.FindLedgerCosmosUserApp()
	if err != nil {
		return nil, err
	}

	return device, nil
}
