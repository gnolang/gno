package validators

import (
	"testing"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestAllowedRealms_AddRemove(t *testing.T) {
	allowedRealms = bptree.NewBPTree32()

	uassert.False(t, IsAllowedRealm("gno.land/r/demo/ics"))

	allowedRealms.Set("gno.land/r/demo/ics", true)
	uassert.True(t, IsAllowedRealm("gno.land/r/demo/ics"))

	realms := GetAllowedRealms()
	uassert.Equal(t, 1, len(realms))
	uassert.Equal(t, "gno.land/r/demo/ics", realms[0])

	allowedRealms.Remove("gno.land/r/demo/ics")
	uassert.False(t, IsAllowedRealm("gno.land/r/demo/ics"))
	uassert.Equal(t, 0, len(GetAllowedRealms()))
}

func TestAddValidator_NotAllowed(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	resetValset(t)
	resetCache()

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/unknown"))

	op := testutils.TestAddress("op-A")
	urequire.AbortsWithMessage(t,
		"realm gno.land/r/demo/unknown is not allowed to modify the validator set",
		func() {
			AddValidator(cross, ValoperChange{OperatorAddress: op, Power: 10}, pubKeyA, mustAddr(t, pubKeyA))
		},
	)
}

func TestAddValidator_UserNotAllowed(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	resetValset(t)
	resetCache()

	testing.SetRealm(testing.NewUserRealm(mustAddr(t, pubKeyA)))

	op := testutils.TestAddress("op-A")
	urequire.AbortsWithMessage(t,
		"caller must be a realm, not a user",
		func() {
			AddValidator(cross, ValoperChange{OperatorAddress: op, Power: 10}, pubKeyA, mustAddr(t, pubKeyA))
		},
	)
}

func TestAddValidator_Allowed(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	resetValset(t)
	resetCache()
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	op := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyB, keepRunning: true}})

	const allowedPath = "gno.land/r/demo/ics"
	allowedRealms.Set(allowedPath, true)

	testing.SetRealm(testing.NewCodeRealm(allowedPath))

	AddValidator(cross, ValoperChange{OperatorAddress: op, Power: 10}, pubKeyB, mustAddr(t, pubKeyB))
	uassert.True(t, IsValidator(mustAddr(t, pubKeyB)))
}

func TestRemoveValidator_Allowed(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	resetValset(t)
	resetCache()
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	op := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyB, keepRunning: true}})

	const allowedPath = "gno.land/r/demo/ics"
	allowedRealms.Set(allowedPath, true)

	testing.SetRealm(testing.NewCodeRealm(allowedPath))

	addrB := mustAddr(t, pubKeyB)
	AddValidator(cross, ValoperChange{OperatorAddress: op, Power: 10}, pubKeyB, mustAddr(t, pubKeyB))
	uassert.True(t, IsValidator(addrB))

	RemoveValidator(cross, op)
	uassert.False(t, IsValidator(addrB))
}

func TestRemoveValidator_NotAllowed(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	resetValset(t)
	resetCache()

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/unknown"))

	op := testutils.TestAddress("op-A")
	urequire.AbortsWithMessage(t,
		"realm gno.land/r/demo/unknown is not allowed to modify the validator set",
		func() { RemoveValidator(cross, op) },
	)
}

func TestAddValidator_RejectsPowerZero(cur realm, t *testing.T) {
	allowedRealms = bptree.NewBPTree32()
	const allowedPath = "gno.land/r/demo/ics"
	allowedRealms.Set(allowedPath, true)

	testing.SetRealm(testing.NewCodeRealm(allowedPath))

	op := testutils.TestAddress("op-A")
	urequire.AbortsWithMessage(t,
		"AddValidator requires Power > 0; use RemoveValidator for removals",
		func() {
			AddValidator(cross, ValoperChange{OperatorAddress: op, Power: 0}, pubKeyA, mustAddr(t, pubKeyA))
		},
	)
}

func TestNewPropAllowedRealmUpdateRequest_RejectsEmptyTitle(t *testing.T) {
	uassert.PanicsContains(t, "proposal title is empty", func() {
		_ = NewPropAllowedRealmUpdateRequest(
			[]string{"gno.land/r/demo/ics"}, nil, "   ", "",
		)
	})
}

func TestNewPropAllowedRealmUpdateRequest_RejectsEmptyAddAndRemove(t *testing.T) {
	uassert.PanicsContains(t, "proposal must add or remove at least one realm", func() {
		_ = NewPropAllowedRealmUpdateRequest(nil, nil, "title", "")
	})
}

func TestNewPropAllowedRealmUpdateRequest_RejectsDuplicateInAdd(t *testing.T) {
	uassert.PanicsContains(t, "duplicate realm in proposal: gno.land/r/demo/ics", func() {
		_ = NewPropAllowedRealmUpdateRequest(
			[]string{"gno.land/r/demo/ics", "gno.land/r/demo/ics"}, nil, "dup-add", "",
		)
	})
}

func TestNewPropAllowedRealmUpdateRequest_RejectsDuplicateInRemove(t *testing.T) {
	uassert.PanicsContains(t, "duplicate realm in proposal: gno.land/r/demo/ics", func() {
		_ = NewPropAllowedRealmUpdateRequest(
			nil, []string{"gno.land/r/demo/ics", "gno.land/r/demo/ics"}, "dup-remove", "",
		)
	})
}

func TestNewPropAllowedRealmUpdateRequest_RejectsAddRemoveOverlap(t *testing.T) {
	uassert.PanicsContains(t, "duplicate realm in proposal: gno.land/r/demo/ics", func() {
		_ = NewPropAllowedRealmUpdateRequest(
			[]string{"gno.land/r/demo/ics"},
			[]string{"gno.land/r/demo/ics"},
			"overlap", "",
		)
	})
}

func TestNewPropAllowedRealmUpdateRequest_DescriptionRendering(t *testing.T) {
	pr := NewPropAllowedRealmUpdateRequest(
		[]string{"gno.land/r/demo/ics"},
		[]string{"gno.land/r/demo/legacy"},
		"mixed",
		"context line",
	)

	desc := pr.Description()
	urequire.True(t, len(desc) > 0)
	uassert.True(t, contains(desc, "context line"))
	uassert.True(t, contains(desc, "## Allowed Realm Updates"))
	uassert.True(t, contains(desc, "- add: gno.land/r/demo/ics"))
	uassert.True(t, contains(desc, "- remove: gno.land/r/demo/legacy"))
}
