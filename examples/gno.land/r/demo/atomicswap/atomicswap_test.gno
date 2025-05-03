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

	// Setup
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	sender := testutils.TestAddress("sender1")
	recipient := testutils.TestAddress("recipient1")
	amount := std.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)
	testing.IssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	testing.SetOriginSend(amount)
	id, swap := cross(NewCustomCoinSwap)(recipient, hashlockHex, timelock)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wgc47h6lta047h6lta047h6l56jtjc
- recipient: g1wfjkx6tsd9jkuap3ta047h6lta047h6lkk20gv
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
	testing.SetRealm(std.NewUserRealm(recipient))
	uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })

	expected = `- status: claimed
- sender: g1wdjkuer9wgc47h6lta047h6lta047h6l56jtjc
- recipient: g1wfjkx6tsd9jkuap3ta047h6lta047h6lkk20gv
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewCustomCoinSwap_Refund(t *testing.T) {
	defer resetTestState()

	// Setup
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	sender := testutils.TestAddress("sender2")
	recipient := testutils.TestAddress("recipient2")
	amount := std.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	testing.SetOriginSend(amount)
	id, swap := cross(NewCustomCoinSwap)(recipient, hashlockHex, timelock) // Create a new swap
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wge97h6lta047h6lta047h6ltfacad
- recipient: g1wfjkx6tsd9jkuapjta047h6lta047h6lducc3v
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test Refund
	testing.IssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")
	expected = `- status: refunded
- sender: g1wdjkuer9wge97h6lta047h6lta047h6ltfacad
- recipient: g1wfjkx6tsd9jkuapjta047h6lta047h6lducc3v
- amount: 1ugnot
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-13T22:31:30Z
- remaining: 0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewCustomGRC20Swap_Claim(t *testing.T) {
	defer resetTestState()

	// Setup
	sender := testutils.TestAddress("sender3")
	recipient := testutils.TestAddress("recipient3")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	id, swap := cross(NewCustomGRC20Swap)(recipient, hashlockHex, timelock, test20.Token)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wge47h6lta047h6lta047h6l5rk38l
- recipient: g1wfjkx6tsd9jkuapnta047h6lta047h6ly6k4pv
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`

	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	// uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	testing.SetRealm(std.NewUserRealm(recipient))
	uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(70_000))

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })

	expected = `- status: claimed
- sender: g1wdjkuer9wge47h6lta047h6lta047h6l5rk38l
- recipient: g1wfjkx6tsd9jkuapnta047h6lta047h6ly6k4pv
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewCustomGRC20Swap_Refund(t *testing.T) {
	defer resetTestState()

	// Setup
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	sender := testutils.TestAddress("sender5")
	recipient := testutils.TestAddress("recipient5")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	id, swap := cross(NewCustomGRC20Swap)(recipient, hashlockHex, timelock, test20.Token)
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wg647h6lta047h6lta047h6l5p6k3k
- recipient: g1wfjkx6tsd9jkuap4ta047h6lta047h6lmwmj6v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-14T00:31:30Z
- remaining: 1h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	// Test Refund
	testing.IssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(100_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	expected = `- status: refunded
- sender: g1wdjkuer9wg647h6lta047h6lta047h6l5p6k3k
- recipient: g1wfjkx6tsd9jkuap4ta047h6lta047h6lmwmj6v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-13T22:31:30Z
- remaining: 0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewGRC20Swap_Claim(t *testing.T) {
	defer resetTestState()

	// Setup
	sender := testutils.TestAddress("sender4")
	recipient := testutils.TestAddress("recipient4")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(defaultTimelockDuration)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	id, swap := cross(NewGRC20Swap)(recipient, hashlockHex, "gno.land/r/demo/tests/test20")
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wg697h6lta047h6lta047h6ltt3lty
- recipient: g1wfjkx6tsd9jkuap5ta047h6lta047h6ljg4l2v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-20T23:31:30Z
- remaining: 168h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	// uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	testing.SetRealm(std.NewUserRealm(recipient))
	uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	swap.Claim("secret")
	uassert.True(t, swap.claimed, "expected claimed to be true")

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(70_000))

	// Test refund (should fail because already claimed)
	uassert.PanicsWithMessage(t, "already claimed", swap.Refund)
	uassert.PanicsWithMessage(t, "already claimed", func() { swap.Claim("secret") })

	expected = `- status: claimed
- sender: g1wdjkuer9wg697h6lta047h6lta047h6ltt3lty
- recipient: g1wfjkx6tsd9jkuap5ta047h6lta047h6ljg4l2v
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-20T23:31:30Z
- remaining: 168h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))
}

func TestNewGRC20Swap_Refund(t *testing.T) {
	defer resetTestState()

	// Setup
	pkgAddr := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	sender := testutils.TestAddress("sender6")
	recipient := testutils.TestAddress("recipient6")
	rlm := std.DerivePkgAddr("gno.land/r/demo/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(defaultTimelockDuration)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(std.NewUserRealm(sender))
	id, swap := cross(NewGRC20Swap)(recipient, hashlockHex, "gno.land/r/demo/tests/test20")
	uassert.Equal(t, 1, id)

	expected := `- status: active
- sender: g1wdjkuer9wgm97h6lta047h6lta047h6ltj497r
- recipient: g1wfjkx6tsd9jkuapkta047h6lta047h6lqyf9rv
- amount: 70000TST
- hashlock: 2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b
- timelock: 2009-02-20T23:31:30Z
- remaining: 168h0m0s`
	uassert.Equal(t, expected, swap.String())
	uassert.Equal(t, expected, Render("1"))

	// Test initial state
	uassert.Equal(t, sender, swap.sender, "expected sender to match")
	uassert.Equal(t, recipient, swap.recipient, "expected recipient to match")
	bal := test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	// Test Refund
	testing.IssueCoins(pkgAddr, std.Coins{{"ugnot", 100000000}})
	uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	swap.Refund()
	uassert.True(t, swap.refunded, "expected refunded to be true")

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, uint64(100_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, uint64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, uint64(0))

	expected = `- status: refunded
- sender: g1wdjkuer9wgm97h6lta047h6lta047h6ltj497r
- recipient: g1wfjkx6tsd9jkuapkta047h6lta047h6lqyf9rv
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

	test20.PrivateLedger.Mint(alice, 100_000)
	testing.SetRealm(std.NewUserRealm(alice))

	userTeller := test20.Token.RealmTeller()
	userTeller.Approve(rlm, 10_000)
	_, bobSwap := cross(NewCustomGRC20Swap)(bob, hashlockHex, timelock, test20.Token)

	userTeller.Approve(rlm, 20_000)
	_, _ = cross(NewCustomGRC20Swap)(charly, hashlockHex, timelock, test20.Token)

	testing.SetRealm(std.NewUserRealm(bob))
	bobSwap.Claim("secret")

	expected := `- 2: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(20000TST)> g1vd5xzunv09047h6lta047h6lta047h6lhsyveh - active
- 1: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(10000TST)> g1vfhkyh6lta047h6lta047h6lta047h6l03vdhu - claimed
`
	uassert.Equal(t, expected, Render(""))
}

func resetTestState() {
	swaps = avl.Tree{}
	counter = 0
}
