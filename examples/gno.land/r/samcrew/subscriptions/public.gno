package subscriptions

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strings"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/seqid"
	"gno.land/r/demo/defi/grc20reg"
	"gno.land/r/samcrew/normalizedcoins"
)

var services *avl.Tree // key: ownerAddr:displayName, value: *service
var users *avl.Tree    // key: user address, value: avl.Tree of subscribed service keys (ownerAddr:displayName) and empty struct{}
var id seqid.ID

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
	if strings.Contains(displayName, serviceKeySeparator) {
		panic("display name cannot contain '" + serviceKeySeparator + "'")
	}
	if description == "" {
		panic("description cannot be empty")
	}
	if renewalPeriod <= 0 {
		panic("renewal period must be positive")
	}

	creator := runtime.PreviousRealm().Address()
	key := serviceKey(creator, displayName)
	if services.Has(key) {
		panic("service with the same display name already exists")
	}
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
	services.Set(key, svc)
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

// Subscribe allows a user to subscribe to a service by providing the service key
// - svcKey: the full service key (ownerAddr:displayName)
// - fqName: the fully qualified name of the GRC20 token to use for payment (empty string for native coins)
// - amount: the amount of GRC20 tokens to deposit initially (0 for native coins)
func Subscribe(cur realm, svcKey string, fqName string, amount int64) {
	if svcKey == "" {
		panic("service key cannot be empty")
	}

	caller := runtime.PreviousRealm().Address()
	svc := mustGetService(svcKey)
	if svc.subscribers.Has(caller.String()) {
		panic("already subscribed")
	}

	// Validate GRC20 token exists if fqName is provided
	var grc20Send chain.Coins
	if fqName != "" && amount != 0 {
		token := grc20reg.Get(fqName)
		if token == nil {
			panic("GRC20 token not found: " + fqName)
		}
		teller := token.RealmTeller()
		if err := teller.TransferFrom(caller, runtime.CurrentRealm().Address(), amount); err != nil {
			panic(err)
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
	svc.subscribers.Set(caller.String(), sub)

	var userSubs *avl.Tree
	userSubsRaw, ok := users.Get(caller.String())
	if !ok {
		userSubs = avl.NewTree()
	} else {
		userSubs = userSubsRaw.(*avl.Tree)
	}
	userSubs.Set(svcKey, struct{}{})
	users.Set(caller.String(), userSubs)
}

// Unsubscribe allows a user to unsubscribe from a service by providing the service key
// It refunds any remaining balance after deducting due amounts to the service owner
// - svcKey: the full service key (ownerAddr:displayName)
func Unsubscribe(cur realm, svcKey string) {
	if svcKey == "" {
		panic("service key cannot be empty")
	}

	caller := runtime.PreviousRealm().Address()
	svc := mustGetService(svcKey)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + svcKey)
	}
	subValue, _ := svc.subscribers.Get(caller.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(true, svc.renewalPeriod, svc.price)
	if normalizedcoins.CoinsHasPositive(due) {
		normalizedcoins.SendCoins(svc.owner.Owner(), due)
	}
	remaining := normalizedcoins.SubCoins(sub.vault, due)
	if normalizedcoins.CoinsHasPositive(remaining) {
		normalizedcoins.SendCoins(caller, remaining)
	}
	svc.subscribers.Remove(caller.String())

	userSubsRaw, _ := users.Get(caller.String())
	userSubs := userSubsRaw.(*avl.Tree)
	userSubs.Remove(svcKey)
	if userSubs.Size() == 0 {
		users.Remove(caller.String())
	}
}

// Topup increases a user's subscription balance using native coins, GRC20 tokens,
// or both in the same transaction. Native funds are inferred from the tx.
// - svcKey: the full service key (ownerAddr:displayName)
// - fqName: GRC20 token denom (optional)
// - amount: amount of GRC20 tokens (ignored if fqName empty)
func Topup(cur realm, svcKey string, fqName string, amount int64) {
	if svcKey == "" {
		panic("service key cannot be empty")
	}

	caller := runtime.PreviousRealm().Address()
	svc := mustGetService(svcKey)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + svcKey)
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
		topupNative(cur, svcKey)
	}
	if amount != 0 {
		topupGRC20(cur, svcKey, fqName, amount)
	}
	// XXX: reset subscription if it was inactive and now has enough balance
	if !active && !normalizedcoins.LessCoinsThan(sub.vault, svc.price) {
		sub.startedAt = time.Now()
		sub.lastChargedAt = time.Now()
		sub.initialPaid = false
	}
}

// Withdraw allows a user to withdraw available balance from their subscription
// - svcKey: the full service key (ownerAddr:displayName)
// - denom: the denomination of the coin to withdraw
// - amount: the amount of the coin to withdraw
func Withdraw(cur realm, svcKey string, denom string, amount int64) {
	if svcKey == "" {
		panic("service key cannot be empty")
	}
	if denom == "" {
		panic("denomination cannot be empty")
	}
	if amount <= 0 {
		panic("withdrawal amount must be positive")
	}

	caller := runtime.PreviousRealm().Address()
	svc := mustGetService(svcKey)
	if !svc.subscribers.Has(caller.String()) {
		panic("not subscribed to service: " + svcKey)
	}
	subValue, _ := svc.subscribers.Get(caller.String())
	sub := subValue.(*subscription)
	due, _ := sub.dueAmount(false, svc.renewalPeriod, svc.price)
	available := normalizedcoins.SubCoins(sub.vault, due)
	withdraw := chain.NewCoins(chain.NewCoin(denom, amount))
	if normalizedcoins.LessCoinsThan(available, withdraw) {
		panic("insufficient available balance to withdraw")
	}
	sub.vault = normalizedcoins.SubCoins(sub.vault, withdraw)
	normalizedcoins.SendCoins(caller, withdraw)
}

// IsSubscribed checks if a user is currently subscribed to a service
// - svcKey: the full service key (ownerAddr:displayName)
// - userAddr: the address of the user to check
// Returns true if the user is subscribed and active, false otherwise
func IsSubscribed(cur realm, svcKey string, userAddr string) bool {
	svc := getService(svcKey)
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

// ServiceClaimVault allows the owner of a service to claim the funds in the service's vault.
// It charges all subscribers for their due amounts and transfers the total to the owner.
// - svcKey: the full service key (ownerAddr:displayName)
func ServiceClaimVault(cur realm, svcKey string) {
	svc := mustGetService(svcKey)
	svc.owner.AssertOwnedByPrevious()
	balance := svc.balance(true)

	if !normalizedcoins.CoinsHasPositive(balance) {
		panic("no funds to claim")
	}
	normalizedcoins.SendCoins(svc.owner.Owner(), balance)
}
