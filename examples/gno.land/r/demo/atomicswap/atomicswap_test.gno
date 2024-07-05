package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var testRun bool

func TestAtomicSwap(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestAtomicSwap because another test has already run.")
	}
	testRun = true

	// Setup
	sender := testutils.TestAddress("sender")
	recipient := testutils.TestAddress("recipient")
	amount := std.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	// Create a new swap
	std.TestSetRealm(std.NewUserRealm(sender))
	std.TestSetOrigSend(amount, nil)
	swap := New(recipient, hashlockHex, timelock)

	/*
		swap := &Swap{
			sender:    sender,
			recipient: recipient,
			amount:    amount,
			hashlock:  hashlockHex,
			timelock:  timelock,
			claimed:   false,
			refunded:  false,
		}
	*/

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	uassert.True(t, swap.amount.IsEqual(amount), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	std.TestSetRealm(std.NewUserRealm(recipient))
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	// Test refund (should fail because already claimed)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, got none")
		}
	}()
	swap.Refund()
}

func TestAtomicSwapRefund(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestAtomicSwapRefund because another test has already run.")
	}
	testRun = true

	// Setup
	sender := std.Address("sender")
	recipient := std.Address("recipient")
	amount := std.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(-1 * time.Hour)

	// Create a new swap
	swap := &Swap{
		sender:    sender,
		recipient: recipient,
		amount:    amount,
		hashlock:  hashlockHex,
		timelock:  timelock,
		claimed:   false,
		refunded:  false,
	}

	std.TestSetRealm(std.NewUserRealm(sender))
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	std.TestSetOrigPkgAddr(pkgAddr)
	std.TestIssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")
}
