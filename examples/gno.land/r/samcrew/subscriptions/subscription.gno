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
	"gno.land/r/samcrew/normalizedcoins"
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
		due = normalizedcoins.AddCoins(due, price)
		if normalizedcoins.LessCoinsThan(sub.vault, due) {
			return normalizedcoins.SubCoins(due, price), false
		}
	}

	if charge {
		sub.vault = normalizedcoins.SubCoins(sub.vault, due)
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
	return normalizedcoins.SubCoins(sub.vault, dueAmount)
}

func topupNative(cur realm, svcKey string) {
	caller := runtime.PreviousRealm().Address()
	send := banker.OriginSend()
	if !normalizedcoins.CoinsHasPositive(send) {
		panic("topup amount must be positive")
	}

	prefixedSend, err := normalizedcoins.PrefixCoins(send, nil)
	if err != nil {
		panic(err)
	}

	svc := mustGetService(svcKey)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + svcKey)
	}
	subValue, _ := svc.subscribers.Get(caller.String())
	sub := subValue.(*subscription)
	sub.vault = normalizedcoins.AddCoins(sub.vault, prefixedSend)
}

func topupGRC20(cur realm, svcKey string, fqName string, amount int64) {
	token := grc20reg.MustGet(fqName)
	if token == nil {
		panic(errors.New("token not found: " + fqName))
	}
	caller := runtime.PreviousRealm().Address()

	anyAmount := amount == -1
	if !anyAmount && amount <= 0 {
		panic("topup amount must be positive or -1 for max")
	}

	if anyAmount {
		amount = int64(token.Allowance(caller, runtime.CurrentRealm().Address()))
	}

	grc20Send := chain.NewCoins(chain.NewCoin(fqName, amount))
	prefixedSend, err := normalizedcoins.PrefixCoins(nil, grc20Send)
	if err != nil {
		panic(err)
	}

	teller := token.RealmTeller()
	if err := teller.TransferFrom(caller, runtime.CurrentRealm().Address(), amount); err != nil {
		panic(err)
	}

	svc := mustGetService(svcKey)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + svcKey)
	}
	subValue, _ := svc.subscribers.Get(caller.String())
	sub := subValue.(*subscription)
	sub.vault = normalizedcoins.AddCoins(sub.vault, prefixedSend)
}
