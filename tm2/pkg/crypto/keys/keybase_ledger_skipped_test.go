//go:build !ledger_suite
// +build !ledger_suite

package keys

import "testing"

func TestCreateLedgerUnsupportedAlgo(t *testing.T) {
	t.Parallel()

	t.Skip("this test needs to be run with the `ledger_suite` tag enabled")
}

func TestCreateLedger(t *testing.T) {
	t.Parallel()

	t.Skip("this test needs to be run with the `ledger_suite` tag enabled")
}
