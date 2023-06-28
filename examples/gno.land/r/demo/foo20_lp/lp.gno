package foo20_lp

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// TODO: replace int64 with bigint
// TODO: is it a good contract to show how a contract can take some usage fees? probably.
// TODO: extract the generic & pure part of this contract to a p/ package.
// TODO: add stats?

const (
	foo20PriceInUgnot int64 = 100 // ugnot has 6 digits, foo20 has 4 digits
	foo20MinDeposit   int64 = 2   // min 1
	ugnotMinDeposit   int64 = 300 // min foo20PriceInUgnot
)

var (
	foo20Queue Fifo
	ugnotQueue Fifo
)

func DepositFoo20(amount int64) {
	if amount < foo20MinDeposit {
		panic(ufmt.Sprintf("min deposit: %d foo20.", foo20MinDeposit))
	}
	// XXX: if foo20PriceInUgnot was < 1, we should check for `amount % ugnotPriceInFoo20 == 0`.

	caller := std.GetOrigCaller()

	// TODO: store amounts on contracts' address or work with approve maybe?

	rest := amount
	for _, entry := range ugnotQueue {
		if rest == 0 {
			break
		}

		// full swap
		if rest > entry.Amount {
			println("not implemented: A")
			// TODO: send foo20 to entry.Addr
			// TODO: send ugnot to caller
			// TODO: delete entry from queue
			rest -= entry.Amount // update rest counter
			continue
		}

		// partial swap
		{
			println("not implemented: B")
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
	caller := std.GetOrigCaller()
	pkgaddr := std.GetOrigPkgAddr()
	sent := std.GetOrigSend()
	amount := sent.AmountOf("ugnot")
	if amount < ugnotMinDeposit {
		panic(ufmt.Sprintf("jmin deposit: %d ugnot.", ugnotMinDeposit))
	}
	if amount%foo20PriceInUgnot != 0 {
		panic(ufmt.Sprintf("ugnot deposit should be a multiple of %d, was %d.", foo20PriceInUgnot, amount))
	}
	banker := std.GetBanker(std.BankerTypeOrigSend)
	send := std.Coins{{"ugnot", 1}}
	println("hey morgan")
	// banker.SendCoins(caller, pkgaddr, send)
	println("not implemented: C")
}

func WithdrawFoo20(amount int64) {
	println("not implemented: D")
}

func WithdrawUgnot(amount int64) {
	println("not implemented: E")
}

func Render(path string) string {
	switch path {
	case "": // home
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
			foo20Capa := ugnotTotal / foo20PriceInUgnot
			return ufmt.Sprintf("LP has a total of %d ugnot in %d orders for a capacity of %d foo20.", ugnotTotal, ugnotEntries, foo20Capa)
		case foo20Entries > 0:
			ugnotCapa := foo20Total * foo20PriceInUgnot
			return ufmt.Sprintf("LP has a total of %d foo20 in %d orders for a capacity of %d ugnot.", foo20Total, foo20Entries, ugnotCapa)
		default:
			panic("should not happen")
		}
	case "queues":
		if foo20Queue.Empty() && ugnotQueue.Empty() {
			return "## Queues are empty"
		}

		output := ""
		if !foo20Queue.Empty() {
			output += ufmt.Sprintf("## foo20Queue (total=%d, orders=%d)\n", foo20Queue.TotalAmount(), foo20Queue.Size())
			var cumulated int64
			for _, entry := range foo20Queue {
				cumulated += entry.Amount
				output += ufmt.Sprintf("- addr=%s amount=%d (cum=%d)\n", entry.Addr, entry.Amount, cumulated)
			}
		}
		if !ugnotQueue.Empty() {
			output += ufmt.Sprintf("## ugnotQueue (total=%d, orders=%d)\n", ugnotQueue.TotalAmount(), ugnotQueue.Size())
			var cumulated int64
			for _, entry := range ugnotQueue {
				cumulated += entry.Amount
				output += ufmt.Sprintf("- addr=%s amount=%d (cum=%d)\n", entry.Addr, entry.Amount, cumulated)
			}
		}
		return output
	default:
		return "404"
	}
}
