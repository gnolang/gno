package multisig

import (
	"chain"
	"chain/banker"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/grc20/v0"
	msig "gno.land/p/nt/multisig/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"

	"gno.land/r/nt/grc20reg/v0"
)

func TestTwoSafesStayIsolated(cur realm, t *testing.T) {
	setTime(t0)
	tok, led := grc20.NewToken("Isolation", "ISO", 0, 1, cur)
	key := grc20reg.Register(cross(cur), tok, "")
	sink := testutils.TestAddress("iso-sink")

	at(t0, alice)
	var a, b, prop int64
	func(cur realm) {
		signers := alice.String()
		a = Create(cur, "safe-a", signers, 1, 0)
		b = Create(cur, "safe-b", signers, 1, 0)
		testing.IssueCoins(SafeAddress(a), chain.NewCoins(chain.NewCoin("ugnot", 10)))
		testing.IssueCoins(SafeAddress(b), chain.NewCoins(chain.NewCoin("ugnot", 10)))
		if err := led.Mint(SafeAddress(a), 4); err != nil {
			panic(err)
		}
		if err := led.Mint(SafeAddress(b), 4); err != nil {
			panic(err)
		}
		prop = ProposeCoin(cur, a, sink, "ugnot", 10, t0+1000, "")
		Approve(cur, a, prop)
		Execute(cur, a, prop)
		mustPanic(t, "not pending", func() { Execute(cur, a, prop) })
		gprop := ProposeGRC20(cur, a, sink, key, 4, t0+1000, "")
		Approve(cur, a, gprop)
		Execute(cur, a, gprop)
	}(cross(cur))

	uassert.Equal(t, int64(0), Balance(a, "ugnot"))
	uassert.Equal(t, int64(10), Balance(b, "ugnot"))
	uassert.Equal(t, int64(0), tok.BalanceOf(SafeAddress(a)))
	uassert.Equal(t, int64(4), tok.BalanceOf(SafeAddress(b)))
	uassert.Equal(t, int64(4), tok.BalanceOf(sink))
	if SafeAddress(a) == SafeAddress(b) {
		t.Error("safes share an address")
	}
}

func TestBadDenomUnknownTokenAndRender(cur realm, t *testing.T) {
	at(t0, alice)
	var id int64
	func(cur realm) {
		id = Create(cur, "page](http://evil)", alice.String(), 1, 0)
		mustPanic(t, "invalid denom", func() { WatchDenom(cur, id, "NO") })
		mustPanic(t, "invalid denom", func() { ProposeCoin(cur, id, dest, "NO", 1, t0+1000, "") })
		mustPanic(t, "invalid amount", func() { ProposeCoin(cur, id, dest, "ugnot", 0, t0+1000, "") })
		mustPanic(t, "invalid token", func() {
			prop := ProposeGRC20(cur, id, dest, "gno.land/r/demo/missing.NOPE", 1, t0+1000, "")
			Approve(cur, id, prop)
			Execute(cur, id, prop)
		})
		prop := ProposeGRC20(cur, id, dest, "gno.land/r/demo/missing.NOPE", 1, t0+1000, "memo*bold")
		uassert.Equal(t, "pending", ProposalStatus(id, prop))
		s := mustSafe(id)
		mustPanic(t, "invalid amount", func() {
			applyCoin(cur, s, msig.Effect{Kind: msig.KindCoin, Amount: 0, Denom: "ugnot", To: dest})
		})
		mustPanic(t, "invalid denom", func() {
			applyCoin(cur, s, msig.Effect{Kind: msig.KindCoin, Amount: 1, Denom: "NO", To: dest})
		})
		err := canAfford(s, msig.Effect{Kind: msig.KindCoin, Amount: 1, Denom: "NO"})
		if err == nil || !strings.Contains(err.Error(), "invalid denom") {
			t.Errorf("canAfford denom: %v", err)
		}
	}(cross(cur))

	page := Render(strconv.FormatInt(id, 10))
	if strings.Contains(page, "](http://evil)") {
		t.Errorf("render broke out of the link:\n%s", page)
	}
	if !strings.Contains(page, "active") {
		t.Error("safe page missing status")
	}
	home := Render("")
	if strings.Contains(home, "](http://evil)") {
		t.Errorf("home broke out of the link:\n%s", home)
	}
	if Render("nope") == "" || Render("0") == "" || !strings.Contains(Render("9223372036854775807"), "Unknown safe.") || Render(strconv.FormatInt(id, 10)+"/nope") == "" || Render(strconv.FormatInt(id, 10)+"/1/2") == "" {
		t.Error("unknown paths should still render")
	}
	mustPanic(t, "invalid denom", func() { Balance(id, "NO") })
	mustPanic(t, "unknown safe", func() { SafeName(0) })
	mustPanic(t, "unknown safe", func() { SafeName(9223372036854775807) })
}

func TestDepositCallerAndDenom(cur realm, t *testing.T) {
	saved := testing.GetContext()
	defer testing.SetContext(saved)

	at(t0, alice)
	var id int64
	func(cur realm) {
		id = Create(cur, "dep", alice.String(), 1, 0)
	}(cross(cur))

	factory := chain.PackageAddress(realmPath)
	// Empty -send envelope is rejected.
	testing.SetOriginSend(chain.Coins{})
	at(t0, alice)
	func(cur realm) {
		mustPanic(t, "invalid amount", func() { Deposit(cur, id) })
	}(cross(cur))

	// Zero amount and malformed denom in -send are rejected.
	testing.SetOriginSend(chain.Coins{chain.Coin{Denom: "ugnot", Amount: 0}})
	at(t0, alice)
	func(cur realm) {
		mustPanic(t, "invalid amount", func() { Deposit(cur, id) })
	}(cross(cur))

	testing.SetOriginSend(chain.Coins{chain.Coin{Denom: "NO", Amount: 1}})
	at(t0, alice)
	func(cur realm) {
		mustPanic(t, "invalid denom", func() { Deposit(cur, id) })
	}(cross(cur))

	// Full multi-denom -send envelope is forwarded, leaving nothing stranded on factory.
	multi := chain.NewCoins(chain.NewCoin("foo", 3), chain.NewCoin("ugnot", 7))
	testing.IssueCoins(factory, multi)
	testing.SetOriginSend(multi)
	at(t0, alice)
	func(cur realm) {
		Deposit(cur, id)
	}(cross(cur))
	uassert.Equal(t, int64(7), Balance(id, "ugnot"))
	uassert.Equal(t, int64(3), Balance(id, "foo"))
	uassert.Equal(t, int64(0), banker.NewReadonlyBanker().GetCoin(factory, "ugnot"))
	uassert.Equal(t, int64(0), banker.NewReadonlyBanker().GetCoin(factory, "foo"))

	// Code-realm caller is rejected.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/evil"))
	caught := revive(func() { Deposit(cross(cur), id) })
	if !strings.Contains(panicText(caught), "direct user call") {
		t.Errorf("deposit caller: %v", caught)
	}
}

func TestSafeCap(cur realm, t *testing.T) {
	saved := safes
	safes = make([]*msig.Safe, maxSafes)
	defer func() { safes = saved }()
	at(t0, alice)
	func(cur realm) {
		mustPanic(t, "too many safes", func() {
			Create(cur, "overflow", alice.String(), 1, 0)
		})
	}(cross(cur))
}

func TestKindsPagesAndQueries(cur realm, t *testing.T) {
	at(t0, alice)
	var id int64
	worker := "gno.land/r/demo/worker"
	func(cur realm) {
		id = Create(cur, "kinds", alice.String()+","+bob.String(), 1, 0)
		testing.IssueCoins(SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 3)))
		coin := ProposeCoin(cur, id, dest, "ugnot", 3, t0+5000, "coin")
		call := ProposeCall(cur, id, worker, "payload", t0+5000, "call")
		add := ProposeAddSigner(cur, id, carol, t0+5000, "add")
		rm := ProposeRemoveSigner(cur, id, bob, t0+5000, "rm")
		th := ProposeThreshold(cur, id, 1, t0+5000, "th")
		dl := ProposeDelay(cur, id, 0, t0+5000, "dl")
		pz := ProposePause(cur, id, t0+5000, "pz")
		for _, p := range []int64{coin, call, add, rm, th, dl, pz} {
			Approve(cur, id, p)
		}
		Execute(cur, id, pz)
		uassert.True(t, SafePaused(id))
		mustPanic(t, "paused", func() { Execute(cur, id, coin) })
		un := ProposeUnpause(cur, id, t0+5000, "un")
		Approve(cur, id, un)
		Execute(cur, id, un)
		Execute(cur, id, coin)
		mustPanic(t, "not pending", func() { Execute(cur, id, coin) })
		Execute(cur, id, add)
		Execute(cur, id, rm)
		Execute(cur, id, th)
		Execute(cur, id, dl)
		Execute(cur, id, call)
		uassert.False(t, IsSigner(id, bob))
		uassert.True(t, IsSigner(id, carol))
		uassert.Equal(t, 1, SafeThreshold(id))
		uassert.Equal(t, int64(0), SafeDelay(id))
		uassert.False(t, SafePaused(id))
	}(cross(cur))

	uassert.Equal(t, "kinds", SafeName(id))
	uassert.True(t, SafeAddress(id).IsValid())
	if SignerCount(id) < 1 || SignerAt(id, 0) == "" {
		t.Error("signers")
	}
	if ProposalCount(id) < 1 || SafeNonce(id) < 1 {
		t.Error("counts")
	}
	page := Render(strconv.FormatInt(id, 10))
	if !strings.Contains(page, "coin") || !strings.Contains(page, "call") {
		t.Errorf("safe page:\n%s", page)
	}
	for n := int64(1); n <= int64(ProposalCount(id)); n++ {
		body := Render(strconv.FormatInt(id, 10) + "/" + strconv.FormatInt(n, 10))
		if strings.Contains(body, "Unknown proposal") {
			t.Errorf("missing proposal %d", n)
		}
	}
	if !strings.Contains(Render(strconv.FormatInt(id, 10)+"/1"), "Send ") {
		t.Error("coin page")
	}
}

func TestFundsConserved(cur realm, t *testing.T) {
	sink := testutils.TestAddress("prop-sink")
	before := banker.NewReadonlyBanker().GetCoin(sink, "ugnot")
	at(t0, alice)
	var id int64
	func(cur realm) {
		id = Create(cur, "conserved", alice.String(), 1, 0)
		testing.IssueCoins(SafeAddress(id), chain.NewCoins(chain.NewCoin("ugnot", 1000)))
	}(cross(cur))
	seed := uint64(99)
	left := int64(1000)
	moved := int64(0)
	for step := 0; step < 12; step++ {
		if left == 0 {
			break
		}
		seed = seed*1664525 + 1013904223
		amount := int64(seed%20) + 1
		if amount > left {
			amount = left
		}
		now := t0 + int64(step)
		at(now, alice)
		func(cur realm) {
			prop := ProposeCoin(cur, id, sink, "ugnot", amount, now+1000, "")
			Approve(cur, id, prop)
			Execute(cur, id, prop)
			mustPanic(t, "not pending", func() { Execute(cur, id, prop) })
		}(cross(cur))
		left -= amount
		moved += amount
		if Balance(id, "ugnot") != left {
			t.Errorf("step %d safe balance", step)
		}
		got := banker.NewReadonlyBanker().GetCoin(sink, "ugnot") - before
		if got != moved {
			t.Errorf("step %d sink", step)
		}
	}
	if moved == 0 || Balance(id, "ugnot")+moved != 1000 {
		t.Error("funds were not conserved")
	}
}

func TestRenderAndWrappers(cur realm, t *testing.T) {
	saved := testing.GetContext()
	defer testing.SetContext(saved)

	at(t0, alice)
	for SafeCount() < 21 {
		Create(cross(cur), "extra", alice.String(), 1, 0)
	}
	home := Render("")
	if !strings.Contains(home, "Showing the latest 20") {
		t.Error("home should truncate")
	}

	id := Create(cross(cur), "pages", alice.String()+","+bob.String(), 1, 0)
	sid := strconv.FormatInt(id, 10)
	if !strings.Contains(Render(sid), "None yet") {
		t.Error("empty proposal list")
	}
	if !strings.Contains(Render(sid+"/99"), "Unknown proposal") {
		t.Error("unknown proposal page")
	}

	coin := ProposeCoin(cross(cur), id, dest, "ugnot", 1, t0+10_000, "")
	page := Render(sid + "/" + strconv.FormatInt(coin, 10))
	if !strings.Contains(page, "None.") {
		t.Error("proposal with no approvals")
	}
	Approve(cross(cur), id, coin)
	Revoke(cross(cur), id, coin)
	Cancel(cross(cur), id, coin)

	// Watching "ugnot" must not duplicate the default ugnot balance row, and
	// inline code paths must not contain backslash-escaped dots.
	WatchDenom(cross(cur), id, "ugnot")
	WatchToken(cross(cur), id, "gno.land/r/demo/missing.NO")
	watched := Render(sid)
	if strings.Count(watched, "- ugnot:") != 1 {
		t.Errorf("ugnot row duplicated:\n%s", watched)
	}
	if !strings.Contains(watched, "`gno.land/r/demo/missing.NO`") || !strings.Contains(watched, "not in the registry") {
		t.Errorf("watched token row malformed:\n%s", watched)
	}
	UnwatchDenom(cross(cur), id, "ugnot")
	UnwatchToken(cross(cur), id, "gno.land/r/demo/missing.NO")

	// Self-removal preserves the removed signer's approval with "(former signer)".
	rmBob := ProposeRemoveSigner(cross(cur), id, bob, t0+10_000, "")
	at(t0, bob)
	Approve(cross(cur), id, rmBob)
	Execute(cross(cur), id, rmBob)
	rmPage := Render(sid + "/" + strconv.FormatInt(rmBob, 10))
	if !strings.Contains(rmPage, bob.String()) || !strings.Contains(rmPage, "former signer") {
		t.Errorf("self-removal page missing former signer audit entry:\n%s", rmPage)
	}

	at(t0, alice)
	pause := ProposePause(cross(cur), id, t0+10_000, "")
	Approve(cross(cur), id, pause)
	Execute(cross(cur), id, pause)
	if !strings.Contains(Render(sid), "paused") {
		t.Error("paused page")
	}
	unpause := ProposeUnpause(cross(cur), id, t0+10_000, "")
	Approve(cross(cur), id, unpause)
	Execute(cross(cur), id, unpause)

	busy := Create(cross(cur), "busy", alice.String(), 1, 0)
	for i := 0; i < 21; i++ {
		ProposeCoin(cross(cur), busy, dest, "ugnot", 1, t0+10_000, "")
	}
	if !strings.Contains(Render(strconv.FormatInt(busy, 10)), "Showing the latest") {
		t.Error("proposal list should truncate")
	}

	call := ProposeCall(cross(cur), id, "gno.land/r/demo/worker", "secret", t0+10_000, "memo")
	Approve(cross(cur), id, call)
	Execute(cross(cur), id, call)
	callPage := Render(sid + "/" + strconv.FormatInt(call, 10))
	if !strings.Contains(callPage, "`gno.land/r/demo/worker`") {
		t.Errorf("call target code span corrupted:\n%s", callPage)
	}
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/worker"))
	if Consume(cross(cur), id, call) != "secret" {
		t.Error("consume")
	}
	if !strings.Contains(Render(sid+"/"+strconv.FormatInt(call, 10)), "already consumed") {
		t.Error("consumed page")
	}
}
