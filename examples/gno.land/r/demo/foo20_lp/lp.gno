package foo20_lp

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// TODO: replace int64 with bigint
// TODO: is it a good contract to show how a contract can take some usage fees? probably.
// TODO: extract the generic & pure part of this contract to a p/ package.

const (
	foo20MinDeposit   int64 = 10
	ugnotMinDeposit   int64 = 100_000
	foo20PriceInUgnot int64 = 100 // ugnot has 6 digits, foo20 has 4 digits
)

var (
	foo20Queue Fifo
	ugnotQueue Fifo
)

func DepositFoo20(amount int64) {
	if amount < foo20MinDeposit {
		panic(ufmt.Sprintf("min deposit: %d $FOO20", foo20MinDeposit))
	}

	caller := std.GetOrigCaller()

	// TODO: store amounts on contracts' address or work with approve maybe?

	rest := amount
	for _, entry := range ugnotQueue {
		if rest == 0 {
			break
		}

		// full swap
		if rest > entry.Amount {
			println("not implemented")
			// TODO: send foo20 to entry.Addr
			// TODO: send ugnot to caller
			// TODO: delete entry from queue
			rest -= entry.Amount // update rest counter
			continue
		}

		// partial swap
		{
			println("not implemented")
			// partial swap
			// TODO: send `rest` foo20 to entry.Addr
			// TODO: send `rest` ugnot to caller
			entry.Amount -= rest // patch entry
			rest = 0
		}
	}

	// if we have rest, add a new entry in the queue
	if rest > 0 {
		entry := Entry{Addr: caller, Amount: rest}
		foo20Queue = append(foo20Queue, entry)
	}
}

func DepositUgnot() {
	sent := std.GetOrigSend()
	amount := sent.AmountOf("ugnot")
	if amount < ugnotMinDeposit {
		panic(ufmt.Sprintf("min deposit: %d $UGNOT", ugnotMinDeposit))
	}
	println("not implemented")
}

func WithdrawFoo20(amount int64) {
	println("not implemented")
}

func WithdrawUgnot(amount int64) {
	println("not implemented")
}

func Render(path string) string {
	var (
		ugnotEntries = ugnotQueue.Size()
		ugnotTotal   = ugnotQueue.TotalAmount()
		foo20Entries = foo20Queue.Size()
		foo20Total   = foo20Queue.TotalAmount()
	)
	switch {
	case ugnotEntries == 0 && foo20Entries == 0:
		return "LP is empty."
	case ugnotEntries > 0:
		foo20Capa := ugnotTotal * foo20PriceInUgnot
		return ufmt.Sprintf("LP has a total of %d ugnot in %d orders for a capacity of %d foo20.", ugnotTotal, ugnotEntries, foo20Capa)
	case foo20Entries > 0:
		ugnotCapa := foo20Total / foo20PriceInUgnot
		return ufmt.Sprintf("LP has a total of %d foo20 in %d orders for a capacity of %d ugnot.", foo20Total, foo20Entries, ugnotCapa)
	default:
		panic("should not happen")
	}
}
