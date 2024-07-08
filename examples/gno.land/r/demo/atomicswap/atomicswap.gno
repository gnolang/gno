package atomicswap

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
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
	var token grc20.Token
	// token := grc20reg.Get(tokenRegistryKey)
	panic("not implemented")
	return NewCustomGRC20Swap(recipient, hashlock, timelock, token)
}

// NewCoinSwapWithTimelock creates a new atomic swap contract for native coin.
// It allows specifying a custom timelock duration.
// It is not callable with `gnokey maketx call`, but can be imported by another contract or `gnokey maketx run`.
func NewCustomCoinSwap(recipient std.Address, hashlock string, timelock time.Time) (int, *Swap) {
	sender := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	require(len(sent) != 0, "at least one coin needs to be sent")
	sendFn := func(to std.Address) {
		banker := std.GetBanker(std.BankerTypeOrigSend)
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
func NewCustomGRC20Swap(recipient std.Address, hashlock string, timelock time.Time, token grc20.Token) (int, *Swap) {
	sender := std.PrevRealm().Addr()
	curAddr := std.CurrentRealm().Addr()

	allowance := token.Allowance(sender, curAddr)
	require(allowance > 0, "no allowance")
	err := token.TransferFrom(sender, curAddr, allowance)
	require(err == nil, "cannot retrieve tokens from allowance")
	amountStr := ufmt.Sprintf("%d%s", allowance, token.GetSymbol())
	sendFn := func(to std.Address) {
		err := token.Transfer(to, allowance)
		require(err == nil, "cannot transfer tokens")
	}
	swap := newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)
	counter++
	id := strconv.Itoa(counter)
	swaps.Set(id, swap)
	return counter, swap
}

func Render(path string) string {
	if path == "" { // home
		output := ""
		size := swaps.Size()
		max := 10
		swaps.ReverseIterateByOffset(size-max, max, func(key string, value interface{}) bool {
			swap := value.(*Swap)
			output += ufmt.Sprintf("- %s: %s -(%s)> %s - %s\n",
				key, swap.sender, swap.amountStr, swap.recipient, swap.Status())
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
