package gnousd

import (
	"std"

	"gno.land/p/avl"
	grc20 "gno.land/p/grc/grc20/impl"
	"gno.land/p/ufmt"
)

const (
	mintConstant uint64 = 10 // supply contraction/expansion as a percentage of price change
	// TODO: check heights
	cycleInterval uint64 = 2000 // in heights
	targetPrice   uint64 = 1e6  // $1. in gnousd-USD
	minBid        uint64 = 1e16
	decimals      int    = 18
)

type Controller struct {
	shares     *grc20.Token
	coins      *grc20.Token
	oracle     std.Address
	coinPrice  uint64
	sharePrice uint64
	// TODO: cycles []Cycle?
	cycles  *avl.MutTree // uint64 -> Cycle
	counter uint64
}

func newController(shares, coins *grc20.Token, oracle std.Address) *Controller {
	if shares.GetDecimals() != coins.GetDecimals() {
		panic("coins and share should have the same amount of decimals")
	}

	ctrl := &Controller{
		shares:     shares,
		coins:      coins,
		oracle:     oracle,
		coinPrice:  1e6,   // in gnousd-USD, 1e6 = $1
		sharePrice: 100e6, // in gnousd-coins, 100e6 = 1 coin
		cycles:     avl.NewMutTree(),
		counter:    0,
	}

	// start first cycle
	currentHeight := getCurrentHeight() // TODO: this does not work, at least in tests, probably because `init` has a different context?
	// currentHeight := uint64(0)
	cycle := Cycle{
		direction:   Neutral,
		toMint:      0,
		startHeight: currentHeight,
		bidTotal:    0,
		bids:        avl.NewMutTree(),
	}
	ctrl.setCycle(cycle)

	return ctrl
}

func (ctrl *Controller) UpdateCoinPrice(price uint64) {
	// the updated price must be within 10% of the old price, to prevent accidental mispricing.
	// it's possible to apply bigger changes by making multiple transactions.
	if price <= ctrl.coinPrice*9/10 {
		panic("The new price must not be more than 10% different from the old one.")
	}
	if price >= ctrl.coinPrice*11/10 {
		panic("The new price must not be more than 10% different from the old one.")
	}
	ctrl.coinPrice = price
}

func (ctrl *Controller) UpdateSharePrice(price uint64) {
	// the updated price must be within 10% of the old price, to prevent accidental mispricing.
	// it's possible to apply bigger changes by making multiple transactions.
	if price <= ctrl.sharePrice*9/10 {
		panic("The new price must not be more than 10% different from the old one.")
	}
	if price >= ctrl.sharePrice*11/10 {
		panic("The new price must not be more than 10% different from the old one.")
	}
	ctrl.sharePrice = price
}

func (ctrl *Controller) NewCycle() {
	oldCycle := ctrl.getCurrentCycle()

	currentHeight := getCurrentHeight()
	if currentHeight < oldCycle.startHeight+cycleInterval {
		panic(ufmt.Sprintf("new cycle can only happen after %d heights", cycleInterval))
	}

	// burn previous cycle's bids
	switch oldCycle.direction {
	case Contracting:
		ctrl.coins.Burn(reserveAddr, oldCycle.bidTotal)
	case Expanding:
		ctrl.shares.Burn(reserveAddr, oldCycle.bidTotal)
	}

	// determine monetary policy for cycle.
	newCycle := Cycle{
		startHeight: currentHeight,
		bidTotal:    0,
	}
	targetSupply := ctrl.coins.TotalSupply() * ctrl.coinPrice / targetPrice
	switch {
	case ctrl.coinPrice == targetPrice:
		newCycle.direction = Neutral
	case ctrl.coinPrice < targetPrice:
		newCycle.direction = Contracting
		newCycle.toMint = (ctrl.coins.TotalSupply() - targetSupply) * 1e6 / ctrl.sharePrice

		// mint shares
		ctrl.shares.Mint(reserveAddr, newCycle.toMint)
	default:
		newCycle.direction = Expanding
		newCycle.toMint = (targetSupply - ctrl.coins.TotalSupply()) * mintConstant / 100

		// mint coins
		ctrl.coins.Mint(reserveAddr, newCycle.toMint)
	}
	ctrl.setCycle(newCycle)
}

func (ctrl *Controller) Bid(amount uint64) {
	caller := std.GetCallerAt(3)
	cycle := ctrl.getCurrentCycle()
	currentHeight := getCurrentHeight()

	if currentHeight >= cycle.startHeight+cycleInterval {
		panic("current cycle is too old to accept new bids")
	}
	if cycle.direction == Neutral {
		panic("cannot take bid while the price is not varying") // TODO: dig on more papers -> why?
	}
	if amount < minBid {
		panic(ufmt.Sprintf("minimum bid amount is: %d", minBid))
	}

	// transfer coins/shares to the reserve.
	// the user needs to call Approve(reserve, amount), else it will raise an error.
	switch cycle.direction {
	case Contracting:
		ctrl.coins.TransferFrom(reserveAddr, caller, reserveAddr, amount)
	case Expanding:
		ctrl.shares.TransferFrom(reserveAddr, caller, reserveAddr, amount)
	}

	currentBid := cycle.getBids(caller)
	newBid := currentBid + amount
	cycle.setBids(caller, newBid)
	cycle.bidTotal += amount
}

func (ctrl *Controller) Claim(cycleID uint64) {
	caller := std.GetCallerAt(3)
	cycle := ctrl.getCycle(cycleID)
	userBids := cycle.getBids(caller)
	currentHeight := getCurrentHeight()

	if currentHeight <= cycle.startHeight+cycleInterval {
		panic("cycle is still active and cannot be claimed")
	}
	if userBids == 0 {
		panic("nothing to claim")
	}
	if cycle.direction == Neutral {
		panic("cannot claim during a neutral cycle")
	}

	amount := userBids * cycle.toMint / cycle.bidTotal

	switch cycle.direction {
	case Contracting:
		ctrl.shares.Transfer(reserveAddr, caller, amount)
	case Expanding:
		ctrl.coins.Transfer(reserveAddr, caller, amount)
	}

	cycle.delBids(caller)
}

func (ctrl *Controller) UserBids(cycleID uint64, addr std.Address) uint64 {
	cycle := ctrl.getCycle(cycleID)
	return cycle.getBids(addr)
}

func (ctrl *Controller) CurrentBidPrice() uint64 {
	cycle := ctrl.getCurrentCycle()
	price := cycle.bidTotal * 1e6 / cycle.toMint
	return price
}

func (ctrl *Controller) getCurrentCycle() Cycle {
	return ctrl.getCycle(ctrl.counter)
}

func (ctrl *Controller) getCycle(cycleID uint64) Cycle {
	idStr := ufmt.Sprintf("%d", cycleID)
	cycle, found := ctrl.cycles.Get(idStr)
	if !found {
		panic("no such cycle")
	}
	return cycle.(Cycle)
}

func (ctrl *Controller) setCycle(cycle Cycle) {
	ctrl.counter++
	counterStr := ufmt.Sprintf("%d", ctrl.counter)
	ctrl.cycles.Set(counterStr, cycle)
}

type Cycle struct {
	direction   Direction
	toMint      uint64
	startHeight uint64
	bidTotal    uint64
	bids        *avl.MutTree // std.Address -> uint64
}

func (c *Cycle) getBids(addr std.Address) uint64 {
	bid, found := c.bids.Get(addr.String())
	if !found {
		return 0
	}
	return bid.(uint64)
}
func (c *Cycle) setBids(addr std.Address, amount uint64) { c.bids.Set(addr.String(), amount) }
func (c *Cycle) delBids(addr std.Address)                { c.bids.Remove(addr.String()) }
func (c *Cycle) maxHeight() uint64                       { return c.startHeight + cycleInterval }

type Direction uint

const (
	Neutral     Direction = iota
	Expanding   Direction = iota
	Contracting Direction = iota
)
