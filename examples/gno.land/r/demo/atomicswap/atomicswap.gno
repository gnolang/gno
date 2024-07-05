package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"time"
)

// Swap represents an atomic swap contract.
type Swap struct {
	sender    std.Address
	recipient std.Address
	amount    std.Coins
	hashlock  string
	timelock  time.Time
	claimed   bool
	refunded  bool
}

// New creates a new atomic swap contract.
func New(recipient std.Address, hashlock string, timelock time.Time) *Swap {
	require(time.Now().Before(timelock), "timelock must be in the future")
	return &Swap{
		sender:    std.PrevRealm().Addr(),
		recipient: recipient,
		amount:    std.GetOrigSend(),
		hashlock:  hashlock,
		timelock:  timelock,
		claimed:   false,
		refunded:  false,
	}
}

// Claim allows the recipient to claim the funds if they provide the correct preimage.
func (s *Swap) Claim(preimage []byte) {
	require(!s.claimed, "already claimed")
	require(!s.refunded, "already refunded")
	require(std.GetOrigCaller() == s.recipient, "unauthorized")
	require(hex.EncodeToString(sha256.Sum256(preimage)[:]) == s.hashlock, "invalid preimage")

	s.claimed = true
	std.SendCoins(s.recipient, s.amount)
}

// Refund allows the sender to refund the funds after the timelock has expired.
func (s *Swap) Refund() {
	require(!s.claimed, "already claimed")
	require(!s.refunded, "already refunded")
	require(std.GetOrigCaller() == s.sender, "unauthorized")
	require(time.Now().After(s.timelock), "timelock not expired")

	s.refunded = true
	std.SendCoins(s.sender, s.amount)
}

// require checks a condition and panics with a message if the condition is false.
func require(check bool, msg string) {
	if !check {
		panic(msg)
	}
}

// Render returns the current state of the swap.
func (s *Swap) Render(_ string) string {
	return std.Sprintf(
		"sender=%s\nrecipient=%s\namount=%s\nhashlock=%s\ntimelock=%s\nclaimed=%t\nrefunded=%t",
		s.sender, s.recipient, s.amount.String(), s.hashlock, s.timelock, s.claimed, s.refunded,
	)
}
