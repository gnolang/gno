// Package atomicswap implements a hash time-locked contract (HTLC) for atomic swaps
// between native coins (ugnot) or GRC20 tokens.
//
// An atomic swap allows two parties to exchange assets in a trustless way, where
// either both transfers happen or neither does. The process works as follows:
//
//  1. Alice wants to swap with Bob. She generates a secret and creates a swap with
//     Bob's address and the hash of the secret (hashlock).
//
//  2. Bob can claim the assets by providing the correct secret before the timelock expires.
//     The secret proves Bob knows the preimage of the hashlock.
//
// 3. If Bob doesn't claim in time, Alice can refund the assets back to herself.
//
// Example usage for native coins:
//
//	// Alice creates a swap with 1000ugnot for Bob
//	secret := "mysecret"
//	hashlock := hex.EncodeToString(sha256.Sum256([]byte(secret)))
//	id, _ := atomicswap.NewCoinSwap(bobAddr, hashlock) // -send 1000ugnot
//
//	// Bob claims the swap by providing the secret
//	atomicswap.Claim(id, "mysecret")
//
// Example usage for GRC20 tokens:
//
//	// Alice approves the swap contract to spend her tokens
//	token.Approve(swapAddr, 1000)
//
//	// Alice creates a swap with 1000 tokens for Bob
//	id, _ := atomicswap.NewGRC20Swap(bobAddr, hashlock, "gno.land/r/demo/token")
//
//	// Bob claims the swap by providing the secret
//	atomicswap.Claim(id, "mysecret")
//
// If Bob doesn't claim in time (default 1 week), Alice can refund:
//
//	atomicswap.Refund(id)
package atomicswap

import (
	"chain/banker"
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/demo/defi/grc20reg"
)

const defaultTimelockDuration = 7 * 24 * time.Hour // 1w

var (
	swaps   avl.Tree // id -> *Swap
	counter int
)

// NewCoinSwap creates a new atomic swap contract for native coins.
// It uses a default timelock duration.
func NewCoinSwap(cur realm, recipient address, hashlock string) (int, *Swap) {
	timelock := time.Now().Add(defaultTimelockDuration)
	return NewCustomCoinSwap(cur, recipient, hashlock, timelock)
}

// NewGRC20Swap creates a new atomic swap contract for grc20 tokens.
// It uses gno.land/r/demo/defi/grc20reg to lookup for a registered token.
func NewGRC20Swap(cur realm, recipient address, hashlock string, tokenRegistryKey string) (int, *Swap) {
	timelock := time.Now().Add(defaultTimelockDuration)
	token := grc20reg.MustGet(tokenRegistryKey)
	return NewCustomGRC20Swap(cur, recipient, hashlock, timelock, token)
}

// NewCoinSwapWithTimelock creates a new atomic swap contract for native coin.
// It allows specifying a custom timelock duration.
// It is not callable with `gnokey maketx call`, but can be imported by another contract or `gnokey maketx run`.
func NewCustomCoinSwap(cur realm, recipient address, hashlock string, timelock time.Time) (int, *Swap) {
	sender := runtime.PreviousRealm().Address()
	sent := banker.OriginSend()
	require(len(sent) != 0, "at least one coin needs to be sent")

	// Create the swap
	sendFn := func(cur realm, to address) {
		banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
		pkgAddr := runtime.CurrentRealm().Address()
		banker_.SendCoins(pkgAddr, to, sent)
	}
	amountStr := sent.String()
	swap := newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)

	counter++
	id := strconv.Itoa(counter)
	swaps.Set(id, swap)
	return counter, swap
}

// NewCustomGRC20Swap creates a new atomic swap contract for grc20 tokens.
// It is not callable with `gnokey maketx call`, but can be imported by another contract or `gnokey maketx run`.
func NewCustomGRC20Swap(cur realm, recipient address, hashlock string, timelock time.Time, token *grc20.Token) (int, *Swap) {
	sender := runtime.PreviousRealm().Address()
	curAddr := runtime.CurrentRealm().Address()

	allowance := token.Allowance(sender, curAddr)
	require(allowance > 0, "no allowance")

	userTeller := token.RealmTeller()
	err := userTeller.TransferFrom(sender, curAddr, allowance)
	require(err == nil, "cannot retrieve tokens from allowance")

	amountStr := ufmt.Sprintf("%d%s", allowance, token.GetSymbol())
	sendFn := func(cur realm, to address) {
		err := userTeller.Transfer(to, allowance)
		require(err == nil, "cannot transfer tokens")
	}

	swap := newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)

	counter++
	id := strconv.Itoa(counter)
	swaps.Set(id, swap)

	return counter, swap
}

// Claim loads a registered swap and tries to claim it.
func Claim(cur realm, id int, secret string) {
	swap := mustGet(id)
	swap.Claim(secret)
}

// Refund loads a registered swap and tries to refund it.
func Refund(cur realm, id int) {
	swap := mustGet(id)
	swap.Refund()
}

// Render returns a list of swaps (simplified) for the homepage, and swap details when specifying a swap ID.
func Render(path string) string {
	if path == "" { // home
		output := ""
		size := swaps.Size()
		max := 10
		swaps.ReverseIterateByOffset(size-max, max, func(key string, value any) bool {
			swap := value.(*Swap)
			output += ufmt.Sprintf("- %s: %s -(%s)> %s - %s\n",
				key, swap.sender, swap.amountStr, swap.recipient, swap.Status())
			return false
		})
		return output
	} else { // by id
		swap, ok := swaps.Get(path)
		if !ok {
			return "404"
		}
		return swap.(*Swap).String()
	}
}

// require checks a condition and panics with a message if the condition is false.
func require(check bool, msg string) {
	if !check {
		panic(msg)
	}
}

// mustGet retrieves a swap by its id or panics.
func mustGet(id int) *Swap {
	key := strconv.Itoa(id)
	swap, ok := swaps.Get(key)
	if !ok {
		panic("unknown swap ID")
	}
	return swap.(*Swap)
}
