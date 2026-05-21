package valopers

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	// Second valid pubkey used as the rotation target. Distinct from
	// validValidatorInfo's pubkey so the signingRegistry uniqueness
	// check is exercised.
	rotateTargetPubKey = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq3ds6sdvc0shfkq02h6xx5g0jp04aadexfnpsmgjxu72xz9y30aqfrlpny"

	// Test-local mirror of the rotation_period_blocks sys-param's
	// default. resetState() seeds the same value into sysparams.
	testRotationPeriodBlocks = int64(600)
)

// registerForRotation does the boilerplate setup that every rotation
// test starts from: clear state, register the valoper as info.Address,
// and return the info struct.
func registerForRotation(t *testing.T) struct {
	Moniker     string
	Description string
	ServerType  string
	Address     address
	PubKey      string
} {
	t.Helper()
	resetState()
	info := validValidatorInfo(t)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	testing.SetOriginSend(chain.Coins{minFee})
	Register(cross, info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
	return info
}

func TestUpdateSigningKey_HappyPath(t *testing.T) {
	info := registerForRotation(t)

	// Advance past the throttle window. SkipHeights doesn't preserve
	// the realm context across the boundary, so re-set it before the
	// next cross-call.
	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))

	uassert.NotAborts(t, func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})

	v := GetByAddr(info.Address)
	uassert.Equal(t, rotateTargetPubKey, v.SigningPubKey)

	newSigningAddr, err := chain.PubKeyAddress(rotateTargetPubKey)
	urequire.NoError(t, err)
	uassert.Equal(t, newSigningAddr, v.SigningAddress)

	// New entry active in registry.
	rawNew, ok := signingRegistry.Get(newSigningAddr.String())
	urequire.True(t, ok, "new signing addr in registry")
	newEntry := rawNew.(regEntry)
	uassert.Equal(t, info.Address, newEntry.OperatorAddress)
	uassert.Equal(t, int64(0), newEntry.RetiredAtHeight)

	// Old entry retired (still present, RetiredAtHeight > 0).
	oldSigningAddr, err := chain.PubKeyAddress(info.PubKey)
	urequire.NoError(t, err)
	rawOld, ok := signingRegistry.Get(oldSigningAddr.String())
	urequire.True(t, ok, "old signing addr retained as retired")
	oldEntry := rawOld.(regEntry)
	uassert.Equal(t, info.Address, oldEntry.OperatorAddress)
	uassert.True(t, oldEntry.RetiredAtHeight > 0, "old entry must be retired (RetiredAtHeight > 0)")
}

func TestUpdateSigningKey_ThrottleRejection(t *testing.T) {
	info := registerForRotation(t)
	testing.SetRealm(testing.NewUserRealm(info.Address))

	// Try to rotate immediately, before throttle elapses.
	uassert.AbortsWithMessage(t, ErrRotationThrottled.Error(), func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})

	// Advance just shy of the threshold; still rejected.
	testing.SkipHeights(testRotationPeriodBlocks - 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	uassert.AbortsWithMessage(t, ErrRotationThrottled.Error(), func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})

	// One more block — now exactly at the threshold; allowed.
	testing.SkipHeights(1)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	uassert.NotAborts(t, func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})
}

func TestUpdateSigningKey_RejectsNonAuthListCaller(t *testing.T) {
	info := registerForRotation(t)
	testing.SkipHeights(testRotationPeriodBlocks + 1)

	// Switch to a caller not on the auth list.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("attacker")))

	// Pin the exact authorizable error so this catches an auth
	// regression rather than any abort. Empty-substring match (the
	// previous form) accepted any panic, including unrelated ones.
	uassert.AbortsContains(t, "not in authorized list", func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})
}

func TestUpdateSigningKey_RejectsReuseOfActiveKey(t *testing.T) {
	info := registerForRotation(t)
	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))

	// Try to "rotate" to the same key — already in registry.
	uassert.AbortsWithMessage(t, ErrSigningKeyTaken.Error(), func() {
		UpdateSigningKey(cross, info.Address, info.PubKey)
	})
}

func TestUpdateSigningKey_RejectsRotationOntoActiveValidator(t *testing.T) {
	// Front-running guard: a rotation whose derived signing address is
	// already an active validator must be rejected. Mirrors the same
	// guard in Register (valopers.gno: ErrFrontrunValidator). Without
	// this, any valoper could rotate onto an unregistered (e.g.
	// genesis-seeded) validator slot — v3's executor would overwrite
	// the active validator's entry in the effective valset with this
	// operator's claim, then a subsequent govDAO remove-op proposal
	// would delete it. signingRegistry uniqueness alone doesn't catch
	// this case: a genesis-seeded validator never went through
	// Register or UpdateSigningKey, so its signing address is absent
	// from signingRegistry.
	info := registerForRotation(t)

	// Seed v3's valset:current with the address that rotateTargetPubKey
	// derives to. ValsetEffective is read from valset:current when the
	// dirty flag is unset, which is the default.
	testing.SetSysParamStrings("node", "valset", "current",
		[]string{rotateTargetPubKey + ":1"})

	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))

	uassert.AbortsWithMessage(t, ErrFrontrunValidator.Error(), func() {
		UpdateSigningKey(cross, info.Address, rotateTargetPubKey)
	})

	// Cleanup: clear the seeded valset to avoid leaking into sibling
	// subtests if the runner doesn't reset package state between them.
	testing.SetSysParamStrings("node", "valset", "current", []string{})
}

func TestUpdateSigningKey_RejectsReuseOfRetiredKey(t *testing.T) {
	info := registerForRotation(t)

	// First rotation: info.PubKey -> rotateTargetPubKey. info.PubKey
	// becomes retired.
	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	UpdateSigningKey(cross, info.Address, rotateTargetPubKey)

	// Second rotation back to info.PubKey must fail — it's retired
	// but signingRegistry retains it forever.
	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	uassert.AbortsWithMessage(t, ErrSigningKeyTaken.Error(), func() {
		UpdateSigningKey(cross, info.Address, info.PubKey)
	})
}

func TestUpdateSigningKey_LastRotationHeightAdvances(t *testing.T) {
	info := registerForRotation(t)
	testing.SkipHeights(testRotationPeriodBlocks + 1)
	testing.SetRealm(testing.NewUserRealm(info.Address))
	UpdateSigningKey(cross, info.Address, rotateTargetPubKey)

	v := GetByAddr(info.Address)
	uassert.True(t, v.LastRotationHeight > 0, "LastRotationHeight set on rotation")

	// Immediately after, throttle rejects again.
	uassert.AbortsWithMessage(t, ErrRotationThrottled.Error(), func() {
		UpdateSigningKey(cross, info.Address, "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqddddqg2glc8x4fl7vxjlnr7p5a3czm5kcdp4239sg6yqdc4rc2r5cjrffs")
	})
}
