package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
	"gno.land/r/demo/defi/grc20reg"
	"gno.land/r/samcrew/normalizedcoins"
)

var services *avl.Tree // key: displayName, value: *service
var users *avl.Tree    // key: user address, value: avl.Tree of subscribed service names and empty struct{}
var id seqid.ID

// Subscribe allows a user to subscribe to a service by providing the service name
// - serviceName: the name of the service to subscribe to
// - fqName: the fully qualified name of the GRC20 token to use for payment (empty string for native coins)
// - amount: the amount of GRC20 tokens to deposit initially (0 for native coins)
func Subscribe(cur realm, serviceName string, fqName string, amount int64) {
	if serviceName == "" {
		panic("service name cannot be empty")
	}

	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if svc.subscribers.Has(creator.String()) {
		panic("already subscribed")
	}

	// Validate GRC20 token exists if fqName is provided
	var grc20Send chain.Coins
	if fqName != "" && amount != 0 {
		if grc20reg.Get(fqName) == nil {
			panic("GRC20 token not found: " + fqName)
		}
		grc20Send = chain.NewCoins(chain.NewCoin(fqName, amount))
	}

	natSend := banker.OriginSend()
	sCoins, err := normalizedcoins.PrefixCoins(natSend, grc20Send)
	if err != nil {
		panic(err)
	}
	// XXX: support multiple initial GRC20 tokens
	if !normalizedcoins.CoinsHasPositive(sCoins) || normalizedcoins.LessCoinsThan(sCoins, svc.price) {
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
	if serviceName == "" {
		panic("service name cannot be empty")
	}

	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed to service: " + serviceName)
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(true, svc.renewalPeriod, svc.price)
	if normalizedcoins.CoinsHasPositive(due) {
		normalizedcoins.SendCoins(svc.owner.Owner(), due)
	}
	remaining := normalizedcoins.SubCoins(sub.vault, due)
	if normalizedcoins.CoinsHasPositive(remaining) {
		normalizedcoins.SendCoins(creator, remaining)
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
	if serviceName == "" {
		panic("service name cannot be empty")
	}

	caller := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + serviceName)
	}

	// Validate GRC20 token exists if fqName is provided
	if fqName != "" && amount != 0 {
		if grc20reg.Get(fqName) == nil {
			panic("GRC20 token not found: " + fqName)
		}
	}

	subRaw, _ := svc.subscribers.Get(caller.String())
	sub := subRaw.(*subscription)
	_, active := sub.dueAmount(false, svc.renewalPeriod, svc.price)
	native := banker.OriginSend()
	if normalizedcoins.CoinsHasPositive(native) {
		topupNative(cur, serviceName)
	}
	if amount != 0 {
		topupGRC20(cur, serviceName, fqName, amount)
	}
	// XXX: reset subscription if it was inactive and now has enough balance
	if !active && !normalizedcoins.LessCoinsThan(sub.vault, svc.price) {
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
	if serviceName == "" {
		panic("service name cannot be empty")
	}
	if denom == "" {
		panic("denomination cannot be empty")
	}
	if amount <= 0 {
		panic("withdrawal amount must be positive")
	}

	creator := runtime.PreviousRealm().Address()
	svc := mustGetServiceByName(serviceName)
	if !svc.subscribers.Has(creator.String()) {
		panic("not subscribed to service: " + serviceName)
	}
	subValue, _ := svc.subscribers.Get(creator.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(false, svc.renewalPeriod, svc.price)
	available := normalizedcoins.SubCoins(sub.vault, due)
	withdraw := chain.NewCoins(chain.NewCoin(denom, amount))
	if normalizedcoins.LessCoinsThan(available, withdraw) {
		panic("insufficient available balance to withdraw")
	}
	sub.vault = normalizedcoins.SubCoins(sub.vault, withdraw)
	normalizedcoins.SendCoins(creator, withdraw)
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
	if displayName == "" {
		panic("display name cannot be empty")
	}
	if description == "" {
		panic("description cannot be empty")
	}
	if renewalPeriod <= 0 {
		panic("renewal period must be positive")
	}

	// XXX: prefix with creator realm to avoid name clashes?
	if services.Has(displayName) {
		panic("service with the same display name already exists")
	}
	creator := runtime.PreviousRealm().Address()
	sCoins, err := normalizedcoins.PrefixCoins(native, grc20)
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

// NewServiceSimple creates a new subscription service with simple parameters
// This is a helper function that simplifies service creation by accepting
// denomination strings and amounts instead of chain.Coins objects.
// - displayName: the name of the service
// - description: a brief description of the service
// - renewalPeriod: the duration between renewals
// - nativeDenom: the denomination of native coins (e.g., "ugnot"), empty string for none
// - nativeAmount: the amount of native coins, 0 for none
// - grc20Denom: the fully qualified name of GRC20 token (e.g., "gno.land/r/demo/foo20"), empty string for none
// - grc20Amount: the amount of GRC20 tokens, 0 for none
func NewServiceSimple(cur realm, displayName, description string, renewalPeriod time.Duration, nativeDenom string, nativeAmount int64, grc20Denom string, grc20Amount int64) {
	native := chain.Coins{}
	if nativeDenom != "" && nativeAmount > 0 {
		native = chain.NewCoins(chain.NewCoin(nativeDenom, nativeAmount))
	}

	grc20 := chain.Coins{}
	if grc20Denom != "" && grc20Amount > 0 {
		grc20 = chain.NewCoins(chain.NewCoin(grc20Denom, grc20Amount))
	}

	NewService(cur, displayName, description, renewalPeriod, native, grc20)
}

// ServiceClaimVault allows the owner of a service to claim the funds in the service's vault.
// It charges all subscribers for their due amounts and transfers the total to the owner.
// - displayName: the name of the service
func ServiceClaimVault(cur realm, displayName string) {
	svc := mustGetServiceByName(displayName)
	svc.owner.AssertOwnedByPrevious()
	balance := svc.balance(true)

	if !normalizedcoins.CoinsHasPositive(balance) {
		panic("no funds to claim")
	}
	normalizedcoins.SendCoins(svc.owner.Owner(), balance)
}
