package multisig

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const t0 int64 = 1_700_000_000

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carol = testutils.TestAddress("carol")
	dave  = testutils.TestAddress("dave")
	dest  = testutils.TestAddress("dest")
)

func host() (string, address) {
	const pkg = "gno.land/r/nt/multisig/v0"
	return pkg, chain.PackageAddress(pkg)
}

func newSafe(threshold int, delay int64) *Safe {
	pkg, addr := host()
	signers := alice.String() + "," + bob.String() + "," + carol.String()
	s, err := New(pkg, addr, alice, "treasury", signers, threshold, delay)
	if err != nil {
		panic(err)
	}
	if _, err = s.Bind(1); err != nil {
		panic(err)
	}
	return s
}

func TestThresholdAndExecuteOnce(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	var id int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 10, t0+100, "pay", t0)
		urequire.NoError(t, err)
		_, err = s.Execute(0, cur, id, t0)
		uassert.ErrorContains(t, err, ErrNotReady.Error())
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(dave))
	func(cur realm) {
		uassert.ErrorContains(t, s.Approve(0, cur, id, t0), ErrNotSigner.Error())
		_, err := s.Execute(0, cur, id, t0)
		uassert.ErrorContains(t, err, ErrNotSigner.Error())
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		uassert.ErrorContains(t, s.Approve(0, cur, id, t0), ErrDuplicateApproval.Error())
	}(cross(cur))
	n, err := s.ApprovalCount(id)
	urequire.NoError(t, err)
	uassert.Equal(t, 1, n)
	ready, err := s.Executable(id, t0)
	urequire.NoError(t, err)
	uassert.False(t, ready)

	testing.SetRealm(testing.NewUserRealm(carol))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		eff, err := s.Execute(0, cur, id, t0)
		urequire.NoError(t, err)
		uassert.Equal(t, KindCoin.String(), eff.Kind.String())
		uassert.Equal(t, dest.String(), eff.To.String())
		uassert.Equal(t, "ugnot", eff.Denom)
		uassert.Equal(t, int64(10), eff.Amount)
		_, err = s.Execute(0, cur, id, t0)
		uassert.ErrorContains(t, err, ErrNotPending.Error())
	}(cross(cur))
	uassert.Equal(t, int64(1), s.Nonce())
	view, err := s.View(id)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusExecuted.String(), view.Status.String())
}

func TestTimelockRestartsAfterRevoke(cur realm, t *testing.T) {
	s := newSafe(2, 100)
	var id int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+1000, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, id, t0))
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		_, err := s.Execute(0, cur, id, t0)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = s.Execute(0, cur, id, t0+99)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, s.Revoke(0, cur, id, t0+50))
		urequire.NoError(t, s.Approve(0, cur, id, t0+50))
		_, err := s.Execute(0, cur, id, t0+149)
		uassert.ErrorContains(t, err, ErrTimelock.Error())
		_, err = s.Execute(0, cur, id, t0+150)
		urequire.NoError(t, err)
	}(cross(cur))
}

func TestExpiry(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	var id int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+10, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		_, err = s.Execute(0, cur, id, t0+10)
		uassert.ErrorContains(t, err, ErrExpired.Error())
		_, err = s.Execute(0, cur, id, t0+10)
		uassert.ErrorContains(t, err, ErrExpired.Error())
	}(cross(cur))
	view, err := s.View(id)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusExpired.String(), view.Status.String())
	uassert.Equal(t, StatusExpired.String(), DisplayStatus(view.Status, view.ExpiresAt, t0+10).String())
}

func TestCancelAndArmedLock(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	var id int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(dave))
	func(cur realm) {
		uassert.ErrorContains(t, s.Cancel(0, cur, id, t0), ErrNotSigner.Error())
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, s.Cancel(0, cur, id, t0))
	}(cross(cur))
	view, err := s.View(id)
	urequire.NoError(t, err)
	uassert.Equal(t, StatusCancelled.String(), view.Status.String())

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, id, t0))
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(carol))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		uassert.ErrorContains(t, s.Cancel(0, cur, id, t0), ErrArmed.Error())
		urequire.NoError(t, s.Revoke(0, cur, id, t0))
		urequire.NoError(t, s.Cancel(0, cur, id, t0))
	}(cross(cur))
}

func TestSignerRemovalClearsApprovals(cur realm, t *testing.T) {
	s := newSafe(2, 0)
	var spend, remove int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		spend, err = s.ProposeCoin(0, cur, dest, "ugnot", 5, t0+100, "", t0)
		urequire.NoError(t, err)
		remove, err = s.ProposeRemoveSigner(0, cur, carol, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
		urequire.NoError(t, s.Approve(0, cur, remove, t0))
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(carol))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		urequire.NoError(t, s.Approve(0, cur, remove, t0))
		_, err := s.Execute(0, cur, remove, t0)
		urequire.NoError(t, err)
	}(cross(cur))
	uassert.False(t, s.IsSigner(carol))
	n, err := s.ApprovalCount(spend)
	urequire.NoError(t, err)
	uassert.Equal(t, 1, n)
	ready, err := s.Executable(spend, t0)
	urequire.NoError(t, err)
	uassert.False(t, ready)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		add, err := s.ProposeAddSigner(0, cur, carol, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, add, t0))
	}(cross(cur))
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		// The add proposal is the latest one; find it.
		add := int64(s.ProposalCount())
		urequire.NoError(t, s.Approve(0, cur, add, t0))
		_, err := s.Execute(0, cur, add, t0)
		urequire.NoError(t, err)
	}(cross(cur))
	uassert.True(t, s.IsSigner(carol))
	has, err := s.HasApproval(spend, carol)
	urequire.NoError(t, err)
	uassert.False(t, has, "removal must not leave a vote that returns with the signer")
}

func TestPauseBlocksSpend(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		spend, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		pause, err := s.ProposePause(0, cur, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, spend, t0))
		urequire.NoError(t, s.Approve(0, cur, pause, t0))
		_, err = s.Execute(0, cur, pause, t0)
		urequire.NoError(t, err)
		_, err = s.Execute(0, cur, spend, t0)
		uassert.ErrorContains(t, err, ErrPaused.Error())
		open, err := s.ProposeUnpause(0, cur, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, open, t0))
		_, err = s.Execute(0, cur, open, t0)
		urequire.NoError(t, err)
		_, err = s.Execute(0, cur, spend, t0)
		urequire.NoError(t, err)
	}(cross(cur))
}

func TestCallAuthorizationOnce(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	const target = "gno.land/r/demo/worker"
	var id int64
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		var err error
		id, err = s.ProposeCall(0, cur, target, "do-the-thing", t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		eff, err := s.Execute(0, cur, id, t0)
		urequire.NoError(t, err)
		uassert.Equal(t, target, eff.TargetPkg)
		uassert.Equal(t, "do-the-thing", eff.Payload)
		_, err = s.Consume(0, cur, id)
		uassert.ErrorContains(t, err, ErrUnauthorized.Error())
	}(cross(cur))
	testing.SetRealm(testing.NewCodeRealm(target))
	func(cur realm) {
		payload, err := s.Consume(0, cur, id)
		urequire.NoError(t, err)
		uassert.Equal(t, "do-the-thing", payload)
		_, err = s.Consume(0, cur, id)
		uassert.ErrorContains(t, err, ErrConsumed.Error())
	}(cross(cur))
}

func TestRejectsBadConfig(cur realm, t *testing.T) {
	pkg, addr := host()
	_, err := New(pkg, addr, alice, "ok", alice.String()+","+alice.String(), 1, 0)
	uassert.ErrorContains(t, err, ErrDuplicateSigner.Error())
	_, err = New(pkg, addr, alice, "ok", alice.String(), 2, 0)
	uassert.ErrorContains(t, err, ErrThreshold.Error())
	_, err = New(pkg, addr, alice, "ok", alice.String(), 1, MaxDelay+1)
	uassert.ErrorContains(t, err, ErrDelay.Error())
	_, err = New(pkg, addr, alice, "", alice.String(), 1, 0)
	uassert.ErrorContains(t, err, ErrName.Error())
	_, err = New("gno.land/p/nt/multisig/v0", addr, alice, "ok", alice.String(), 1, 0)
	uassert.ErrorContains(t, err, ErrHost.Error())

	s := newSafe(2, 0)
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		_, err := s.ProposeCoin(0, cur, dest, "ugnot", 0, t0+10, "", t0)
		uassert.ErrorContains(t, err, ErrAmount.Error())
		_, err = s.ProposeCoin(0, cur, s.Address(), "ugnot", 1, t0+10, "", t0)
		uassert.ErrorContains(t, err, ErrSelf.Error())
		_, err = s.ProposeCoin(0, cur, dest, "ugnot", 1, t0, "", t0)
		uassert.ErrorContains(t, err, ErrExpiry.Error())
		_, err = s.ProposeRemoveSigner(0, cur, dave, t0+10, "", t0)
		uassert.ErrorContains(t, err, ErrUnknownSigner.Error())
	}(cross(cur))
}

func TestDelayChangeDoesNotRewritePending(cur realm, t *testing.T) {
	s := newSafe(1, 0)
	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		id, err := s.ProposeCoin(0, cur, dest, "ugnot", 1, t0+100, "", t0)
		urequire.NoError(t, err)
		change, err := s.ProposeDelay(0, cur, 500, t0+100, "", t0)
		urequire.NoError(t, err)
		urequire.NoError(t, s.Approve(0, cur, change, t0))
		_, err = s.Execute(0, cur, change, t0)
		urequire.NoError(t, err)
		uassert.Equal(t, int64(500), s.Delay())
		view, err := s.View(id)
		urequire.NoError(t, err)
		uassert.Equal(t, int64(0), view.Delay)
		urequire.NoError(t, s.Approve(0, cur, id, t0))
		_, err = s.Execute(0, cur, id, t0)
		urequire.NoError(t, err)
	}(cross(cur))
}
