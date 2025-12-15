package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/r/demo/defi/grc20reg"
)

func init() {
	users = avl.NewTree()
	services = avl.NewTree()
}

type subscription struct {
	serviceID     seqid.ID
	vault         chain.Coins
	startedAt     time.Time
	lastChargedAt time.Time
	initialPaid   bool
}

// dueAmount calculates how much the subscriber owes based on elapsed time.
// If charge is true, it also deducts the amount from vault and updates lastChargedAt.
// Returns the amount due and whether the subscription can cover all periods.
func (sub *subscription) dueAmount(charge bool, renewalPeriod time.Duration, price chain.Coins) (chain.Coins, bool) {
	if renewalPeriod <= 0 {
		panic("renewal period must be positive")
	}

	elapsed := time.Now().Sub(sub.lastChargedAt)
	periodsElapsed := int64(elapsed / renewalPeriod)

	// If initial payment not yet made, add one period
	if !sub.initialPaid {
		periodsElapsed += 1
	}
	if periodsElapsed <= 0 {
		return chain.Coins{}, true
	}

	due := chain.Coins{} // dummy initial value
	for i := int64(0); i < periodsElapsed; i++ {
		due = addCoins(due, price)
		if lessCoinsThan(sub.vault, due) {
			return subCoins(due, price), false
		}
	}

	if charge {
		sub.vault = subCoins(sub.vault, due)
		sub.lastChargedAt = sub.lastChargedAt.Add(time.Duration(periodsElapsed) * renewalPeriod)
		sub.initialPaid = true
	}

	return due, true
}

func (sub *subscription) isActive(renewalPeriod time.Duration, price chain.Coins) bool {
	_, ok := sub.dueAmount(false, renewalPeriod, price)
	return ok
}

func (sub *subscription) balance(renewalPeriod time.Duration, price chain.Coins) chain.Coins {
	dueAmount, _ := sub.dueAmount(false, renewalPeriod, price)
	return subCoins(sub.vault, dueAmount)
}

func topupNative(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()

	runtime.AssertOriginCall()

	send := banker.OriginSend()
	if !coinsHasPositive(send) {
		panic("topup amount must be positive")
	}
	
	prefixedSend, err := prefixCoins(cur, send, nil)
	if err != nil {
		panic(err)
	}

	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed to service: " + serviceName)
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	sub.vault = addCoins(sub.vault, prefixedSend)
}

func topupGRC20(cur realm, serviceName string, fqName string, amount int64) {
	token := grc20reg.MustGet(fqName)
	if token == nil {
		panic(errors.New("token not found: " + fqName))
	}
	creator := runtime.PreviousRealm().Address()

	anyAmount := amount == -1
	if !anyAmount && amount <= 0 {
		panic("topup amount must be positive or -1 for max")
	}

	if anyAmount {
		amount = int64(token.Allowance(creator, runtime.CurrentRealm().Address()))
	}

	grc20Send := chain.NewCoins(chain.NewCoin(fqName, amount))
	prefixedSend, err := prefixCoins(cur, nil, grc20Send)
	if err != nil {
		panic(err)
	}

	teller := token.RealmTeller()
	if err := teller.TransferFrom(creator, runtime.CurrentRealm().Address(), amount); err != nil {
		panic(err)
	}

	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed to service: " + serviceName)
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	sub.vault = addCoins(sub.vault, prefixedSend)
}
