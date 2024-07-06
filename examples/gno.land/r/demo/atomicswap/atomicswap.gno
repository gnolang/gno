package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"time"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
)

// Swap represents an atomic swap contract.
type Swap struct {
	sender    std.Address
	recipient std.Address
	hashlock  string
	timelock  time.Time
	claimed   bool
	refunded  bool
	amountStr string
	sendFn    func(to std.Address)
}

// NewCoinSwap creates a new atomic swap contract for native coins.
func NewCoinSwap(recipient std.Address, hashlock string, timelock time.Time) *Swap {
	sender := std.PrevRealm().Addr()
	sent := std.GetOrigSend()
	require(len(sent) != 0, "at least one coin needs to be sent")
	sendFn := func(to std.Address) {
		banker := std.GetBanker(std.BankerTypeOrigSend)
		banker.SendCoins(std.GetOrigPkgAddr(), to, sent)
	}
	amountStr := sent.String()
	return newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)
}

func NewGRC20Swap(recipient std.Address, hashlock string, timelock time.Time, token grc20.IGRC20) *Swap {
	sender := std.PrevRealm().Addr()
	curAddr := std.CurrentRealm().Addr()

	allowance, err := token.Allowance(sender, curAddr)
	require(err == nil, "allowance error")
	require(allowance > 0, "no allowance")
	err = token.TransferFrom(sender, curAddr, allowance)
	require(err == nil, "cannot retrieve tokens from allowance")
	amountStr := ufmt.Sprintf("%d%s", allowance, token.GetSymbol())
	sendFn := func(to std.Address) {
		err := token.Transfer(to, allowance)
		require(err == nil, "cannot transfer tokens")
	}
	return newSwap(sender, recipient, hashlock, timelock, amountStr, sendFn)
}

func newSwap(
	sender std.Address,
	recipient std.Address,
	hashlock string,
	timelock time.Time,
	amountStr string,
	sendFn func(std.Address),
) *Swap {
	require(time.Now().Before(timelock), "timelock must be in the future")
	require(hashlock != "", "hashlock must not be empty")
	return &Swap{
		recipient: recipient,
		sender:    sender,
		hashlock:  hashlock,
		timelock:  timelock,
		claimed:   false,
		refunded:  false,
		sendFn:    sendFn,
		amountStr: amountStr,
	}
}

// Claim allows the recipient to claim the funds if they provide the correct preimage.
func (s *Swap) Claim(preimage string) {
	require(!s.claimed, "already claimed")
	require(!s.refunded, "already refunded")
	require(std.PrevRealm().Addr() == s.recipient, "unauthorized")

	hashlock := sha256.Sum256([]byte(preimage))
	hashlockHex := hex.EncodeToString(hashlock[:])
	require(hashlockHex == s.hashlock, "invalid preimage")

	s.claimed = true
	s.sendFn(s.recipient)
}

// Refund allows the sender to refund the funds after the timelock has expired.
func (s *Swap) Refund() {
	require(!s.claimed, "already claimed")
	require(!s.refunded, "already refunded")
	require(std.PrevRealm().Addr() == s.sender, "unauthorized")
	require(time.Now().After(s.timelock), "timelock not expired")

	s.refunded = true
	s.sendFn(s.sender)
}

// Render returns the current state of the swap.
func (s *Swap) Render(_ string) string {
	return ufmt.Sprintf(
		"sender=%s\nrecipient=%s\namount=%s\nhashlock=%s\ntimelock=%s\nclaimed=%t\nrefunded=%t",
		s.sender, s.recipient, s.amountStr, s.hashlock, s.timelock, s.claimed, s.refunded,
	)
}

// require checks a condition and panics with a message if the condition is false.
func require(check bool, msg string) {
	if !check {
		panic(msg)
	}
}
