package atomicswap

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

const defaultTimelockDuration = 7 * 24 * time.Hour // 1w

var (
	swaps   avl.Tree // id -> *Swap
	counter int
)

// NewCoinSwap creates a new atomic swap contract for native coins.
// It uses a default timelock duration.
func NewCoinSwap(recipient std.Address, hashlock string) (int, *Swap) {
	timelock := time.Now().Add(defaultTimelockDuration)
	return NewCustomCoinSwap(recipient, hashlock, timelock)
}

// NewGRC20Swap creates a new atomic swap contract for grc20 tokens.
// It uses gno.land/r/demo/grc20reg to lookup for a registered token.
func NewGRC20Swap(recipient std.Address, hashlock string, tokenRegistryKey string) (int, *Swap) {
	timelock := time.Now().Add(defaultTimelockDuration)
	panic("not implemented (depends on #2516)")
	bank := grc20reg.Get(tokenRegistryKey)
	if bank == nil {
		panic("no such token")
	}
	return NewCustomGRC20Swap(recipient, hashlock, timelock, bank)
}

// NewCoinSwapWithTimelock creates a new atomic swap contract for native coin.
// It allows specifying a custom timelock duration.
// It is not callable with `gnokey maketx call`, but can be imported by another contract or `gnokey maketx run`.
func NewCustomCoinSwap(recipient std.Address, hashlock string, timelock time.Time) (int, *Swap) {
	sender := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	require(len(sent) != 0, "at least one coin needs to be sent")
	sendFn := func(to std.Address) {
		banker := std.GetBanker(std.BankerTypeRealmSend)
		banker.SendCoins(std.GetOrigPkgAddr(), to, sent)
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
func NewCustomGRC20Swap(recipient std.Address, hashlock string, timelock time.Time, bank *grc20.Bank) (int, *Swap) {
	sender := std.PrevRealm().Addr()
	curAddr := std.CurrentRealm().Addr()
	userBanker := grc20.PrevRealmBanker(bank)

	allowance := userBanker.Allowance(sender, curAddr)
	require(allowance > 0, "no allowance")
	err := userBanker.TransferFrom(sender, curAddr, allowance)
	require(err == nil, "cannot retrieve tokens from allowance")
	amountStr := ufmt.Sprintf("%d%s", allowance, userBanker.GetSymbol())
	sendFn := func(to std.Address) {
		err := userBanker.Transfer(to, allowance)
		require(err == nil, "cannot transfer tokens")
	}
	swap := newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)
	counter++
	id := strconv.Itoa(counter)
	swaps.Set(id, swap)
	return counter, swap
}

// Claim loads a registered swap and tries to claim it.
func Claim(id int, secret string) {
	swap := mustGet(id)
	swap.Claim(secret)
}

// Refund loads a registered swap and tries to refund it.
func Refund(id int) {
	swap := mustGet(id)
	swap.Refund()
}

// Render returns a list of swaps (simplified) for the homepage, and swap details when specifying a swap ID.
func Render(path string) string {
	if path == "" { // home
		output := ""
		size := swaps.Size()
		max := 10
		swaps.ReverseIterateByOffset(size-max, max, func(key string, value interface{}) bool {
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

// mustGet retrieves a swap by its id or panic.
func mustGet(id int) *Swap {
	key := strconv.Itoa(id)
	swap, ok := swaps.Get(key)
	if !ok {
		panic("unknown swap ID")
	}
	return swap.(*Swap)
}
