package valopers

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/ownable/v0/exts/authorizable"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
)

// Test-local fee constant. Production reads register_fee from sysparams;
// these tests use this value as a reference Coin for OriginSend setup
// and (when needed) seed it via testing.SetSysParamUint64.
var minFee = chain.NewCoin("ugnot", 20*1_000_000)

// cur is a zero-value realm used as a placeholder when forwarding to
// uassert dispatch helpers that gained an `rlm realm` param. These tests
// pass `func()` callbacks (no crossing inside the callback), so rlm is
// ignored — a nil realm here is safe.
var cur realm

// resetState clears realm-level state and zeroes the valoper sys-params
// so subtests don't leak through valopers (operator slots),
// signingRegistry (signing-address uniqueness), or sys-param values.
func resetState() {
	valopers = avl.NewTree()
	signingRegistry = bptree.NewBPTree32()
	testing.SetSysParamUint64("node", "valoper", "register_fee", 0)
	testing.SetSysParamUint64("node", "valoper", "rotation_fee", 0)
	testing.SetSysParamInt64("node", "valoper", "rotation_period_blocks", 600)
}

// enableRegisterFee seeds register_fee = minFee.Amount in sysparams
// so the Register fee path fires. Subtests that exercise fee
// rejection or sufficient-fee acceptance call this after resetState.
func enableRegisterFee() {
	testing.SetSysParamUint64("node", "valoper", "register_fee", uint64(minFee.Amount))
}

func validValidatorInfo(t *testing.T) struct {
	Moniker     string
	Description string
	ServerType  string
	Address     address
	PubKey      string
} {
	t.Helper()

	return struct {
		Moniker     string
		Description string
		ServerType  string
		Address     address
		PubKey      string
	}{
		Moniker:     "test-1",
		Description: "test-1's description",
		ServerType:  ServerTypeOnPrem,
		Address:     address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h"),
		PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p",
	}
}

func TestValopers_Register(cur realm, t *testing.T) {
	t.Run("already a valoper", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))

		v := Valoper{
			Moniker:         info.Moniker,
			Description:     info.Description,
			ServerType:      info.ServerType,
			OperatorAddress: info.Address,
			SigningPubKey:   info.PubKey,
			KeepRunning:     true,
		}

		// Add the valoper directly to the slot.
		valopers.Set(v.OperatorAddress.String(), v)

		// Send coins.
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.AbortsWithMessage(t, cur, ErrValoperExists.Error(), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("no coins deposited", func(cur realm, t *testing.T) {
		resetState()
		enableRegisterFee()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))

		// Send no coins.
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", 0)})

		uassert.AbortsWithMessage(t, cur, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("insufficient coins amount deposited", func(cur realm, t *testing.T) {
		resetState()
		enableRegisterFee()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))

		// Send invalid coins.
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", minFee.Amount-1)})

		uassert.AbortsWithMessage(t, cur, ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("coin amount deposited is not ugnot", func(cur realm, t *testing.T) {
		resetState()
		enableRegisterFee()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))

		// Send invalid coins.
		testing.SetOriginSend(chain.Coins{chain.NewCoin("gnogno", minFee.Amount)})

		uassert.AbortsWithMessage(t, cur, "incompatible coin denominations: gnogno, ugnot", func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("squat guard rejects mismatched OriginCaller", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		// Caller is NOT info.Address: post-genesis squat guard fires.
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("attacker")))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.AbortsWithMessage(t, cur, ErrOperatorSquatGuard.Error(), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})
	})

	t.Run("successful registration", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))

		// Send coins.
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotAborts(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)

			uassert.Equal(t, info.Moniker, valoper.Moniker)
			uassert.Equal(t, info.Description, valoper.Description)
			uassert.Equal(t, info.ServerType, valoper.ServerType)
			uassert.Equal(t, info.Address, valoper.OperatorAddress)
			uassert.Equal(t, info.PubKey, valoper.SigningPubKey)
			uassert.Equal(t, true, valoper.KeepRunning)

			// SigningAddress is derived from the pubkey and present.
			derived, err := chain.PubKeyAddress(info.PubKey)
			uassert.NoError(t, err)
			uassert.Equal(t, derived, valoper.SigningAddress)

			// signingRegistry tracks the active entry.
			_, exists := signingRegistry.Get(derived.String())
			uassert.True(t, exists, "signingRegistry must contain the active entry")
		})
	})

	t.Run("signing-key reuse rejected", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		// First registration succeeds.
		uassert.NotAborts(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Second attempt with a different operator addr but the same
		// pubkey must fail signingRegistry uniqueness.
		other := testutils.TestAddress("other-op")
		testing.SetRealm(testing.NewUserRealm(other))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.AbortsWithMessage(t, cur, ErrSigningKeyTaken.Error(), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, other, info.PubKey)
		})
	})

	t.Run("front-running guard rejects post-genesis if signing addr already validates", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		// Seed v3's valset:current with the very signing address that
		// `info.PubKey` derives to (g1sp8v98...). Any post-genesis
		// Register attempting the same pubkey now trips
		// `ChainHeight()>0 && validators.IsValidator(signingAddr)`.
		testing.SetSysParamStrings("node", "valset", "current", []string{info.PubKey + ":1"})

		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.AbortsWithMessage(t, cur, ErrFrontrunValidator.Error(), func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Cleanup: clear the seeded valset to avoid leaking into
		// later subtests if the package state isn't reset between
		// them in this test runner mode.
		testing.SetSysParamStrings("node", "valset", "current", []string{})
	})
}

func TestValopers_Register_AuthOwnerIsOperatorAddress(cur realm, t *testing.T) {
	// Pin: the Authorizable owner is bound to the OperatorAddress
	// (the addr arg), NOT to OriginCaller. This matters in the
	// genesis-mode deployer pattern: one signer (e.g., the hardfork
	// ceremony deployer) registers profiles for many operators.
	// Each operator must end up on their own profile's auth list
	// so they can manage it post-genesis without depending on the
	// deployer.
	t.Run("genesis deployer pattern: operator (not deployer) is owner", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		deployer := testutils.TestAddress("deployer")

		// Genesis mode: ChainHeight()==0 bypasses the squat guard so
		// deployer (OriginCaller) can register a profile for a
		// different operator addr.
		testing.SetHeight(0)
		testing.SetRealm(testing.NewUserRealm(deployer))
		testing.SetOriginCaller(deployer)
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Auth owner must be the operator addr (info.Address), NOT
		// the deployer.
		v := GetByAddr(info.Address)
		uassert.Equal(t, info.Address.String(), v.Auth().Owner().String(),
			"auth owner must equal OperatorAddress, not OriginCaller")

		// Operator can manage their own profile post-genesis.
		testing.SetHeight(100)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginCaller(info.Address)
		uassert.NotPanics(t, cur, func() {
			UpdateMoniker(cross(cur), info.Address, "operator-renamed")
		})
		uassert.Equal(t, "operator-renamed", GetByAddr(info.Address).Moniker)

		// Deployer cannot manage the operator's profile (not on the
		// auth list).
		testing.SetRealm(testing.NewUserRealm(deployer))
		testing.SetOriginCaller(deployer)
		uassert.AbortsContains(t, cur, "caller is not in authorized list", func() {
			UpdateMoniker(cross(cur), info.Address, "deployer-attempt")
		})
	})

	t.Run("post-genesis self-Register: operator is owner", func(cur realm, t *testing.T) {
		// At H>0 the squat guard forces OriginCaller==addr, so owner
		// would be the same regardless. This subtest pins that the
		// post-genesis behavior is unchanged.
		resetState()

		info := validValidatorInfo(t)
		testing.SetHeight(100)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginCaller(info.Address)
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		v := GetByAddr(info.Address)
		uassert.Equal(t, info.Address.String(), v.Auth().Owner().String())
	})
}

func TestValopers_UpdateAuthMembers(cur realm, t *testing.T) {
	test2Address := testutils.TestAddress("test2")

	t.Run("unauthorized member adds member", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		// Add the valoper.
		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// A different caller (not on the auth list) tries to add a member.
		testing.SetRealm(testing.NewUserRealm(test2Address))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotSuperuser.Error(), func() {
			AddToAuthList(cross(cur), info.Address, test2Address)
		})
	})

	t.Run("unauthorized member deletes member", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, cur, func() {
			AddToAuthList(cross(cur), info.Address, test2Address)
		})

		// A different caller tries to delete a member.
		testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("attacker")))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotSuperuser.Error(), func() {
			DeleteFromAuthList(cross(cur), info.Address, test2Address)
		})
	})

	t.Run("authorized member adds member", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, cur, func() {
			AddToAuthList(cross(cur), info.Address, test2Address)
		})

		testing.SetRealm(testing.NewUserRealm(test2Address))

		newMoniker := "new moniker"
		uassert.NotPanics(t, cur, func() {
			UpdateMoniker(cross(cur), info.Address, newMoniker)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})

	t.Run("invalid member address rejected", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Caller is the owner, but the member address is invalid: the
		// bech32 guard rejects it before reaching the auth list. Empty
		// and malformed-but-nonempty (bad checksum) are both rejected,
		// so a future weakening to a mere empty-check would fail here.
		for _, bad := range []address{address(""), address("g1nonsense")} {
			uassert.AbortsWithMessage(t, cur, ErrInvalidAddress.Error(), func() {
				AddToAuthList(cross(cur), info.Address, bad)
			})

			uassert.AbortsWithMessage(t, cur, ErrInvalidAddress.Error(), func() {
				DeleteFromAuthList(cross(cur), info.Address, bad)
			})
		}
	})
}

func TestValopers_RenderSanitizesDescription(cur realm, t *testing.T) {
	resetState()

	info := validValidatorInfo(t)
	// Description that tries to inject a top-level heading. Input
	// validation accepts it (only length is checked); Render must
	// neutralize the structural markdown before it reaches the profile
	// page and the govdao proposal that embeds Render.
	info.Description = "# FORGEDHEADING\nbody FORGEDBODY"

	testing.SetRealm(testing.NewUserRealm(info.Address))
	testing.SetOriginSend(chain.Coins{minFee})
	uassert.NotPanics(t, cur, func() {
		Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
	})

	out := GetByAddr(info.Address).Render()

	// Prose survives.
	uassert.True(t, strings.Contains(out, "FORGEDBODY"), "description body must be preserved")
	uassert.True(t, strings.Contains(out, "FORGEDHEADING"), "heading text must be preserved as prose")
	// But not as a live ATX heading line.
	uassert.False(t, strings.Contains(out, "\n# FORGEDHEADING"), "forged heading must be neutralized")

	// A non-empty description that sanitizes to empty (lone link-
	// reference definition) must not glue the heading onto the chrome.
	testing.SetRealm(testing.NewUserRealm(info.Address))
	uassert.NotPanics(t, cur, func() {
		UpdateDescription(cross(cur), info.Address, "[a]: https://x.example")
	})
	out2 := GetByAddr(info.Address).Render()
	uassert.True(t, strings.Contains(out2, "## "+info.Moniker+"\n"), "heading line must stay terminated")
	uassert.True(t, strings.Contains(out2, "\n- Operator Address:"), "operator-address line must not merge into the heading")
}

func TestValopers_UpdateMoniker(cur realm, t *testing.T) {
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)

		uassert.AbortsWithMessage(t, cur, ErrValoperMissing.Error(), func() {
			UpdateMoniker(cross(cur), info.Address, "new moniker")
		})
	})

	t.Run("invalid caller", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		// Change the caller to someone not on the auth list.
		testing.SetRealm(testing.NewUserRealm(test2Address))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateMoniker(cross(cur), info.Address, "new moniker")
		})
	})

	t.Run("invalid moniker", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		invalidMonikers := []string{
			"",     // Empty
			"    ", // Whitespace
			"a",    // Too short
			"a very long moniker that is longer than 32 characters", // Too long
			"!@#$%^&*()+{}|:<>?/.,;'",                               // Invalid characters
			" space in front",
			"space in back ",
		}

		for _, invalidMoniker := range invalidMonikers {
			uassert.AbortsWithMessage(t, cur, ErrInvalidMoniker.Error(), func() {
				UpdateMoniker(cross(cur), info.Address, invalidMoniker)
			})
		}
	})

	t.Run("too long moniker", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.AbortsWithMessage(t, cur, ErrInvalidMoniker.Error(), func() {
			UpdateMoniker(cross(cur), info.Address, strings.Repeat("a", MonikerMaxLength+1))
		})
	})

	t.Run("successful update", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		newMoniker := "new moniker"
		uassert.NotPanics(t, cur, func() {
			UpdateMoniker(cross(cur), info.Address, newMoniker)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, newMoniker, valoper.Moniker)
		})
	})
}

func TestValopers_UpdateDescription(cur realm, t *testing.T) {
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(cur realm, t *testing.T) {
		resetState()

		uassert.AbortsWithMessage(t, cur, ErrValoperMissing.Error(), func() {
			UpdateDescription(cross(cur), validValidatorInfo(t).Address, "new description")
		})
	})

	t.Run("invalid caller", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		testing.SetRealm(testing.NewUserRealm(test2Address))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateDescription(cross(cur), info.Address, "new description")
		})
	})

	t.Run("empty description", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.AbortsWithMessage(t, cur, ErrInvalidDescription.Error(), func() {
			UpdateDescription(cross(cur), info.Address, "")
		})
	})

	t.Run("too long description", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.AbortsWithMessage(t, cur, ErrInvalidDescription.Error(), func() {
			UpdateDescription(cross(cur), info.Address, strings.Repeat("a", DescriptionMaxLength+1))
		})
	})

	t.Run("successful update", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		newDescription := "new description"
		uassert.NotPanics(t, cur, func() {
			UpdateDescription(cross(cur), info.Address, newDescription)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, newDescription, valoper.Description)
		})
	})
}

func TestValopers_UpdateKeepRunning(cur realm, t *testing.T) {
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(cur realm, t *testing.T) {
		resetState()

		uassert.AbortsWithMessage(t, cur, ErrValoperMissing.Error(), func() {
			UpdateKeepRunning(cross(cur), validValidatorInfo(t).Address, false)
		})
	})

	t.Run("invalid caller", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		testing.SetRealm(testing.NewUserRealm(test2Address))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateKeepRunning(cross(cur), info.Address, false)
		})
	})

	t.Run("successful update", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, cur, func() {
			UpdateKeepRunning(cross(cur), info.Address, false)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, false, valoper.KeepRunning)
		})
	})
}

func TestValopers_UpdateServerType(cur realm, t *testing.T) {
	test2Address := testutils.TestAddress("test2")

	t.Run("non-existing valoper", func(cur realm, t *testing.T) {
		resetState()

		uassert.AbortsWithMessage(t, cur, ErrValoperMissing.Error(), func() {
			UpdateServerType(cross(cur), validValidatorInfo(t).Address, ServerTypeCloud)
		})
	})

	t.Run("invalid caller", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		testing.SetRealm(testing.NewUserRealm(test2Address))

		uassert.AbortsWithMessage(t, cur, authorizable.ErrNotInAuthList.Error(), func() {
			UpdateServerType(cross(cur), info.Address, ServerTypeCloud)
		})
	})

	t.Run("invalid server type", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		invalidServerTypes := []string{
			"",
			"invalid",
			"Cloud",      // case sensitive
			"ON-PREM",    // case sensitive
			"datacenter", // wrong format
		}

		for _, invalidType := range invalidServerTypes {
			uassert.AbortsWithMessage(t, cur, ErrInvalidServerType.Error(), func() {
				UpdateServerType(cross(cur), info.Address, invalidType)
			})
		}
	})

	t.Run("successful update", func(cur realm, t *testing.T) {
		resetState()

		info := validValidatorInfo(t)
		testing.SetRealm(testing.NewUserRealm(info.Address))
		testing.SetOriginSend(chain.Coins{minFee})

		uassert.NotPanics(t, cur, func() {
			Register(cross(cur), info.Moniker, info.Description, info.ServerType, info.Address, info.PubKey)
		})

		uassert.NotPanics(t, cur, func() {
			UpdateServerType(cross(cur), info.Address, ServerTypeCloud)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, ServerTypeCloud, valoper.ServerType)
		})

		uassert.NotPanics(t, cur, func() {
			UpdateServerType(cross(cur), info.Address, ServerTypeDataCenter)
		})

		uassert.NotPanics(t, cur, func() {
			valoper := GetByAddr(info.Address)
			uassert.Equal(t, ServerTypeDataCenter, valoper.ServerType)
		})
	})
}
