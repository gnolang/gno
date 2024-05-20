//go:build !ledger_suite
// +build !ledger_suite

package client

import "testing"

func TestAdd_Ledger(t *testing.T) {
	t.Skip("Please enable the 'ledger_suite' build tags")
}
