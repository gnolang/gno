package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"time"

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
	sent := std.GetOrigSend()
	require(len(sent) != 0, "at least one coin needs to be sent")
	sendFn := func(to std.Address) {
		banker := std.GetBanker(std.BankerTypeOrigSend)
		banker.SendCoins(std.GetOrigPkgAddr(), to, sent)
	}
	amountStr := sent.String()
	return newSwap(recipient, hashlock, timelock, amountStr, sendFn)
}

func newSwap(
	recipient std.Address,
	hashlock string,
	timelock time.Time,
	amountStr string,
	sendFn func(std.Address),
) *Swap {
	require(time.Now().Before(timelock), "timelock must be in the future")
	require(hashlock != "", "hashlock must not be empty")
	sender := std.PrevRealm().Addr()
	return &Swap{
		recipient: recipient,
		sender:    sender,
		hashlock:  hashlock,
		timelock:  timelock,
		claimed:   false,
		refunded:  false,
		sendFn:    sendFn,
	}
}

// Claim allows the recipient to claim the funds if they provide the correct preimage.
func (s *Swap) Claim(preimage string) {
	require(!s.claimed, "already claimed")
	require(!s.refunded, "already refunded")
	require(std.PrevRealm().Addr() == s.recipient, "unauthorized")
	require(hex.EncodeToString(sha256.Sum256([]byte(preimage))[:]) == s.hashlock, "invalid preimage")

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
