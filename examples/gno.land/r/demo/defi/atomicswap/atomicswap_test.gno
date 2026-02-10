package atomicswap

import (
	"chain"
	"chain/runtime"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/r/tests/vm/test20"
)

var testRun bool

func crossThrough(rlm runtime.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}

func TestNewCustomCoinSwap_Claim(cur realm, t *testing.T) {
	defer resetTestState()

	// Setup
	pkgAddr := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	sender := testutils.TestAddress("sender1")
	recipient := testutils.TestAddress("recipient1")
	amount := chain.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)
	testing.IssueCoins(pkgAddr, chain.Coins{{"ugnot", 100000000}})

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	testing.SetOriginSend(amount)
	id, swap := NewCustomCoinSwap(cross, recipient, hashlockHex, timelock)
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
	testing.SetRealm(testing.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	})

	testing.SetRealm(testing.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		swap.Claim("secret")
		uassert.True(t, swap.claimed, "expected claimed to be true")
	})

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

func TestNewCustomCoinSwap_Refund(cur realm, t *testing.T) {
	defer resetTestState()

	// Setup
	pkgAddr := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	sender := testutils.TestAddress("sender2")
	recipient := testutils.TestAddress("recipient2")
	amount := chain.Coins{{Denom: "ugnot", Amount: 1}}
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	testing.SetOriginSend(amount)
	id, swap := NewCustomCoinSwap(cross, recipient, hashlockHex, timelock) // Create a new swap
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
	// testing.SetRealm(runtime.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		testing.IssueCoins(pkgAddr, chain.Coins{{"ugnot", 100000000}})
		uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	})
	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		swap.Refund()
		uassert.True(t, swap.refunded, "expected refunded to be true")
	})
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
	rlm := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	id, swap := NewCustomGRC20Swap(cross, recipient, hashlockHex, timelock, test20.Token)
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
	uassert.Equal(t, bal, int64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

	// uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	testing.SetRealm(testing.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
	})

	testing.SetRealm(testing.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		swap.Claim("secret")
		uassert.True(t, swap.claimed, "expected claimed to be true")
	})

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, int64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(70_000))

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
	pkgAddr := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	sender := testutils.TestAddress("sender5")
	recipient := testutils.TestAddress("recipient5")
	rlm := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	id, swap := NewCustomGRC20Swap(cross, recipient, hashlockHex, timelock, test20.Token)
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
	uassert.Equal(t, bal, int64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

	// Test Refund
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		testing.IssueCoins(pkgAddr, chain.Coins{{"ugnot", 100000000}})
		uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	})

	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		swap.Refund()
		uassert.True(t, swap.refunded, "expected refunded to be true")
	})

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, int64(100_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

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
	rlm := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(defaultTimelockDuration)

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	id, swap := NewGRC20Swap(cross, recipient, hashlockHex, "gno.land/r/tests/vm/test20")
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
	uassert.Equal(t, bal, int64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

	// uassert.Equal(t, swap.amountStr, amount.String(), "expected amount to match")
	uassert.Equal(t, hashlockHex, swap.hashlock, "expected hashlock to match")
	uassert.True(t, swap.timelock.Equal(timelock), "expected timelock to match")
	uassert.False(t, swap.claimed, "expected claimed to be false")
	uassert.False(t, swap.refunded, "expected refunded to be false")

	// Test claim
	testing.SetRealm(testing.NewUserRealm(recipient))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		uassert.PanicsWithMessage(t, "invalid preimage", func() { swap.Claim("invalid") })
		swap.Claim("secret")
		uassert.True(t, swap.claimed, "expected claimed to be true")
	})

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, int64(30_000), bal)
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, int64(0), bal)
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, int64(70_000), bal)

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
	pkgAddr := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	sender := testutils.TestAddress("sender6")
	recipient := testutils.TestAddress("recipient6")
	rlm := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])

	test20.PrivateLedger.Mint(sender, 100_000)
	test20.PrivateLedger.Approve(sender, rlm, 70_000)

	// Create a new swap
	testing.SetRealm(testing.NewUserRealm(sender))
	id, swap := NewGRC20Swap(cross, recipient, hashlockHex, "gno.land/r/tests/vm/test20")
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
	uassert.Equal(t, bal, int64(30_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(70_000))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

	// Test Refund
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		testing.IssueCoins(pkgAddr, chain.Coins{{"ugnot", 100000000}})
		uassert.PanicsWithMessage(t, "timelock not expired", swap.Refund)
	})

	swap.timelock = time.Now().Add(-1 * time.Hour) // override timelock
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		swap.Refund()
		uassert.True(t, swap.refunded, "expected refunded to be true")
	})

	bal = test20.Token.BalanceOf(sender)
	uassert.Equal(t, bal, int64(100_000))
	bal = test20.Token.BalanceOf(rlm)
	uassert.Equal(t, bal, int64(0))
	bal = test20.Token.BalanceOf(recipient)
	uassert.Equal(t, bal, int64(0))

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
	rlm := chain.PackageAddress("gno.land/r/demo/defi/atomicswap")
	hashlock := sha256.Sum256([]byte("secret"))
	hashlockHex := hex.EncodeToString(hashlock[:])
	timelock := time.Now().Add(1 * time.Hour)

	test20.PrivateLedger.Mint(alice, 100_000)
	testing.SetRealm(testing.NewUserRealm(alice))

	userTeller := test20.Token.RealmTeller()
	userTeller.Approve(rlm, 10_000)
	_, bobSwap := NewCustomGRC20Swap(cross, bob, hashlockHex, timelock, test20.Token)

	userTeller.Approve(rlm, 20_000)
	_, _ = NewCustomGRC20Swap(cross, charly, hashlockHex, timelock, test20.Token)

	testing.SetRealm(testing.NewUserRealm(bob))
	crossThrough(testing.NewCodeRealm("gno.land/r/atomicswap/test"), func() {
		bobSwap.Claim("secret")
		expected := `- 2: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(20000TST)> g1vd5xzunv09047h6lta047h6lta047h6lhsyveh - active
- 1: g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh -(10000TST)> g1vfhkyh6lta047h6lta047h6lta047h6l03vdhu - claimed
`
		uassert.Equal(t, expected, Render(""))
	})
}

func resetTestState() {
	swaps = avl.Tree{}
	counter = 0
}
