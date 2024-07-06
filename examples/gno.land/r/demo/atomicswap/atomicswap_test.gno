package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var testRun bool

func TestNewCoinSwap_Claim(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestNewCoinSwap_Claim because another test has already run.")
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
	swap := NewCoinSwap(recipient, hashlockHex, timelock)

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	std.TestSetRealm(std.NewUserRealm(recipient))
	uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })
}

func TestNewCoinSwap_Refund(t *testing.T) {
	if testRun {
		t.Skip("Skipping TestNewCoinSwap_Refund because another test has already run.")
	}
	testRun = true

	// Setup
	sender := std.Address("sender")
	recipient := std.Address("recipient")
	amount := std.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	// Create a new swap
	std.TestSetRealm(std.NewUserRealm(sender))
	std.TestSetOrigSend(amount, nil)
	swap := NewCoinSwap(recipient, hashlockHex, timelock) // Create a new swap

	// Test Refund
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	std.TestSetOrigPkgAddr(pkgAddr)
	std.TestIssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")
}

func TestNewGRC20_Claim(t *testing.T) {
	// Setup
	sender := testutils.TestAddress("sender")
	recipient := testutils.TestAddress("recipient")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	mock20Banker := grc20.NewAdminToken("mock", "MOCK", 4)
	mock20Banker.Mint(sender, 100_000)
	mock20Banker.Approve(sender, rlm, 70_000)
	mock20 := mock20Banker.GRC20()

	// Create a new swap
	std.TestSetRealm(std.NewUserRealm(sender))
	swap := NewGRC20Swap(recipient, hashlockHex, timelock, mock20)

	return

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal, _ := mock20.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(40_000))
	// uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	std.TestSetRealm(std.NewUserRealm(recipient))
	uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })
}
