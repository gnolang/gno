package multisig

import (
	"chain"
	"math"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func as(user address, rlm realm, fn func(cur realm)) {
	ctx := testing.GetContext()
	ctx.CurrentRealm = testing.NewUserRealm(user)
	ctx.OriginCaller = user
	testing.SetContext(ctx)
	fn(cross(rlm))
}

func asCode(pkgPath string, rlm realm, fn func(cur realm)) {
	ctx := testing.GetContext()
	ctx.CurrentRealm = testing.NewCodeRealm(pkgPath)
	ctx.OriginCaller = alice
	testing.SetContext(ctx)
	fn(cross(rlm))
}

func pick(seed uint64, n int) int {
	if n <= 0 {
		return 0
	}
	return int(seed % uint64(n))
}

func assertAdd64(t *testing.T, a, b int64, wantOK bool) {
	t.Helper()
	got, ok := add64(a, b)
	if ok != wantOK {
		t.Errorf("add64(%d, %d) ok=%v", a, b, ok)
	}
	if !wantOK && got != 0 {
		t.Errorf("add64(%d, %d) = %d", a, b, got)
	}
}

func wantErr(t *testing.T, err, want error) {
	t.Helper()
	if want == nil {
		urequire.NoError(t, err)
		return
	}
	uassert.ErrorContains(t, err, want.Error())
}

func signerCSV(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(testutils.TestAddress("s" + strconv.Itoa(i)).String())
	}
	return b.String()
}

func signerN(i int) address {
	return testutils.TestAddress("s" + strconv.Itoa(i))
}

func TestNewBoundaries(cur realm, t *testing.T) {
	pkg, addr := host()
	ok := alice.String() + "," + bob.String()
	cases := []struct {
		hostPkg  string
		hostAddr address
		creator  address
		name     string
		signers  string
		th       int
		delay    int64
		want     error
	}{
		{pkg, addr, alice, "treasury", ok, 1, 0, nil},
		{pkg, addr, alice, "treasury", ok, 2, MaxDelay, nil},
		{"gno.land/p/nt/multisig/v0", addr, alice, "treasury", ok, 1, 0, ErrHost},
		{"gno.land/r/nt/multisig/v0#s", addr, alice, "treasury", ok, 1, 0, ErrHost},
		{pkg, alice, alice, "treasury", ok, 1, 0, ErrHost},
		{pkg, addr, "", "treasury", ok, 1, 0, ErrInvalidAddress},
		{pkg, addr, alice, "", ok, 1, 0, ErrName},
		{pkg, addr, alice, strings.Repeat("n", MaxNameLen+1), ok, 1, 0, ErrName},
		{pkg, addr, alice, "bad\nname", ok, 1, 0, ErrName},
		{pkg, addr, alice, "treasury", "", 1, 0, ErrInvalidAddress},
		{pkg, addr, alice, "treasury", "not-an-address", 1, 0, ErrInvalidAddress},
		{pkg, addr, alice, "treasury", alice.String() + "," + alice.String(), 1, 0, ErrDuplicateSigner},
		{pkg, addr, alice, "treasury", " " + alice.String() + " , " + bob.String(), 1, 0, nil},
		{pkg, addr, alice, "treasury", ok, 0, 0, ErrThreshold},
		{pkg, addr, alice, "treasury", ok, 3, 0, ErrThreshold},
		{pkg, addr, alice, "treasury", ok, 1, -1, ErrDelay},
		{pkg, addr, alice, "treasury", ok, 1, MaxDelay + 1, ErrDelay},
		{pkg, addr, alice, "treasury", addr.String(), 1, 0, ErrSelf},
		{pkg, addr, alice, "treasury", signerCSV(MaxSigners), 1, 0, nil},
		{pkg, addr, alice, "treasury", signerCSV(MaxSigners + 1), 1, 0, ErrSignerCount},
	}
	for _, tc := range cases {
		s, err := New(tc.hostPkg, tc.hostAddr, tc.creator, tc.name, tc.signers, tc.th, tc.delay)
		wantErr(t, err, tc.want)
		if tc.want == nil && s == nil {
			t.Errorf("New returned nil")
		}
	}
	uassert.Equal(t, "unknown", Kind(0).String())
	uassert.Equal(t, "unknown", Status(9).String())
	uassert.False(t, Spend(KindPause))
	uassert.True(t, Spend(KindCoin) && Spend(KindGRC20) && Spend(KindCall))
	uassert.Equal(t, StatusExecuted.String(), DisplayStatus(StatusExecuted, 1, 99).String())
	uassert.Equal(t, StatusExpired.String(), DisplayStatus(StatusPending, 10, 10).String())
	assertAdd64(t, math.MaxInt64, 1, false)
	assertAdd64(t, -1, 1, false)
	assertAdd64(t, 2, 3, true)
	uassert.True(t, ValidDenom("ugnot"))
	uassert.True(t, ValidDenom("/gno.land/r/nt/foo:bar"))
	uassert.False(t, ValidDenom("NO"))
	uassert.False(t, ValidDenom("ab"))
	uassert.False(t, ValidDenom(""))
}

func TestBindBoundaries(cur realm, t *testing.T) {
	pkg, addr := host()
	var none *Safe
	_, err := none.Bind(1)
	uassert.ErrorContains(t, err, ErrUnbound.Error())

	signers := alice.String() + "," + bob.String()
	s, err := New(pkg, addr, alice, "treasury", signers, 1, 0)
	urequire.NoError(t, err)
	_, err = s.Bind(0)
	uassert.ErrorContains(t, err, ErrHost.Error())
	_, err = s.Bind(1)
	urequire.NoError(t, err)
	uassert.Equal(t, "s/1", s.Subpath())
	uassert.Equal(t, chain.DerivePkgSubAddr(pkg, "s/1").String(), s.Address().String())
	_, err = s.Bind(2)
	uassert.ErrorContains(t, err, ErrBound.Error())

	sub := chain.DerivePkgSubAddr(pkg, "s/7")
	clash, err := New(pkg, addr, alice, "clash", sub.String(), 1, 0)
	urequire.NoError(t, err)
	_, err = clash.Bind(7)
	uassert.ErrorContains(t, err, ErrSelf.Error())

	raw, err := New(pkg, addr, alice, "raw", signers, 1, 0)
	urequire.NoError(t, err)
	as(alice, cur, func(cur realm) {
		_, err = raw.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrUnbound.Error())
		_, err = raw.Execute(0, cur, 1, t0)
		uassert.ErrorContains(t, err, ErrUnbound.Error())
	})
	err = none.Approve(0, cur, 1, t0)
	uassert.ErrorContains(t, err, ErrUnbound.Error())
}

func TestProposalKindsAndErrors(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	worker := "gno.land/r/demo/worker"
	hostPkg, _ := host()
	as(dave, cur, func(cur realm) {
		_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrNotSigner.Error())
	})
	stale := testing.MakeRealm(dave, "gno.land/r/demo/stale", testing.OriginRealm())
	_, err := s.ProposeCoin(0, stale, dest, "ugnot", 1, t0+100, "", t0)
	uassert.ErrorContains(t, err, ErrStaleRealm.Error())

	type row struct {
		name string
		run  func(cur realm) error
		want error
	}
	rows := []row{
		{"coin", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "pay", t0)
			return err
		}, nil},
		{"zero amount", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 0, t0+100, "", t0)
			return err
		}, ErrAmount},
		{"negative amount", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", -1, t0+100, "", t0)
			return err
		}, ErrAmount},
		{"bad denom", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "NO", 1, t0+100, "", t0)
			return err
		}, ErrDenom},
		{"short denom", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ab", 1, t0+100, "", t0)
			return err
		}, ErrDenom},
		{"self safe", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, s.Address(), "ugnot", 1, t0+100, "", t0)
			return err
		}, ErrSelf},
		{"self host", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, s.hostAddr, "ugnot", 1, t0+100, "", t0)
			return err
		}, ErrSelf},
		{"bad to", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, "", "ugnot", 1, t0+100, "", t0)
			return err
		}, ErrInvalidAddress},
		{"expiry equals ready", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0, "", t0)
			return err
		}, ErrExpiry},
		{"expiry max", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+MaxLifetime, "m", t0)
			return err
		}, nil},
		{"expiry past max", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+MaxLifetime+1, "", t0)
			return err
		}, ErrExpiry},
		{"memo long", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, strings.Repeat("m", MaxMemoLen+1), t0)
			return err
		}, ErrMemo},
		{"memo control", func(cur realm) error {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "a\nb", t0)
			return err
		}, ErrMemo},
		{"grc20", func(cur realm) error {
			_, err := s.ProposeGRC20(0, cur, dest, "gno.land/r/demo/tok.TOK", 2, t0+100, "", t0)
			return err
		}, nil},
		{"grc20 bad key", func(cur realm) error {
			_, err := s.ProposeGRC20(0, cur, dest, "has space", 2, t0+100, "", t0)
			return err
		}, ErrToken},
		{"call", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, worker, "payload", t0+100, "", t0)
			return err
		}, nil},
		{"call empty payload", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, worker, "", t0+100, "", t0)
			return err
		}, nil},
		{"call long payload", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, worker, strings.Repeat("p", MaxPayloadLen+1), t0+100, "", t0)
			return err
		}, ErrPayload},
		{"call host", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, hostPkg, "x", t0+100, "", t0)
			return err
		}, ErrTarget},
		{"call host sub", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, hostPkg+"#s/1", "x", t0+100, "", t0)
			return err
		}, ErrTarget},
		{"call package", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, "gno.land/p/nt/multisig/v0", "x", t0+100, "", t0)
			return err
		}, ErrTarget},
		{"call two hash", func(cur realm) error {
			_, err := s.ProposeCall(0, cur, "gno.land/r/demo/worker#a#b", "x", t0+100, "", t0)
			return err
		}, ErrTarget},
		{"add", func(cur realm) error {
			_, err := s.ProposeAddSigner(0, cur, dave, t0+100, "", t0)
			return err
		}, nil},
		{"add dup", func(cur realm) error {
			_, err := s.ProposeAddSigner(0, cur, alice, t0+100, "", t0)
			return err
		}, ErrDuplicateSigner},
		{"add safe", func(cur realm) error {
			_, err := s.ProposeAddSigner(0, cur, s.Address(), t0+100, "", t0)
			return err
		}, ErrSelf},
		{"remove unknown", func(cur realm) error {
			_, err := s.ProposeRemoveSigner(0, cur, dave, t0+100, "", t0)
			return err
		}, ErrUnknownSigner},
		{"remove", func(cur realm) error {
			_, err := s.ProposeRemoveSigner(0, cur, carol, t0+100, "", t0)
			return err
		}, nil},
		{"threshold", func(cur realm) error {
			_, err := s.ProposeThreshold(0, cur, 1, t0+100, "", t0)
			return err
		}, nil},
		{"threshold zero", func(cur realm) error {
			_, err := s.ProposeThreshold(0, cur, 0, t0+100, "", t0)
			return err
		}, ErrThreshold},
		{"threshold high", func(cur realm) error {
			_, err := s.ProposeThreshold(0, cur, 4, t0+100, "", t0)
			return err
		}, ErrThreshold},
		{"delay", func(cur realm) error {
			_, err := s.ProposeDelay(0, cur, MaxDelay, t0+100, "", t0)
			return err
		}, nil},
		{"delay negative", func(cur realm) error {
			_, err := s.ProposeDelay(0, cur, -1, t0+100, "", t0)
			return err
		}, ErrDelay},
		{"pause", func(cur realm) error {
			_, err := s.ProposePause(0, cur, t0+100, "", t0)
			return err
		}, nil},
		{"unpause early", func(cur realm) error {
			_, err := s.ProposeUnpause(0, cur, t0+100, "", t0)
			return err
		}, ErrNotPaused},
	}
	as(alice, cur, func(cur realm) {
		for _, row := range rows {
			err := row.run(cur)
			if row.want == nil {
				urequire.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, row.want.Error())
			}
		}
	})
	err = s.validateDraft(draft{kind: Kind(99), expiresAt: t0 + 100}, t0)
	uassert.ErrorContains(t, err, ErrKind.Error())
	err = s.applyGovernance(&proposal{kind: Kind(99)})
	uassert.ErrorContains(t, err, ErrKind.Error())
}

func TestStateMachine(cur realm, t *testing.T) {
	s := newSafe(2, 100)
	var id int64
	as(alice, cur, func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 5, t0+10_000, "across", t0)
		urequire.NoError(t, err)
		uassert.ErrorContains(t, s.Revoke(0, cur, id, t0), ErrNotApproved.Error())
		urequire.NoError(t, s.Approve(0, cur, id, t0))
	})
	ready, err := s.Executable(id, t0)
	urequire.NoError(t, err)
	uassert.False(t, ready)
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0+2))
		_, err := s.Execute(0, cur, id, t0+2)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = s.Execute(0, cur, id, t0+101)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
	})
	ready, err = s.Executable(id, t0+102)
	urequire.NoError(t, err)
	uassert.True(t, ready)
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Revoke(0, cur, id, t0+50))
	})
	view, err := s.View(id)
	urequire.NoError(t, err)
	uassert.Equal(t, int64(0), view.ArmedAt)
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0+80))
		_, err := s.Execute(0, cur, id, t0+179)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		eff, err := s.Execute(0, cur, id, t0+180)
		urequire.NoError(t, err)
		uassert.Equal(t, int64(5), eff.Amount)
		_, err = s.Execute(0, cur, id, t0+180)
		uassert.ErrorContains(t, err, ErrNotPending.Error())
		uassert.ErrorContains(t, s.Approve(0, cur, id, t0+180), ErrNotPending.Error())
		uassert.ErrorContains(t, s.Cancel(0, cur, id, t0+180), ErrNotPending.Error())
	})
	uassert.Equal(t, int64(1), s.Nonce())

	var dying int64
	as(alice, cur, func(cur realm) {
		dying, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+300, "", t0+190)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, dying, t0+190))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, dying, t0+190))
		_, err = s.Execute(0, cur, dying, t0+300)
		uassert.ErrorContains(t, err, ErrExpired.Error())
		_, err = s.Execute(0, cur, dying, t0+301)
		uassert.ErrorContains(t, err, ErrExpired.Error())
		uassert.ErrorContains(t, s.Approve(0, cur, dying, t0+301), ErrExpired.Error())
	})
	st, err := s.View(dying)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusExpired.String(), st.Status.String())

	var cancelID int64
	as(alice, cur, func(cur realm) {
		cancelID, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+5000, "", t0+300)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, cancelID, t0+300))
	})
	as(dave, cur, func(cur realm) {
		uassert.ErrorContains(t, s.Cancel(0, cur, cancelID, t0+300), ErrNotSigner.Error())
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, cancelID, t0+300))
		uassert.ErrorContains(t, s.Cancel(0, cur, cancelID, t0+300), ErrArmed.Error())
		urequire.NoError(t, s.Revoke(0, cur, cancelID, t0+301))
		urequire.NoError(t, s.Cancel(0, cur, cancelID, t0+301))
		uassert.ErrorContains(t, s.Cancel(0, cur, cancelID, t0+302), ErrNotPending.Error())
	})
	_, err = s.View(99)
	uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
	_, err = s.Executable(99, t0)
	uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
	ok, err := s.HasApproval(99, alice)
	uassert.False(t, ok)
	uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
}

func TestSignerMembership(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	var spend, remove, add int64
	as(alice, cur, func(cur realm) {
		var err error
		spend, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+1000, "", t0)
		urequire.NoError(t, err)
		remove, err = s.ProposeRemoveSigner(0, cur, alice, t0+1000, "self", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
		urequire.NoError(t, s.Approve(0, cur, remove, t0))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, remove, t0))
		_, err := s.Execute(0, cur, remove, t0)
		urequire.NoError(t, err)
	})
	uassert.False(t, s.IsSigner(alice))
	n, err := s.ApprovalCount(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, 1, n)
	has, err := s.HasApproval(spend, alice)
	urequire.NoError(t, err)
	uassert.False(t, has)
	has, err = s.HasApproval(spend, carol)
	urequire.NoError(t, err)
	uassert.True(t, has)

	as(bob, cur, func(cur realm) {
		var err error
		add, err = s.ProposeAddSigner(0, cur, alice, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, add, t0))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, add, t0))
		_, err := s.Execute(0, cur, add, t0)
		urequire.NoError(t, err)
	})
	uassert.True(t, s.IsSigner(alice))
	n, err = s.ApprovalCount(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, 1, n)

	var dropBob, dropCarol int64
	as(alice, cur, func(cur realm) {
		var err error
		dropBob, err = s.ProposeRemoveSigner(0, cur, bob, t0+1000, "", t0)
		urequire.NoError(t, err)
		dropCarol, err = s.ProposeRemoveSigner(0, cur, carol, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, dropBob, t0))
		urequire.NoError(t, s.Approve(0, cur, dropCarol, t0))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, dropBob, t0))
		urequire.NoError(t, s.Approve(0, cur, dropCarol, t0))
		_, err := s.Execute(0, cur, dropBob, t0)
		urequire.NoError(t, err)
		_, err = s.Execute(0, cur, dropCarol, t0)
		uassert.ErrorContains(t, err, ErrLastSigner.Error())
	})
	uassert.Equal(t, 2, s.SignerCount())
	uassert.Equal(t, 2, s.Threshold())
	view, err := s.View(dropCarol)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusPending.String(), view.Status.String())

	lone := newConfigured(1, 0, alice.String())
	as(alice, cur, func(cur realm) {
		_, err := lone.ProposeRemoveSigner(0, cur, alice, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrLastSigner.Error())
	})
}

func TestThresholdAndDelayInteract(cur realm, t *testing.T) {
	s := newSafe(3, 50)
	var spend, lower, raise, delayID int64
	as(alice, cur, func(cur realm) {
		var err error
		spend, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		lower, err = s.ProposeThreshold(0, cur, 2, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
		urequire.NoError(t, s.Approve(0, cur, lower, t0))
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
		urequire.NoError(t, s.Approve(0, cur, lower, t0))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, lower, t0))
		_, err := s.Execute(0, cur, lower, t0)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = s.Execute(0, cur, lower, t0+50)
		urequire.NoError(t, err)
	})
	uassert.Equal(t, 2, s.Threshold())
	view, err := s.View(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, t0+50, view.ArmedAt)
	ready, err := s.Executable(spend, t0+99)
	urequire.NoError(t, err)
	uassert.False(t, ready)
	ready, err = s.Executable(spend, t0+100)
	urequire.NoError(t, err)
	uassert.True(t, ready)

	as(alice, cur, func(cur realm) {
		var err error
		raise, err = s.ProposeThreshold(0, cur, 3, t0+20_000, "", t0+100)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, raise, t0+100))
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, raise, t0+100))
		_, err := s.Execute(0, cur, raise, t0+149)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = s.Execute(0, cur, raise, t0+150)
		urequire.NoError(t, err)
	})
	uassert.Equal(t, 3, s.Threshold())
	view, err = s.View(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, int64(0), view.ArmedAt)
	uassert.Equal(t, int64(50), view.Delay)
	as(alice, cur, func(cur realm) {
		var err error
		delayID, err = s.ProposeDelay(0, cur, 0, t0+30_000, "", t0+150)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, delayID, t0+150))
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, delayID, t0+150))
	})
	as(carol, cur, func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, delayID, t0+150))
		_, err := s.Execute(0, cur, delayID, t0+200)
		urequire.NoError(t, err)
	})
	uassert.Equal(t, int64(0), s.Delay())
	view, err = s.View(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, int64(50), view.Delay)
}

func TestPauseUnpause(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	worker := "gno.land/r/demo/worker"
	var pauseID, coinID, callID, unpauseID int64
	as(alice, cur, func(cur realm) {
		var err error
		coinID, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+1000, "", t0)
		urequire.NoError(t, err)
		callID, err = s.ProposeCall(0, cur, worker, "go", t0+1000, "", t0)
		urequire.NoError(t, err)
		pauseID, err = s.ProposePause(0, cur, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, coinID, t0))
		urequire.NoError(t, s.Approve(0, cur, callID, t0))
		urequire.NoError(t, s.Approve(0, cur, pauseID, t0))
		_, err = s.Execute(0, cur, pauseID, t0)
		urequire.NoError(t, err)
		_, err = s.ProposePause(0, cur, t0+1000, "", t0)
		uassert.ErrorContains(t, err, ErrAlreadyPaused.Error())
		_, err = s.Execute(0, cur, coinID, t0)
		uassert.ErrorContains(t, err, ErrPaused.Error())
		_, err = s.Execute(0, cur, callID, t0)
		uassert.ErrorContains(t, err, ErrPaused.Error())
		_, err = s.ExecuteChecked(0, cur, coinID, t0, func(Effect) error { return ErrInsufficientBalance })
		uassert.ErrorContains(t, err, ErrPaused.Error())
	})
	uassert.True(t, s.Paused())
	view, err := s.View(coinID)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusPending.String(), view.Status.String())
	as(alice, cur, func(cur realm) {
		var err error
		_, err = s.Execute(0, cur, callID, t0)
		uassert.ErrorContains(t, err, ErrPaused.Error())
		unpauseID, err = s.ProposeUnpause(0, cur, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, unpauseID, t0))
		_, err = s.Execute(0, cur, unpauseID, t0)
		urequire.NoError(t, err)
		_, err = s.ProposeUnpause(0, cur, t0+1000, "", t0)
		uassert.ErrorContains(t, err, ErrNotPaused.Error())
		eff, err := s.Execute(0, cur, callID, t0)
		urequire.NoError(t, err)
		uassert.Equal(t, "go", eff.Payload)
		_, err = s.Consume(0, cur, callID)
		uassert.ErrorContains(t, err, ErrUnauthorized.Error())
	})
	asCode(worker, cur, func(cur realm) {
		payload, err := s.Consume(0, cur, callID)
		urequire.NoError(t, err)
		uassert.Equal(t, "go", payload)
		_, err = s.Consume(0, cur, callID)
		uassert.ErrorContains(t, err, ErrConsumed.Error())
	})
	stale := testing.MakeRealm(dave, "gno.land/r/demo/stale", testing.OriginRealm())
	_, err = s.Consume(0, stale, callID)
	uassert.ErrorContains(t, err, ErrStaleRealm.Error())
	_, err = s.Consume(0, cur, coinID)
	// cur here is the test frame, whose previous is not the worker.
	uassert.ErrorContains(t, err, ErrNotConsumable.Error())
}

func TestConsumeAndAfford(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	worker := "gno.land/r/demo/worker"
	var id int64
	as(alice, cur, func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 3, t0+1000, "", t0)
		urequire.NoError(t, err)
		_, err = s.Consume(0, cur, id)
		uassert.ErrorContains(t, err, ErrNotConsumable.Error())
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		_, err = s.ExecuteChecked(0, cur, id, t0, func(Effect) error { return ErrInsufficientBalance })
		uassert.ErrorContains(t, err, ErrInsufficientBalance.Error())
	})
	view, err := s.View(id)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusPending.String(), view.Status.String())
	uassert.Equal(t, int64(0), s.Nonce())
	as(alice, cur, func(cur realm) {
		eff, err := s.ExecuteChecked(0, cur, id, t0, func(Effect) error { return nil })
		urequire.NoError(t, err)
		uassert.Equal(t, int64(3), eff.Amount)
	})

	var callID int64
	as(alice, cur, func(cur realm) {
		var err error
		callID, err = s.ProposeCall(0, cur, worker, "once", t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, callID, t0))
		_, err = s.Execute(0, cur, callID, t0)
		urequire.NoError(t, err)
	})
	asCode("gno.land/r/demo/other", cur, func(cur realm) {
		_, err := s.Consume(0, cur, callID)
		uassert.ErrorContains(t, err, ErrUnauthorized.Error())
	})
	paused := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		pid, err := paused.ProposeCall(0, cur, worker, "p", t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, paused.Approve(0, cur, pid, t0))
		_, err = paused.Execute(0, cur, pid, t0)
		urequire.NoError(t, err)
		pauseID, err := paused.ProposePause(0, cur, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, paused.Approve(0, cur, pauseID, t0))
		_, err = paused.Execute(0, cur, pauseID, t0)
		urequire.NoError(t, err)
		asHold := pid
		_ = asHold
		_, err = paused.Consume(0, cur, pid)
		uassert.ErrorContains(t, err, ErrPaused.Error())
	})
}

func TestWatchLists(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	as(dave, cur, func(cur realm) {
		uassert.ErrorContains(t, s.WatchDenom(0, cur, "ugnot"), ErrNotSigner.Error())
		uassert.ErrorContains(t, s.WatchToken(0, cur, "gno.land/r/demo/tok.T"), ErrNotSigner.Error())
	})
	as(alice, cur, func(cur realm) {
		uassert.ErrorContains(t, s.WatchDenom(0, cur, "NO"), ErrDenom.Error())
		uassert.ErrorContains(t, s.WatchDenom(0, cur, "ab"), ErrDenom.Error())
		urequire.NoError(t, s.WatchDenom(0, cur, "ugnot"))
		urequire.NoError(t, s.WatchDenom(0, cur, "ugnot"))
		for i := 0; i < MaxWatched-1; i++ {
			urequire.NoError(t, s.WatchDenom(0, cur, "d"+strconv.Itoa(i)+"x"))
		}
		uassert.ErrorContains(t, s.WatchDenom(0, cur, "overflowx"), ErrWatch.Error())
		urequire.NoError(t, s.UnwatchDenom(0, cur, "ugnot"))
		urequire.NoError(t, s.UnwatchDenom(0, cur, "missingx"))
		urequire.NoError(t, s.WatchDenom(0, cur, "ugnot"))
		uassert.ErrorContains(t, s.WatchToken(0, cur, "bad key"), ErrToken.Error())
		for i := 0; i < MaxWatched; i++ {
			urequire.NoError(t, s.WatchToken(0, cur, "gno.land/r/demo/tok.T"+strconv.Itoa(i)))
		}
		uassert.ErrorContains(t, s.WatchToken(0, cur, "gno.land/r/demo/tok.MORE"), ErrWatch.Error())
		urequire.NoError(t, s.UnwatchToken(0, cur, "gno.land/r/demo/tok.T0"))
		urequire.NoError(t, s.UnwatchToken(0, cur, "gno.land/r/demo/tok.ABSENT"))
	})
	uassert.Equal(t, MaxWatched, len(s.WatchedDenoms()))
	uassert.Equal(t, MaxWatched-1, len(s.WatchedTokens()))
}

func TestPendingAndHistoryCaps(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		for i := 0; i < MaxPending; i++ {
			_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
			urequire.NoError(t, err)
		}
		_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrTooManyPending.Error())
		_, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+200, "", t0+100)
		urequire.NoError(t, err)
	})
	uassert.Equal(t, 1, s.PendingCount())

	hist := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		for i := 0; i < MaxRetained; i++ {
			id, err := hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
			urequire.NoError(t, err)
			urequire.NoError(t, hist.Approve(0, cur, id, t0))
			_, err = hist.Execute(0, cur, id, t0)
			urequire.NoError(t, err)
		}
		id, err := hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		_, err = hist.View(1)
		uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
		view, err := hist.View(id)
		urequire.NoError(t, err)
		uassert.Equal(t, StatusPending.String(), view.Status.String())
	})
	uassert.Equal(t, MaxRetained, hist.ProposalCount())
	uassert.Equal(t, int64(MaxRetained), hist.Nonce())

	pinned := newSafe(1, 0)
	worker := "gno.land/r/demo/worker"
	as(alice, cur, func(cur realm) {
		for i := 0; i < MaxRetained; i++ {
			id, err := pinned.ProposeCall(0, cur, worker, "p", t0+10_000, "", t0)
			urequire.NoError(t, err)
			urequire.NoError(t, pinned.Approve(0, cur, id, t0))
			_, err = pinned.Execute(0, cur, id, t0)
			urequire.NoError(t, err)
		}
		_, err := pinned.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		uassert.ErrorContains(t, err, ErrHistory.Error())
	})
	asCode(worker, cur, func(cur realm) {
		payload, err := pinned.Consume(0, cur, 1)
		urequire.NoError(t, err)
		uassert.Equal(t, "p", payload)
	})
	as(alice, cur, func(cur realm) {
		id, err := pinned.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0+1)
		urequire.NoError(t, err)
		_, err = pinned.View(1)
		uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
		view, err := pinned.View(id)
		urequire.NoError(t, err)
		uassert.Equal(t, KindCoin.String(), view.Kind.String())
	})
}

func TestFullSignerSet(cur realm, t *testing.T) {
	csv := signerCSV(MaxSigners)
	s := newConfigured(MaxSigners, 0, csv)
	uassert.Equal(t, MaxSigners, s.SignerCount())
	uassert.Equal(t, signerN(0).String(), s.SignerAt(0).String())
	uassert.Equal(t, "", s.SignerAt(-1).String())
	uassert.Equal(t, "", s.SignerAt(MaxSigners).String())
	var id int64
	as(signerN(0), cur, func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		_, err = s.ProposeAddSigner(0, cur, dave, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrSignerCount.Error())
	})
	for i := 1; i < MaxSigners; i++ {
		i := i
		as(signerN(i), cur, func(cur realm) {
			urequire.NoError(t, s.Approve(0, cur, id, t0))
		})
	}
	as(signerN(0), cur, func(cur realm) {
		_, err := s.Execute(0, cur, id, t0)
		uassert.ErrorContains(t, err, ErrNotReady.Error())
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		_, err = s.Execute(0, cur, id, t0)
		urequire.NoError(t, err)
	})

	one := newConfigured(1, 0, csv)
	as(signerN(3), cur, func(cur realm) {
		pid, err := one.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, one.Approve(0, cur, pid, t0))
		_, err = one.Execute(0, cur, pid, t0)
		urequire.NoError(t, err)
	})
	uassert.Equal(t, int64(1), one.Nonce())
}

func TestDelayZeroAndMax(cur realm, t *testing.T) {
	zero := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		id, err := zero.ProposeCoin(0, cur, dest, "ugnot", 1, t0+1, "", t0)
		urequire.NoError(t, err)
		_, err = zero.ProposeCoin(0, cur, dest, "ugnot", 1, t0, "", t0)
		uassert.ErrorContains(t, err, ErrExpiry.Error())
		urequire.NoError(t, zero.Approve(0, cur, id, t0))
		_, err = zero.Execute(0, cur, id, t0)
		urequire.NoError(t, err)
	})
	slow := newConfigured(1, MaxDelay, alice.String())
	as(alice, cur, func(cur realm) {
		_, err := slow.ProposeCoin(0, cur, dest, "ugnot", 1, t0+MaxDelay, "", t0)
		uassert.ErrorContains(t, err, ErrExpiry.Error())
		id, err := slow.ProposeCoin(0, cur, dest, "ugnot", 1, t0+MaxDelay+1, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, slow.Approve(0, cur, id, t0))
		_, err = slow.Execute(0, cur, id, t0+MaxDelay-1)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = slow.Execute(0, cur, id, t0+MaxDelay)
		urequire.NoError(t, err)
	})
}

func TestClockOverflow(cur realm, t *testing.T) {
	s := newSafe(1, 100)
	as(alice, cur, func(cur realm) {
		now := int64(math.MaxInt64 - 5)
		_, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, now+1, "", now)
		uassert.ErrorContains(t, err, ErrExpiry.Error())
		_, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, 100, "", -1)
		uassert.ErrorContains(t, err, ErrExpiry.Error())
	})

	late := newConfigured(1, MaxDelay, alice.String())
	born := int64(math.MaxInt64 - MaxLifetime)
	as(alice, cur, func(cur realm) {
		id, err := late.ProposeCoin(0, cur, dest, "ugnot", 1, math.MaxInt64, "", born)
		urequire.NoError(t, err)
		urequire.NoError(t, late.Approve(0, cur, id, math.MaxInt64-1))
		_, err = late.Execute(0, cur, id, math.MaxInt64-1)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
	})
}

func TestRecentOrder(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		pid, err := s.ProposePause(0, cur, t0+100, "0", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, pid, t0))
		_, err = s.Execute(0, cur, pid, t0)
		urequire.NoError(t, err)
		_, err = s.ProposePause(0, cur, t0+100, "again", t0)
		uassert.ErrorContains(t, err, ErrAlreadyPaused.Error())
		_, err = s.ProposeCoin(0, cur, dest, "ugnot", 2, t0+100, "", t0)
		urequire.NoError(t, err)
		_, err = s.ProposeCoin(0, cur, dest, "ugnot", 3, t0+100, "", t0)
		urequire.NoError(t, err)
	})
	got := s.Recent(2)
	uassert.Equal(t, 2, len(got))
	uassert.Equal(t, int64(3), got[0].ID)
	uassert.Equal(t, int64(2), got[1].ID)
	uassert.Equal(t, 0, len(s.Recent(0)))
	copies := s.Signers()
	copies[0] = dave
	uassert.True(t, s.IsSigner(alice))
	uassert.Equal(t, alice.String(), s.Creator().String())
	uassert.Equal(t, "treasury", s.Name())
}

func TestPropertySequence(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	seed := uint64(0xC0FFEE)
	executed := 0
	var lastExec int64
	for step := 0; step < 80; step++ {
		seed = seed*6364136223846793005 + 1
		now := t0 + int64(step)
		op := pick(seed, 6)
		who := []address{alice, bob, carol, dave}[pick(seed>>8, 4)]
		ids := livePending(s)
		as(who, cur, func(cur realm) {
			switch op {
			case 0:
				if s.PendingCount() >= MaxPending {
					return
				}
				_, _ = s.ProposeCoin(0, cur, dest, "ugnot", 1, now+20_000, "", now)
			case 1:
				if len(ids) == 0 {
					return
				}
				_ = s.Approve(0, cur, ids[pick(seed, len(ids))], now)
			case 2:
				if len(ids) == 0 {
					return
				}
				_ = s.Revoke(0, cur, ids[pick(seed, len(ids))], now)
			case 3:
				if len(ids) == 0 {
					return
				}
				id := ids[pick(seed, len(ids))]
				if _, err := s.Execute(0, cur, id, now); err == nil {
					executed++
					lastExec = id
				}
			case 4:
				if len(ids) == 0 {
					return
				}
				_ = s.Cancel(0, cur, ids[pick(seed, len(ids))], now)
			default:
				target := []address{bob, carol, dave}[pick(seed>>16, 3)]
				var id int64
				var err error
				if s.IsSigner(target) {
					if s.SignerCount() <= s.Threshold() {
						return
					}
					id, err = s.ProposeRemoveSigner(0, cur, target, now+20_000, "", now)
				} else if s.SignerCount() < MaxSigners {
					id, err = s.ProposeAddSigner(0, cur, target, now+20_000, "", now)
				}
				if err != nil {
					return
				}
				if err = s.Approve(0, cur, id, now); err != nil {
					return
				}
				if _, err = s.Execute(0, cur, id, now); err == nil {
					executed++
					lastExec = id
				}
			}
		})
		if s.Threshold() < 1 || s.Threshold() > s.SignerCount() {
			t.Errorf("step %d threshold %d of %d", step, s.Threshold(), s.SignerCount())
		}
		if s.PendingCount() > MaxPending || s.ProposalCount() > MaxRetained {
			t.Errorf("step %d caps", step)
		}
		if s.Nonce() != int64(executed) {
			t.Errorf("step %d nonce %d executed %d", step, s.Nonce(), executed)
		}
		for _, view := range s.Recent(MaxRetained) {
			n, err := s.ApprovalCount(view.ID)
			urequire.NoError(t, err)
			live := 0
			for _, a := range view.Approvals {
				if s.IsSigner(a) {
					live++
				} else if ok, _ := s.HasApproval(view.ID, a); ok {
					t.Errorf("step %d approval from a non-signer", step)
				}
			}
			if live != n {
				t.Errorf("step %d approval count", step)
			}
		}
		if lastExec != 0 {
			as(alice, cur, func(cur realm) {
				_, err := s.Execute(0, cur, lastExec, now)
				if err == nil {
					t.Errorf("step %d double execute", step)
				}
			})
		}
	}
	if executed == 0 {
		t.Error("property sequence never executed")
	}
}

func TestRemainingBranches(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	uassert.Equal(t, int64(1), s.ID())
	uassert.Equal(t, "gno.land/r/nt/multisig/v0", s.HostPkg())
	_, err := s.ApprovalCount(0)
	uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
	_, err = s.View(-1)
	uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
	as(dave, cur, func(cur realm) {
		uassert.ErrorContains(t, s.UnwatchDenom(0, cur, "ugnot"), ErrNotSigner.Error())
		uassert.ErrorContains(t, s.UnwatchToken(0, cur, "k"), ErrNotSigner.Error())
		_, err := s.ProposeAddSigner(0, cur, address("not-an-address"), t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrNotSigner.Error())
	})
	as(alice, cur, func(cur realm) {
		uassert.ErrorContains(t, s.Approve(0, cur, 99, t0), ErrUnknownProposal.Error())
		uassert.ErrorContains(t, s.Revoke(0, cur, 99, t0), ErrUnknownProposal.Error())
		_, err = s.Execute(0, cur, 99, t0)
		uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
		_, err = s.Consume(0, cur, 99)
		uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
		_, err = s.ProposeGRC20(0, cur, dest, "", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrToken.Error())
		_, err = s.ProposeGRC20(0, cur, address(""), "gno.land/r/demo.T", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
		uassert.ErrorContains(t, s.Cancel(0, cur, 99, t0), ErrUnknownProposal.Error())
		_, err = s.ProposeGRC20(0, cur, dest, "bad key", 1, t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrToken.Error())
		_, err = s.ProposeCall(0, cur, "gno.land/r/", "p", t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrTarget.Error())
		_, err = s.ProposeCall(0, cur, "gno.land/r/a b", "p", t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrTarget.Error())
		_, err = s.ProposeAddSigner(0, cur, address("not-an-address"), t0+100, "", t0)
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
		uassert.True(t, ValidDenom("ugnot"))
		uassert.False(t, ValidDenom("1ab"))
		uassert.False(t, ValidDenom("ab!"))
		key := "gno.land/r/demo/tok.T"
		urequire.NoError(t, s.WatchToken(0, cur, key))
		urequire.NoError(t, s.WatchToken(0, cur, key))
		coin, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, coin, t0))
		uassert.ErrorContains(t, s.Revoke(0, cur, coin, t0+100), ErrExpired.Error())
		uassert.ErrorContains(t, s.Cancel(0, cur, coin, t0+100), ErrExpired.Error())
		ready, err := s.Executable(coin, t0+100)
		urequire.NoError(t, err)
		uassert.False(t, ready)
	})

	live := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		coin, err := live.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, coin, t0))
		_, err = live.Execute(0, cur, coin, t0)
		urequire.NoError(t, err)
		ready, err := live.Executable(coin, t0)
		urequire.NoError(t, err)
		uassert.False(t, ready)
		pause, err := live.ProposePause(0, cur, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, pause, t0))
		_, err = live.Execute(0, cur, pause, t0)
		urequire.NoError(t, err)
		held, err := live.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, held, t0))
		ready, err = live.Executable(held, t0)
		urequire.NoError(t, err)
		uassert.False(t, ready)
		u1, err := live.ProposeUnpause(0, cur, t0+10_000, "", t0)
		urequire.NoError(t, err)
		u2, err := live.ProposeUnpause(0, cur, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, u1, t0))
		urequire.NoError(t, live.Approve(0, cur, u2, t0))
		_, err = live.Execute(0, cur, u1, t0)
		urequire.NoError(t, err)
		_, err = live.Execute(0, cur, u2, t0)
		uassert.ErrorContains(t, err, ErrNotPaused.Error())
		a1, err := live.ProposeAddSigner(0, cur, dave, t0+10_000, "", t0)
		urequire.NoError(t, err)
		a2, err := live.ProposeAddSigner(0, cur, dave, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, a1, t0))
		urequire.NoError(t, live.Approve(0, cur, a2, t0))
		_, err = live.Execute(0, cur, a1, t0)
		urequire.NoError(t, err)
		_, err = live.Execute(0, cur, a2, t0)
		uassert.ErrorContains(t, err, ErrDuplicateSigner.Error())
		r1, err := live.ProposeRemoveSigner(0, cur, dave, t0+10_000, "", t0)
		urequire.NoError(t, err)
		r2, err := live.ProposeRemoveSigner(0, cur, dave, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, r1, t0))
		urequire.NoError(t, live.Approve(0, cur, r2, t0))
		_, err = live.Execute(0, cur, r1, t0)
		urequire.NoError(t, err)
		_, err = live.Execute(0, cur, r2, t0)
		uassert.ErrorContains(t, err, ErrUnknownSigner.Error())
		p1, err := live.ProposePause(0, cur, t0+10_000, "", t0)
		urequire.NoError(t, err)
		p2, err := live.ProposePause(0, cur, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, live.Approve(0, cur, p1, t0))
		urequire.NoError(t, live.Approve(0, cur, p2, t0))
		_, err = live.Execute(0, cur, p1, t0)
		urequire.NoError(t, err)
		_, err = live.Execute(0, cur, p2, t0)
		uassert.ErrorContains(t, err, ErrAlreadyPaused.Error())
	})

	wide := newSafe(2, 0)
	as(alice, cur, func(cur realm) {
		drop, err := wide.ProposeRemoveSigner(0, cur, carol, t0+10_000, "", t0)
		urequire.NoError(t, err)
		high, err := wide.ProposeThreshold(0, cur, 3, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, wide.Approve(0, cur, drop, t0))
		urequire.NoError(t, wide.Approve(0, cur, high, t0))
	})
	as(bob, cur, func(cur realm) {
		urequire.NoError(t, wide.Approve(0, cur, 1, t0))
		urequire.NoError(t, wide.Approve(0, cur, 2, t0))
		_, err := wide.Execute(0, cur, 1, t0)
		urequire.NoError(t, err)
		_, err = wide.Execute(0, cur, 2, t0)
		uassert.ErrorContains(t, err, ErrThreshold.Error())
	})

	epoch := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		id, err := epoch.ProposeCoin(0, cur, dest, "ugnot", 1, 100, "", 0)
		urequire.NoError(t, err)
		urequire.NoError(t, epoch.Approve(0, cur, id, 0))
		_, err = epoch.Execute(0, cur, id, 0)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
	})

	hist := newSafe(1, 0)
	as(alice, cur, func(cur realm) {
		gone, err := hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, hist.Cancel(0, cur, gone, t0))
		hold, err := hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		for hist.ProposalCount() < MaxRetained {
			id, err := hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
			urequire.NoError(t, err)
			urequire.NoError(t, hist.Approve(0, cur, id, t0))
			_, err = hist.Execute(0, cur, id, t0)
			urequire.NoError(t, err)
		}
		_, err = hist.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10_000, "", t0)
		urequire.NoError(t, err)
		_, err = hist.View(gone)
		uassert.ErrorContains(t, err, ErrUnknownProposal.Error())
		_, err = hist.View(hold)
		urequire.NoError(t, err)
	})
}

func livePending(s *Safe) []int64 {
	var ids []int64
	for _, view := range s.Recent(MaxRetained) {
		if view.Status == StatusPending {
			ids = append(ids, view.ID)
		}
	}
	return ids
}

func newConfigured(threshold int, delay int64, csv string) *Safe {
	pkg, addr := host()
	s, err := New(pkg, addr, alice, "treasury", csv, threshold, delay)
	if err != nil {
		panic(err)
	}
	if _, err = s.Bind(1); err != nil {
		panic(err)
	}
	return s
}
