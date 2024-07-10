package atomicswap

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/r/demo/tests/test20"
)

var testRun bool

func TestNewCustomCoinSwap_Claim(t *testing.T) {
	defer resetTestState()

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
	id, swap := NewCustomCoinSwap(recipient, hashlockHex, timelock)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

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

	expected = `- status: claimed
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewCustomCoinSwap_Refund(t *testing.T) {
	defer resetTestState()

	if testRun {
		t.Skip("Skipping TestNewCoinSwap_Refund because another test has already run.")
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
	id, swap := NewCustomCoinSwap(recipient, hashlockHex, timelock) // Create a new swap
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test Refund
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	std.TestSetOrigPkgAddr(pkgAddr)
	std.TestIssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")
	expected = `- status: refunded
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-13T22:31:30Z
- remaining: 0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewGRC20_Claim(t *testing.T) {
	defer resetTestState()

	// Setup
	sender := testutils.TestAddress("sender")
	recipient := testutils.TestAddress("recipient")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.Bank.Mint(sender, 100_000)
	test20.Bank.Approve(sender, rlm, 70_000)

	// Create a new swap
	std.TestSetRealm(std.NewUserRealm(sender))
	id, swap := NewCustomGRC20Swap(recipient, hashlockHex, timelock, test20.Bank)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Bank.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Bank.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Bank.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

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

	bal = test20.Bank.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Bank.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Bank.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(70_000))

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })

	expected = `- status: claimed
- sender: g1wdjkuer9wf047h6lta047h6lta047h6ldyx5ax
- recipient: g1wfjkx6tsd9jkuazlta047h6lta047h6lu0zqma
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewGRC20_Refund(t *testing.T) {
	defer resetTestState()

	// Setup
	sender := testutils.TestAddress("sender2")
	recipient := testutils.TestAddress("recipient2")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.Bank.Mint(sender, 100_000)
	test20.Bank.Approve(sender, rlm, 70_000)
	println("AAAAAAAAAAA", "sender", sender, "rlm", rlm)

	// Create a new swap
	std.TestSetRealm(std.NewUserRealm(sender))
	id, swap := NewCustomGRC20Swap(recipient, hashlockHex, timelock, test20.Bank)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wge97h6lta047h6lta047h6ltfacad
- recipient: g1wfjkx6tsd9jkuapjta047h6lta047h6lducc3v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Bank.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Bank.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Bank.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	// Test Refund
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	std.TestSetOrigPkgAddr(pkgAddr)
	std.TestIssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")

	bal = test20.Bank.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(100_000))
	bal = test20.Bank.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Bank.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	expected = `- status: refunded
- sender: g1wdjkuer9wge97h6lta047h6lta047h6ltfacad
- recipient: g1wfjkx6tsd9jkuapjta047h6lta047h6lducc3v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-13T22:31:30Z
- remaining: 0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestRender(t *testing.T) {
	defer resetTestState()

	// Setup
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	charly := testutils.TestAddress("charly")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.Bank.Mint(alice, 100_000)
	std.TestSetRealm(std.NewUserRealm(alice))

	test20.Bank.Approve(alice, rlm, 10_000)
	_, bobSwap := NewCustomGRC20Swap(bob, hashlockHex, timelock, test20.Bank)

	test20.Bank.Approve(alice, rlm, 20_000)
	_, _ = NewCustomGRC20Swap(charly, hashlockHex, timelock, test20.Bank)

	std.TestSetRealm(std.NewUserRealm(bob))
	bobSwap.Claim("secret")

	expected := `- 1: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(10000TST)> g1vfhkyh6lta047h6lta047h6lta047h6l03vdhu - claimed
- 2: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(20000TST)> g1vd5xzunv09047h6lta047h6lta047h6lhsyveh - active
`
	uassert.Equal(t, expected, Render(""))
}

func resetTestState() {
	swaps = avl.Tree{}
	counter = 0
}
