//go:build !ledger_mock
// +build !ledger_mock

package client

import "testing"

func TestAdd_Ledger(t *testing.T) {
	t.Skip("Please enable the 'ledger_mock' build tags")
}
