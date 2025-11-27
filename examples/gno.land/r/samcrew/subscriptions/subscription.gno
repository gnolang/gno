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

// how handle when user have no money left in balance?
func (sub *subscription) dueAmount(renewalPeriod time.Duration, price chain.Coins) chain.Coins {
	elapsed := time.Now().Sub(sub.lastChargedAt)
	periods := int64(elapsed / renewalPeriod)
	if !sub.initialPaid {
		periods += 1
	}
	if periods <= 0 {
		return chain.Coins{}
	}
	// XXX: Compute how many periods user can pay, in case balance is insufficient due amount should be limited to periods user can pay
	// XXX: It's up to the owner of the service to use IsActive to check if subscription is still active based on balance
	return multiplyCoins(price, periods)
}

// check if the subscription is still active based on the last charged time and renewal period
func (sub *subscription) isActive(renewalPeriod time.Duration, price chain.Coins) bool {
	// XXX: don't use dueAmont since the dueAmount implem have to be fixed to handle insufficient balance case
	elapsed := time.Now().Sub(sub.lastChargedAt)
	periods := int64(elapsed / renewalPeriod)
	if !sub.initialPaid {
		periods += 1
	}
	if periods <= 0 {
		return true
	}
	due := multiplyCoins(price, periods)
	return !lessCoinsThan(sub.balance, due)
}
