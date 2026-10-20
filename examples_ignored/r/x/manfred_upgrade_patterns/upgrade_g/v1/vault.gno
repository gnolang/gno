package vault

import "strconv"

type Receipt struct{ Amount int64 }

func (r *Receipt) String() string { return "receipt for " + strconv.FormatInt(r.Amount, 10) + " (v1)" }

var total int64

func Deposit(cur realm, n int64) *Receipt {
	total += n
	return &Receipt{Amount: n}
}

func Total() int64 { return total }
