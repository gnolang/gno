package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
)

var services *avl.Tree // key: displayName, value: *service
var users *avl.Tree    // key: user address, value: avl.Tree of subscribed service names and empty struct{}
var id seqid.ID

// Subscribe allows a user to subscribe to a service by providing the service name
// - serviceName: the name of the service to subscribe to
// - fqName: the fully qualified name of the GRC20 token to use for payment (empty string for native coins)
// - amount: the amount of GRC20 tokens to deposit initially (0 for native coins)
func Subscribe(cur realm, serviceName string, fqName string, amount int64) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if svc.subscribers.Has(creator.String()) {
		panic("already subscribed")
	}
	natSend := banker.OriginSend()
	grc20Send := chain.Coins{}
	if amount != 0 {
		grc20Send = chain.NewCoins(chain.NewCoin("/grc20/"+fqName, amount))
	}
	sCoins, err := Coins(cur, natSend, grc20Send)
	if err != nil {
		panic(err)
	}
	// XXX: support multiple initial GRC20 tokens
	if !coinsHasPositive(sCoins) || lessCoinsThan(sCoins, svc.price) {
		panic("initial deposit must be positive and at least the service price")
	}
	sub := &subscription{
		serviceID:     svc.id,
		vault:         sCoins,
		startedAt:     time.Now(),
		lastChargedAt: time.Now(),
		initialPaid:   false,
	}
	svc.subscribers.Set(creator.String(), sub)

	var userSubs *avl.Tree
	userSubsRaw, ok := users.Get(creator.String())
	if !ok {
		userSubs = avl.NewTree()
	} else {
		userSubs = userSubsRaw.(*avl.Tree)
	}
	userSubs.Set(serviceName, struct{}{})
	users.Set(creator.String(), userSubs)
}

// Unsubscribe allows a user to unsubscribe from a service by providing the service name
// It refunds any remaining balance after deducting due amounts to the service owner
// - serviceName: the name of the service to unsubscribe from
func Unsubscribe(cur realm, serviceName string) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(svc.renewalPeriod, svc.price)
	if coinsHasPositive(due) {
		sendCoins(svc.owner.Owner(), due)
	}
	remaining := subCoins(sub.vault, due)
	if coinsHasPositive(remaining) {
		sendCoins(creator, remaining)
	}
	svc.subscribers.Remove(creator.String())

	userSubsRaw, _ := users.Get(creator.String())
	userSubs := userSubsRaw.(*avl.Tree)
	userSubs.Remove(serviceName)
	if userSubs.Size() == 0 {
		users.Remove(creator.String())
	}
}

// Topup increases a user's subscription balance using native coins, GRC20 tokens,
// or both in the same transaction. Native funds are inferred from the tx.
// - serviceName: target service
// - fqName: GRC20 token denom (optional)
// - amount: amount of GRC20 tokens (ignored if fqName empty)
func Topup(cur realm, serviceName string, fqName string, amount int64) {
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(runtime.PreviousRealm().Address().String()) {
		panic("not subscribed")
	}
	subRaw, _ := svc.subscribers.Get(runtime.PreviousRealm().Address().String())
	sub := subRaw.(*subscription)
	_, active := sub.dueAmount(svc.renewalPeriod, svc.price)
	native := banker.OriginSend()
	if coinsHasPositive(native) {
		topupNative(cur, serviceName)
	}
	if amount != 0 {
		topupGRC20(cur, serviceName, fqName, amount)
	}
	// XXX: reset subscription if it was inactive and now has enough balance
	if !active && !lessCoinsThan(sub.vault, svc.price) {
		sub.startedAt = time.Now()
		sub.lastChargedAt = time.Now()
		sub.initialPaid = false
	}
}

// Withdraw allows a user to withdraw available balance from their subscription
// - serviceName: the name of the service to withdraw from
// - denom: the denomination of the coin to withdraw
// - amount: the amount of the coin to withdraw
func Withdraw(cur realm, serviceName string, denom string, amount int64) {
	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed")
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(svc.renewalPeriod, svc.price)
	available := subCoins(sub.vault, due)
	withdraw := chain.NewCoins(chain.NewCoin(denom, amount))
	if lessCoinsThan(available, withdraw) {
		panic("insufficient available balance to withdraw")
	}
	sub.vault = subCoins(sub.vault, withdraw)
	sendCoins(creator, withdraw)
}

// IsSubscribed checks if a user is currently subscribed to a service
// - serviceName: the name of the service to check
// - userAddr: the address of the user to check
// Returns true if the user is subscribed and active, false otherwise
func IsSubscribed(cur realm, serviceName string, userAddr string) bool {
	svc := getServiceByName(serviceName)
	if svc == nil {
		return false
	}
	if !svc.subscribers.Has(userAddr) {
		return false
	}
	subValue, _ := svc.subscribers.Get(userAddr)
	sub := subValue.(*subscription)
	return sub.isActive(svc.renewalPeriod, svc.price)
}

// NewService creates a new subscription service with the given parameters
// - displayName: the name of the service
// - description: a brief description of the service
// - renewalPeriod: the duration between renewals
// - native: the price of the service in native coins
// - grc20: the price of the service in GRC20 tokens
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

// ServiceClaimVault allows the owner of a service to claim the funds in the service's vault
// - displayName: the name of the service
func ServiceClaimVault(cur realm, displayName string) {
	svc := mustGetServiceByName(displayName)
	svc.owner.AssertOwnedByPrevious()
	balance := svc.balance()
	if !coinsHasPositive(balance) {
		panic("no funds to claim")
	}
	sendCoins(svc.owner.Owner(), balance)

}
