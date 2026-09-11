// Package valopers is designed around the permissionless lifecycle of valoper profiles.
package valopers

import (
	"chain"
	"chain/runtime"
	"chain/runtime/unsafe"
	"crypto/bech32"
	"errors"
	"math"
	"regexp"

	"gno.land/p/moul/realmpath/v0"
	"gno.land/p/nt/avl/pager/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/combinederr/v0"
	"gno.land/p/nt/ownable/exts/authorizable/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ufmt/v0"
	sysparams "gno.land/r/sys/params"
	validators "gno.land/r/sys/validators/v0"
)

const (
	MonikerMaxLength     = 32
	DescriptionMaxLength = 2048

	// Valid server types
	ServerTypeCloud      = "cloud"
	ServerTypeOnPrem     = "on-prem"
	ServerTypeDataCenter = "data-center"
)

var (
	ErrValoperExists        = errors.New("valoper already exists")
	ErrValoperMissing       = errors.New("valoper does not exist")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrInvalidMoniker       = errors.New("moniker is not valid")
	ErrInvalidDescription   = errors.New("description is not valid")
	ErrInvalidServerType    = errors.New("server type is not valid")
	ErrOperatorSquatGuard   = errors.New("post-genesis: caller must equal operator address")
	ErrSigningKeyTaken      = errors.New("signing address already in registry (active or retired)")
	ErrFrontrunValidator    = errors.New("post-genesis: signing address is already an active validator")
	ErrRotationThrottled    = errors.New("rotation throttled: try again later")
	ErrRegistryEntryMissing = errors.New("signing address has no active registry entry (corrupted state)")
	ErrPaidCallNotDirect    = errors.New("a fee is configured: call this directly as a user (maketx call), not from a realm or a maketx run script")
	ErrFeeParamOutOfRange   = errors.New("configured fee exceeds the maximum representable coin amount")
	ErrDisallowedPubKeyType = errors.New("consensus pubkey type is not allowed for validators")
)

var (
	valopers     *avl.Tree // operator-address -> Valoper
	instructions string    // markdown instructions for valoper's registration

	// signingRegistry maps SigningAddress.String() -> regEntry.
	// Permanently retains retired entries to prevent key reuse and
	// to support future slashing-attribution by signing address.
	signingRegistry = bptree.NewBPTree32()

	monikerMaxLengthMiddle = ufmt.Sprintf("%d", MonikerMaxLength-2)
	validateMonikerRe      = regexp.MustCompile(`^[a-zA-Z0-9][\w -]{0,` + monikerMaxLengthMiddle + `}[a-zA-Z0-9]$`) // 32 characters, including spaces, hyphens or underscores in the middle
)

// regEntry tracks signing-address -> operator with retirement metadata.
// retiredAtHeight == 0 means the entry is currently active for the operator.
type regEntry struct {
	OperatorAddress    address
	RegisteredAtHeight int64
	RetiredAtHeight    int64
}

// Valoper represents a validator operator profile.
type Valoper struct {
	Moniker     string // A human-readable name
	Description string // A description and details about the valoper
	ServerType  string // The type of server (cloud/on-prem/data-center)

	OperatorAddress address // operator identity, profile key, stable across rotations
	SigningPubKey   string  // current consensus signing pubkey (bech32 gpub1...)
	SigningAddress  address // = chain.PubKeyAddress(SigningPubKey)

	LastRotationHeight int64 // throttle anchor for UpdateSigningKey

	KeepRunning bool // operator wants this validator running in the active set

	auth *authorizable.Authorizable
}

// AuthOwner returns the operator address that owns this profile's auth
// list. Read-only by construction: it copies out an address rather than
// returning the live *authorizable.Authorizable.
//
// SECURITY: the previous `Auth() *authorizable.Authorizable` was a
// capability leak of the same class this realm's governance authority
// closes by returning a description instead of the live authority.
// `Valoper` is returned BY VALUE from the exported, non-crossing
// GetByAddr, but `auth` is a pointer field, so the copy shared the
// callee's Authorizable. Any realm could therefore obtain a live,
// mutable handle and write through it.
//
// That was a privilege escalation, not merely a wider surface.
// Authorizable's gates read `rlm.Previous().Address()`, so inside a
// hostile realm's frame `Previous()` is whoever called it. When a
// valoper OPERATOR called any function of a hostile realm (faucet,
// airdrop, mint), that realm could reach
// `GetByAddr(operator).Auth().AddToAuthList(...)` and — because
// Previous() was then the operator, the Authorizable's own owner —
// persist itself onto the operator's auth list. From the next
// transaction on it acted alone: UpdateKeepRunning to drain the
// validator, UpdateSigningKey to rotate the consensus key.
//
// The exported wrappers below are NOT equivalent to the raw handle and
// were never the hole: there `cur.Previous()` is the hostile realm
// itself rather than the operator, so the owner check rejects it.
func (v Valoper) AuthOwner() address {
	return v.auth.Owner()
}

func AddToAuthList(cur realm, addr address, member address) {
	v := GetByAddr(addr)
	if err := v.auth.AddToAuthList(0, cur, member); err != nil {
		panic(err)
	}
}

func DeleteFromAuthList(cur realm, addr address, member address) {
	v := GetByAddr(addr)
	if err := v.auth.DeleteFromAuthList(0, cur, member); err != nil {
		panic(err)
	}
}

// Register registers a new valoper. The `addr` parameter is the
// operator address (stable identity, profile key); `pubKey` is the
// consensus signing pubkey, from which the signing address is derived.
//
// Auth shape:
//   - Post-genesis: OriginCaller must equal addr (operator-slot squat
//     guard). Genesis-mode replay (ChainHeight()==0) bypasses, so
//     migration .jsonl txs and historical Register replays succeed.
//   - Signing-address uniqueness: derived(pubKey) must not already be
//     in signingRegistry, active or retired.
//   - Front-running guard: post-genesis, derived(pubKey) must not
//     already be an active validator (a fresh registration cannot
//     squat on the consensus address of an existing validator).
//
// Why OriginCaller==addr is sufficient for the SQUAT guard (no
// IsUserCall): squatting requires the attacker to satisfy
// OriginCaller==victim, which requires the victim's signing key.
//
// The PAYMENT check is a different matter and does need IsUserCall —
// see assertPaidCallIsDirect. An earlier version of this comment
// claimed the fee was "validated against banker.OriginSend in a way
// that's symmetric to IsUserCall via direct comparison"; there was no
// such comparison, and the fee was bypassable (reported by @D4ryl00).
//
// Auth-list seeding: the profile's Authorizable owner is set to addr
// (NOT OriginCaller). At H>0 the squat guard makes them equal anyway;
// at H==0 the deployer pattern (one signer registers many operators)
// requires owner == addr so each operator can manage their own profile
// post-genesis without needing the deployer's auth.
func Register(cur realm, moniker string, description string, serverType string, addr address, pubKey string) {
	// Operator-slot squat guard.
	if runtime.ChainHeight() > 0 && unsafe.OriginCaller() != addr {
		panic(ErrOperatorSquatGuard)
	}

	// Fee enforcement (read from sysparams; defaults to 0 until
	// governance raises it post-transfer-enablement).
	if fee := sysparams.GetValoperRegisterFee(); fee > 0 {
		assertPaidCallIsDirect(0, cur)
		minFee := minFeeCoin(fee)
		sentCoins := unsafe.OriginSend()
		if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
			panic(ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom))
		}
	}

	// Check if the valoper is already registered.
	if isValoper(addr) {
		panic(ErrValoperExists)
	}

	// Reject disallowed key types early (else the EndBlocker drops it silently).
	assertPubKeyTypeAllowed(pubKey)

	// Derive the consensus signing address from the pubkey.
	signingAddr, err := chain.PubKeyAddress(pubKey)
	if err != nil {
		panic(err)
	}

	// Signing-address uniqueness across all profiles, ever.
	if signingRegistry.Has(signingAddr.String()) {
		panic(ErrSigningKeyTaken)
	}

	// Front-running guard: post-genesis, the signing address must
	// not already be an active validator.
	if runtime.ChainHeight() > 0 && validators.IsValidator(signingAddr) {
		panic(ErrFrontrunValidator)
	}

	v := Valoper{
		Moniker:            moniker,
		Description:        description,
		ServerType:         serverType,
		OperatorAddress:    addr,
		SigningPubKey:      pubKey,
		SigningAddress:     signingAddr,
		LastRotationHeight: runtime.ChainHeight(),
		KeepRunning:        true,
		auth:               authorizable.New(ownable.NewWithAddress(addr)),
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}

	// Save the valoper to the set.
	valopers.Set(v.OperatorAddress.String(), v)

	// Insert into the signing-address registry.
	signingRegistry.Set(signingAddr.String(), regEntry{
		OperatorAddress:    addr,
		RegisteredAtHeight: runtime.ChainHeight(),
		RetiredAtHeight:    0,
	})

	// Refresh v0's cache for this operator.
	validators.NotifyValoperChanged(cross(cur), addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// UpdateMoniker updates an existing valoper's moniker.
func UpdateMoniker(cur realm, addr address, moniker string) {
	// Check that the moniker is not empty.
	if err := validateMoniker(moniker); err != nil {
		panic(err)
	}

	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.auth.AssertPreviousOnAuthList(0, cur)

	// Update the moniker.
	v.Moniker = moniker

	// Save the valoper info.
	valopers.Set(addr.String(), v)
}

// UpdateDescription updates an existing valoper's description.
func UpdateDescription(cur realm, addr address, description string) {
	// Check that the description is not empty.
	if err := validateDescription(description); err != nil {
		panic(err)
	}

	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.auth.AssertPreviousOnAuthList(0, cur)

	// Update the description.
	v.Description = description

	// Save the valoper info.
	valopers.Set(addr.String(), v)
}

// UpdateKeepRunning updates an existing valoper's active status.
// Calls v0.NotifyValoperChanged because the cache stores KeepRunning.
func UpdateKeepRunning(cur realm, addr address, keepRunning bool) {
	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.auth.AssertPreviousOnAuthList(0, cur)

	// Update status.
	v.KeepRunning = keepRunning

	// Save the valoper info.
	valopers.Set(addr.String(), v)

	// Refresh v0's cache (KeepRunning is one of the cached fields).
	validators.NotifyValoperChanged(cross(cur), addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// UpdateServerType updates an existing valoper's server type.
func UpdateServerType(cur realm, addr address, serverType string) {
	// Check that the server type is valid.
	if err := validateServerType(serverType); err != nil {
		panic(err)
	}

	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.auth.AssertPreviousOnAuthList(0, cur)

	// Update server type.
	v.ServerType = serverType

	// Save the valoper info.
	valopers.Set(addr.String(), v)
}

// UpdateSigningKey rotates an operator's consensus signing key.
//
// Auth: caller must be on the operator's auth list (defaults to
// operator at Register time; extendable via AddToAuthList).
//
// Invariants checked at entry:
//   - throttle: ChainHeight() - v.LastRotationHeight >=
//     rotationPeriodBlocks
//   - signingRegistry uniqueness: derived(newPubKey) not in registry
//     (active OR retired); permanently blocks key reuse
//   - fee: unsafe.OriginSend() >= rotationFee (mirrors Register's
//     fee-check pattern)
//
// Effect: profile's SigningPubKey/SigningAddress/LastRotationHeight
// updated; old registry entry marked retired (retiredAtHeight =
// ChainHeight()); new entry inserted into signingRegistry; v0 emits
// remove+add to sysparams via RotateValoperSigningKey; v0 cache
// refreshed via NotifyValoperChanged. Rotation lands in consensus
// at H+2.
//
// Atomicity: Gno tx atomicity rolls back all state if any step
// panics. If v0.RotateValoperSigningKey panics, the registry insert
// and profile mutation revert with it.
func UpdateSigningKey(cur realm, addr address, newPubKey string) {
	v := GetByAddr(addr)

	// Auth: caller must be on operator's auth list.
	v.auth.AssertPreviousOnAuthList(0, cur)

	// Throttle: limit one rotation per rotation_period_blocks per
	// operator (per profile, not per caller — multi-member auth lists
	// can't multiplicative-rotate).
	height := runtime.ChainHeight()
	if height-v.LastRotationHeight < sysparams.GetValoperRotationPeriodBlocks() {
		panic(ErrRotationThrottled)
	}

	// Fee: enforce only if non-zero (matches Register's pattern;
	// rotation_fee defaults to zero pre-transfer-enablement).
	if fee := sysparams.GetValoperRotationFee(); fee > 0 {
		assertPaidCallIsDirect(0, cur)
		minFee := minFeeCoin(fee)
		sentCoins := unsafe.OriginSend()
		if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
			panic(ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom))
		}
	}

	// Reject disallowed key types early (else the EndBlocker drops it silently).
	assertPubKeyTypeAllowed(newPubKey)

	// Derive the new signing address from the new pubkey.
	newSigningAddr, err := chain.PubKeyAddress(newPubKey)
	if err != nil {
		panic(err)
	}

	// signingRegistry uniqueness: new key must not have ever been
	// registered (active or retired).
	if signingRegistry.Has(newSigningAddr.String()) {
		panic(ErrSigningKeyTaken)
	}

	// Front-running guard: the derived signing address must not already
	// be an active validator. Mirrors the same guard in Register
	// (ErrFrontrunValidator). signingRegistry uniqueness above only
	// blocks signing addresses that previously went through Register or
	// UpdateSigningKey — genesis-seeded validators bypassed both, so
	// their signing addresses are absent from signingRegistry. Without
	// this check, a valoper could rotate onto such a slot and hijack
	// it: v0.RotateValoperSigningKey would overwrite the active entry
	// with this operator's claim, and a subsequent govDAO remove-op
	// proposal would then delete it.
	if validators.IsValidator(newSigningAddr) {
		panic(ErrFrontrunValidator)
	}

	// Remember the previous signing key for the v0 cross-call.
	oldPubKey := v.SigningPubKey
	oldSigningAddr := v.SigningAddress

	// Mark the old registry entry retired. The entry must exist —
	// it was inserted at Register time.
	rawOld := signingRegistry.Get(oldSigningAddr.String())
	if rawOld == nil {
		panic(ErrRegistryEntryMissing)
	}
	oldEntry := rawOld.(regEntry)
	oldEntry.RetiredAtHeight = height
	signingRegistry.Set(oldSigningAddr.String(), oldEntry)

	// Insert the new entry as active.
	signingRegistry.Set(newSigningAddr.String(), regEntry{
		OperatorAddress:    addr,
		RegisteredAtHeight: height,
		RetiredAtHeight:    0,
	})

	// Update the profile.
	v.SigningPubKey = newPubKey
	v.SigningAddress = newSigningAddr
	v.LastRotationHeight = height
	valopers.Set(addr.String(), v)

	// Apply to consensus via v0, then refresh v0's cache view of the
	// profile. Order matters only in that both must complete; tx
	// atomicity rolls back together on any panic.
	validators.RotateValoperSigningKey(cross(cur), addr, oldPubKey, newPubKey)
	validators.NotifyValoperChanged(cross(cur), addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// GetByAddr fetches the valoper using the operator address, if present.
func GetByAddr(addr address) Valoper {
	valoperRaw := valopers.Get(addr.String())
	if valoperRaw == nil {
		panic(ErrValoperMissing)
	}

	return valoperRaw.(Valoper)
}

// Render renders the current valoper set.
// "/r/gnops/valopers" lists all valopers, paginated.
// "/r/gnops/valopers:addr" shows the detail for the valoper with the addr.
func Render(fullPath string) string {
	req := realmpath.Parse(fullPath)
	if req.Path == "" {
		return renderHome(fullPath)
	} else {
		addr := req.Path
		if len(addr) < 2 || addr[:2] != "g1" {
			return "invalid address " + addr
		}
		valoperRaw := valopers.Get(addr)
		if valoperRaw == nil {
			return "unknown address " + addr
		}
		v := valoperRaw.(Valoper)
		return "Valoper's details:\n" + v.Render()
	}
}

func renderHome(path string) string {
	// if there are no valopers, display instructions
	if valopers.Size() == 0 {
		return ufmt.Sprintf("%s\n\nNo valopers to display.", instructions)
	}

	page := pager.NewPager(valopers, 50, false).MustGetPageByPath(path)

	output := ""

	// if we are on the first page, display instructions
	if page.PageNumber == 1 {
		output += ufmt.Sprintf("%s\n\n", instructions)
	}

	for _, item := range page.Items {
		v := item.Value.(Valoper)
		output += ufmt.Sprintf(" * [%s](/r/gnops/valopers:%s) - [profile](/r/demo/profile:u/%s)\n",
			v.Moniker, v.OperatorAddress, v.OperatorAddress)
	}

	output += "\n"
	output += page.Picker(path)
	return output
}

// Validate checks if the fields of the Valoper are valid.
func (v *Valoper) Validate() error {
	errs := &combinederr.CombinedError{}

	errs.Add(validateMoniker(v.Moniker))
	errs.Add(validateDescription(v.Description))
	errs.Add(validateServerType(v.ServerType))
	errs.Add(validateBech32(v.OperatorAddress))
	errs.Add(validatePubKey(v.SigningPubKey))

	if errs.Size() == 0 {
		return nil
	}

	return errs
}

// Render renders a single valoper with their information.
func (v Valoper) Render() string {
	output := ufmt.Sprintf("## %s\n", v.Moniker)

	if v.Description != "" {
		output += ufmt.Sprintf("%s\n\n", v.Description)
	}

	output += ufmt.Sprintf("- Operator Address: %s\n", v.OperatorAddress.String())
	output += ufmt.Sprintf("- Signing Address: %s\n", v.SigningAddress.String())
	output += ufmt.Sprintf("- Signing PubKey: %s\n", v.SigningPubKey)
	output += ufmt.Sprintf("- Server Type: %s\n\n", v.ServerType)
	output += ufmt.Sprintf("[Profile link](/r/demo/profile:u/%s)\n", v.OperatorAddress)

	return output
}

// isValoper checks if the valoper exists.
func isValoper(addr address) bool {
	return valopers.Has(addr.String())
}

// validateMoniker checks if the moniker is valid.
func validateMoniker(moniker string) error {
	if moniker == "" {
		return ErrInvalidMoniker
	}

	if len(moniker) > MonikerMaxLength {
		return ErrInvalidMoniker
	}

	if !validateMonikerRe.MatchString(moniker) {
		return ErrInvalidMoniker
	}

	return nil
}

// validateDescription checks if the description is valid.
func validateDescription(description string) error {
	if description == "" {
		return ErrInvalidDescription
	}

	if len(description) > DescriptionMaxLength {
		return ErrInvalidDescription
	}

	return nil
}

// validateBech32 checks if the value is a valid bech32 address.
func validateBech32(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidAddress
	}

	return nil
}

// validatePubKey checks if the public key is valid.
func validatePubKey(pubKey string) error {
	if _, _, err := bech32.DecodeNoLimit(pubKey); err != nil {
		return err
	}

	return nil
}

// assertPaidCallIsDirect requires the immediate caller to be a plain EOA,
// and must be paired with every unsafe.OriginSend() amount check in this
// realm. Called only when a fee is actually configured.
//
// unsafe.OriginSend() reports the transaction's declared send ENVELOPE, not
// what this realm received. Only a direct `maketx call` on valopers credits
// the envelope to this realm's address. For a `maketx run` the keeper sets
// pkgAddr := caller (gno.land/pkg/sdk/vm/keeper.go), so the coins move from
// the caller to the caller and never land anywhere at all, while
// OriginSend() still reports the full amount to us. See
// gno.land/adr/pr6062_payable_send_check.md: "MsgRun is exempt: the coins
// are moved from the caller to the caller, so nothing actually moves." So
// the amount check alone verifies intent, never receipt.
//
// Reported by @D4ryl00, with a local-chain reproduction:
// governance sets register_fee to 1000ugnot, an operator runs a script
// calling Register with `-send 1000ugnot`, registration succeeds, valopers
// receives zero, and the operator still holds the coins. No hostile script
// is needed — the envelope is free by construction. Both fees default to 0,
// so nothing was live.
//
// Why IsUserCall and not an address comparison: a run script's realm is
// address-INDISTINGUISHABLE from its user. Inside `main(cur realm)` the
// ephemeral realm's own address IS the caller's EOA address, so every
// address-based guard in this realm accepts a `maketx run` — Register's
// squat guard (OriginCaller == addr) and UpdateSigningKey's auth-list check
// (Previous().Address() on the list) both pass. Only the pkgPath differs,
// and IsUserCall() is the check that reads it. It is therefore the only
// PreviousRealm shape where the envelope is guaranteed to have landed here:
// it excludes intermediate code realms and user-run ephemeral realms alike.
// Same reasoning and same pairing as r/sys/namereg/v1.Register, which reads
// OriginSend for its anti-squatting payment; see the two-guard comment there.
//
// Deliberately inside the `fee > 0` branch rather than at the top of the
// function: while a fee is unset there is no payment to establish receipt
// of, and gating unconditionally would break the operator-authored
// `maketx run` flows that the squat guard is happy to accept. The
// restriction appears exactly when, and only when, money is involved.
//
// KNOWN LIMIT. AddToAuthList takes a plain `member address`, so a REALM
// can sit on an operator's auth list, and once rotation_fee is nonzero
// such a member can no longer drive UpdateSigningKey — it is not a user
// call. An EOA rotation bot is unaffected, which is the shape the
// auth-list docs describe ("HSM-bound or rotation-bot"), so nothing
// documented breaks; a realm-mediated rotation would.
//
// The alternative that would keep realms working is a true receipt check:
// track cumulative fees in realm state and require the realm's own
// balance to have risen by `fee` since the last one. That works for any
// caller shape, but it puts new persisted state and a banker into a
// genesis realm, and lets anyone pre-pay another operator's fee by
// donating to the realm address. Not worth it while both fees are 0 and
// collected fees are unwithdrawable anyway (this realm has no banker, so
// they strand at its address). Revisit if a realm-mediated paid rotation
// is ever actually wanted.
func assertPaidCallIsDirect(_ int, rlm realm) {
	// IsCurrent before Previous, as AGENTS.md requires and every other
	// (_ int, rlm realm) helper in this tree does (authorizable,
	// sys/params/delegate, gov/dao/types, sys/validators/v0/cache).
	// Both call sites pass their live cur, so this is defense-in-depth
	// against a future caller threading a stale or stashed realm value:
	// Previous() reads the prev field verbatim, so on a non-live token
	// it answers for the wrong frame.
	if !rlm.IsCurrent() {
		panic(ErrPaidCallNotDirect)
	}
	if !rlm.Previous().IsUserCall() {
		panic(ErrPaidCallNotDirect)
	}
}

// minFeeCoin converts a configured fee (uint64, from sysparams) into the
// ugnot Coin the amount checks compare against.
//
// Coin.Amount is an int64. A fee above MaxInt64 wraps negative, and
// `sentCoins[0].IsLT(minFee)` is then false for every payment — the realm
// reads the fee as configured, runs assertPaidCallIsDirect, and collects
// 1ugnot while reporting "payment must not be less than -1ugnot". Failing
// closed here keeps an unrepresentable fee from opening the paid path
// instead of closing it.
//
// Reachable two ways today, neither requiring malice: governance can set
// any uint64 through the generic r/sys/params factories, and
// valopers/proposal.ProposeNewMinFeeProposalRequest takes an int64 and
// stores uint64(newMinFee), so a negative fee — the obvious way to write
// "disable the fee" — round-trips to 2^64-1. That signature is preserved
// for historical-replay compatibility and is deliberately left alone; the
// bound belongs at the consumption site, which covers both paths.
func minFeeCoin(fee uint64) chain.Coin {
	if fee > math.MaxInt64 {
		panic(ErrFeeParamOutOfRange)
	}
	return chain.NewCoin("ugnot", int64(fee))
}

// assertPubKeyTypeAllowed panics if pubKey's type is not in the chain's validator allow-list (empty list accepts any).
func assertPubKeyTypeAllowed(pubKey string) {
	allowed := sysparams.GetValsetPubKeyTypes()
	if len(allowed) == 0 {
		return
	}
	typeURL, err := pubKeyTypeURL(pubKey)
	if err != nil {
		panic(err)
	}
	for _, a := range allowed {
		if a == typeURL {
			return
		}
	}
	panic(ErrDisallowedPubKeyType)
}

// pubKeyTypeURL returns the amino type URL (e.g. "/tm.PubKeyEd25519") of a bech32 consensus pubkey.
func pubKeyTypeURL(pubKey string) (string, error) {
	// gpub exceeds bech32's 90-char cap, so decode without the limit.
	_, data5, err := bech32.DecodeNoLimit(pubKey)
	if err != nil {
		return "", err
	}
	data, err := bech32.ConvertBits(data5, 5, 8, false)
	if err != nil {
		return "", err
	}
	// Type URL is the first amino field: 0x0A <len> <typeURL>.
	if len(data) < 2 || data[0] != 0x0A {
		return "", errors.New("malformed consensus pubkey: " + pubKey)
	}
	n := int(data[1])
	if n == 0 || len(data) < 2+n {
		return "", errors.New("malformed consensus pubkey: " + pubKey)
	}
	return string(data[2 : 2+n]), nil
}

// validateServerType checks if the server type is valid.
func validateServerType(serverType string) error {
	if serverType != ServerTypeCloud &&
		serverType != ServerTypeOnPrem &&
		serverType != ServerTypeDataCenter {
		return ErrInvalidServerType
	}

	return nil
}
