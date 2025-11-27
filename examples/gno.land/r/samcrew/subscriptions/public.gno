package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
	"gno.land/r/demo/defi/grc20reg"
)

var services *avl.Tree
var users *avl.Tree
var id seqid.ID

// XXX: Handle GRC20 allowances and transfers for initial deposits or let people subscribe and then topup later?
func Subscribe(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if svc.subscribers.Has(creator.String()) {
		panic("already subscribed")
	}
	send := banker.OriginSend()
	sCoins, err := Coins(cur, send, chain.Coins{})
	if err != nil {
		panic(err)
	}
	if !coinsHasPositive(sCoins) || lessCoinsThan(sCoins, svc.price) {
		panic("initial deposit must be positive and at least the service price")
	}
	sub := &subscription{
		serviceID:     svc.id,
		balance:       sCoins,
		startedAt:     time.Now(),
		lastChargedAt: time.Now(),
		initialPaid:   false,
	}
	svc.subscribers.Set(creator.String(), sub)
}

func Unsubscribe(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due := sub.dueAmount(svc.renewalPeriod, svc.price)
	if coinsHasPositive(due) {
		sendCoins(svc.owner.Owner(), due)
	}
	remaining := subCoins(sub.balance, due)
	if coinsHasPositive(remaining) {
		sendCoins(creator, remaining)
	}
	svc.subscribers.Remove(creator.String())
}

func TopupNative(cur realm, serviceName string) {
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
	sub.balance = addCoins(sub.balance, send)
}

func TopupGRC20(cur realm, serviceName string, fqName string, amount int64) {
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
	sub.balance = addCoins(sub.balance, send)
}

func Withdraw(cur realm, serviceName string, denom string, amount int64) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	// When withdrawing it's important to withdraw only the available balance so without the due amount
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due := sub.dueAmount(svc.renewalPeriod, svc.price)
	available := subCoins(sub.balance, due)
	withdraw := chain.NewCoins(chain.NewCoin(denom, amount))
	if lessCoinsThan(available, withdraw) {
		panic("insufficient available balance to withdraw")
	}
	sub.balance = subCoins(sub.balance, withdraw)
	sendCoins(creator, withdraw)
}

// Helper to create through MsgCall with GNOT native token price.
func NewServiceWithMonthlyRenewal(cur realm, displayName, description string, price uint64) {
	coins := chain.NewCoins(chain.NewCoin("ugnot", int64(price)))
	NewService(cur, displayName, description, 30*28*time.Hour, coins, chain.Coins{})
}

func NewService(cur realm, displayName, description string, renewalPeriod time.Duration, native chain.Coins, grc20 chain.Coins) {
	// XXX: prefix with creator realm to avoid name clashes?
	if services.Has(displayName) {
		panic("service with the same display name already exists")
	}
	creator := runtime.PreviousRealm().Address()
	sCoins, err := Coins(cur, native, grc20)
	if err != nil {
		panic(err)
	}
	svc := &service{
		id:            id.Next(),
		owner:         ownable.NewWithAddress(creator),
		displayName:   displayName,
		description:   description,
		renewalPeriod: renewalPeriod,
		price:         sCoins,
		subscribers:   avl.NewTree(),
	}
	services.Set(displayName, svc)
}

func ServiceClaimVault(cur realm, displayName string) {
	svc := mustGetServiceByName(displayName)
	svc.owner.AssertOwnedByPrevious()
	balance := svc.balance()
	if !coinsHasPositive(balance) {
		panic("no funds to claim")
	}
	sendCoins(svc.owner.Owner(), balance)

}
