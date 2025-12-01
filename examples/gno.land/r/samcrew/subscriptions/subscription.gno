package subscriptions

import (
	"chain"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

func init() {
	users = avl.NewTree()
	services = avl.NewTree()
}

type subscription struct {
	serviceID     seqid.ID
	balance       chain.Coins
	startedAt     time.Time
	lastChargedAt time.Time
	initialPaid   bool
}

func (sub *subscription) dueAmount(renewalPeriod time.Duration, price chain.Coins) (chain.Coins, bool) {
	elapsed := time.Now().Sub(sub.lastChargedAt)
	periodsElapsed := int64(elapsed / renewalPeriod)
	if !sub.initialPaid {
		periodsElapsed += 1
	}
	if periodsElapsed <= 0 {
		return chain.Coins{}, true
	}

	due := chain.Coins{}
	for i := int64(0); i < periodsElapsed; i++ {
		due = addCoins(price, due)
		if lessCoinsThan(sub.balance, due) {
			return due, false
		}
	}

	return due, true
}

func (sub *subscription) isActive(renewalPeriod time.Duration, price chain.Coins) bool {
	_, ok := sub.dueAmount(renewalPeriod, price)
	return ok
}
