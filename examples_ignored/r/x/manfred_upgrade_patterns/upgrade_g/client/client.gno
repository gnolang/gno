package client

import "gno.land/r/x/manfred_upgrade_patterns/upgrade_g/vault"

// An immutable realm could not import vault: it may not depend on code that
// can change under it. This one is upgradeable itself.
var receipts []*vault.Receipt

func Deposit(cur realm, n int64) {
	receipts = append(receipts, vault.Deposit(cross(cur), n))
}

func First() string { return receipts[0].String() }
func Last() string  { return receipts[len(receipts)-1].String() }
