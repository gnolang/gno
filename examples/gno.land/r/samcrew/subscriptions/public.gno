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

type service struct {
	id            seqid.ID
	owner         *ownable.Ownable
	displayName   string
	description   string
	renewalPeriod time.Duration
	price         chain.Coins
	subscribers   *avl.Tree // key: address, value: *subscription
}

func (s *service) balance() chain.Coins {
	balance := chain.Coins{}
	s.subscribers.Iterate("", "", func(key string, value interface{}) bool {
		sub := value.(*subscription)
		due := sub.dueAmount(s.renewalPeriod, s.price)
		balance = addCoins(balance, due)
		return false
	})
	return balance
}

type subscription struct {
	serviceID     seqid.ID
	balance       chain.Coins
	startedAt     time.Time
	lastChargedAt time.Time
	initialPaid   bool
}

func (s *subscription) dueAmount(renewalPeriod time.Duration, price chain.Coins) chain.Coins {
	elapsed := time.Now().Sub(s.lastChargedAt)
	periods := int64(elapsed / renewalPeriod)
	if !s.initialPaid {
		periods += 1
	}
	if periods <= 0 {
		return chain.Coins{}
	}
	return multiplyCoins(price, periods)
}

func Subscribe(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()
	s := mustGetServiceByName(serviceName)
	if s.subscribers.Has(creator.String()) {
		panic("already subscribed")
	}
	send := banker.OriginSend()
	if !coinsHasPositive(send) || lessCoinsThan(send, s.price) {
		panic("initial deposit must be positive and at least the service price")
	}
	sub := &subscription{
		serviceID:     s.id,
		balance:       send,
		startedAt:     time.Now(),
		lastChargedAt: time.Now(),
		initialPaid:   false,
	}
	s.subscribers.Set(creator.String(), sub)
}

func Unsubscribe(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()
	s := mustGetServiceByName(serviceName)
	if !s.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := s.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due := sub.dueAmount(s.renewalPeriod, s.price)
	if coinsHasPositive(due) {
		sendCoins(s.owner.Owner(), due)
	}
	remaining := subCoins(sub.balance, due)
	if coinsHasPositive(remaining) {
		sendCoins(creator, remaining)
	}
	s.subscribers.Remove(creator.String())
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

	s := mustGetServiceByName(serviceName)
	if !s.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := s.subscribers.Get(creator.String())
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

	s := mustGetServiceByName(serviceName)
	if !s.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := s.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	sub.balance = addCoins(sub.balance, send)
}

func Withdraw(cur realm, serviceName string, denom string, amount int64) {
	creator := runtime.PreviousRealm().Address()
	s := mustGetServiceByName(serviceName)
	if !s.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	// When withdrawing it's important to withdraw only the available balance so without the due amount
	subValue, _ := s.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due := sub.dueAmount(s.renewalPeriod, s.price)
	available := subCoins(sub.balance, due)
	withdraw := chain.NewCoins(chain.NewCoin(denom, amount))
	if lessCoinsThan(available, withdraw) {
		panic("insufficient available balance to withdraw")
	}
	sub.balance = subCoins(sub.balance, withdraw)
	sendCoins(creator, withdraw)
}

func init() {
	users = avl.NewTree()
	services = avl.NewTree()
}

func NewService(cur realm, displayName, description string, renewalPeriod time.Duration, price chain.Coins) {
	// XXX: prefix with creator realm to avoid name clashes?
	if services.Has(displayName) {
		panic("service with the same display name already exists")
	}
	creator := runtime.PreviousRealm().Address()
	s := &service{
		id:            id.Next(),
		owner:         ownable.NewWithAddress(creator),
		displayName:   displayName,
		description:   description,
		renewalPeriod: renewalPeriod,
		price:         price,
		subscribers:   avl.NewTree(),
	}
	services.Set(displayName, s)
}

func ServiceClaimVault(cur realm, displayName string) {
	s := mustGetServiceByName(displayName)
	s.owner.AssertOwnedByPrevious()
	balance := s.balance()
	if !coinsHasPositive(balance) {
		panic("no funds to claim")
	}
	sendCoins(s.owner.Owner(), balance)

}

func mustGetServiceByName(displayName string) *service {
	s := getServiceByName(displayName)
	if s == nil {
		panic("service not found")
	}
	return s
}

func getServiceByName(displayName string) *service {
	value, exists := services.Get(displayName)
	if !exists {
		return nil
	}
	return value.(*service)
}
