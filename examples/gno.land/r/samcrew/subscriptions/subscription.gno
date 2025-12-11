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

func (sub *subscription) dueAmount(charge bool, renewalPeriod time.Duration, price chain.Coins) (chain.Coins, bool) {
	elapsed := time.Now().Sub(sub.lastChargedAt)
	periodsElapsed := int64(elapsed / renewalPeriod)
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
	for i := range send {
		send[i].Denom = "/native/" + send[i].Denom
	}

	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	sub.vault = addCoins(sub.vault, send)
}

func topupGRC20(cur realm, serviceName string, fqName string, amount int64) {
	token := grc20reg.MustGet(fqName)
	if token == nil {
		panic(errors.New("token not found: " + fqName))
	}
	creator := runtime.PreviousRealm().Address()
	send := chain.NewCoins(chain.NewCoin("/grc20/"+fqName, amount))

	anyAmount := amount == -1
	if !anyAmount && amount <= 0 {
		panic("topup amount must be positive or -1 for max")
	}

	if anyAmount {
		amount = int64(token.Allowance(creator, runtime.CurrentRealm().Address()))
	}

	teller := token.RealmTeller()
	if err := teller.TransferFrom(creator, runtime.CurrentRealm().Address(), amount); err != nil {
		panic(err)
	}

	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	sub.vault = addCoins(sub.vault, send)
}
