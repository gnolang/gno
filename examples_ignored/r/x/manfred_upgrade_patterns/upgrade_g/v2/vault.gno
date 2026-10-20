package vault

import "strconv"

// Same path as v1: the type keeps its identity, so the receipts v1 handed
// out answer with this String. A public realm may only append.
type Receipt struct{ Amount int64 }

func (r *Receipt) String() string { return "receipt for " + strconv.FormatInt(r.Amount, 10) + " (v2)" }
func (r *Receipt) Doubled() int64 { return r.Amount * 2 } // appended

var total int64 // carried: its initializer, if any, is skipped

// Deposit now credits double.
func Deposit(cur realm, n int64) *Receipt {
	n *= 2
	total += n
	count++
	return &Receipt{Amount: n}
}

func Total() int64 { return total }

var count int // appended: starts at zero

// migrate runs in place of init, handed the version it replaces. It runs on
// every versioned redeploy that declares it: guard on from, or remove it in v3.
func migrate(from int) {
	if from == 1 {
		count = 1 // one deposit happened under v1
	}
}

func Count() int { return count } // appended
