// Package valopers is designed around the permissionless lifecycle of valoper profiles.
package valopers

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"crypto/bech32"
	"errors"
	"regexp"

	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/combinederr/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/ownable/v0/exts/authorizable"
	"gno.land/p/nt/ufmt/v0"
	sysparams "gno.land/r/sys/params"
	validators "gno.land/r/sys/validators/v3"
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

func (v Valoper) Auth() *authorizable.Authorizable {
	return v.auth
}

func AddToAuthList(cur realm, addr address, member address) {
	v := GetByAddr(addr)
	if err := v.Auth().AddToAuthList(member); err != nil {
		panic(err)
	}
}

func DeleteFromAuthList(cur realm, addr address, member address) {
	v := GetByAddr(addr)
	if err := v.Auth().DeleteFromAuthList(member); err != nil {
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
// Why OriginCaller==addr is sufficient (no IsUserCall): squatting
// requires the attacker to be able to satisfy OriginCaller==victim,
// which requires the victim's signing key. r/sys/namereg/v1.Register
// also gates on IsUserCall, but that's because IT reads
// banker.OriginSend() for the anti-squatting payment and IsUserCall
// is needed to ensure the OriginSend envelope reflects what landed at
// this realm rather than a phantom payment from a previous frame.
// valopers.Register has no per-call payment-receipt check (fees are
// validated against banker.OriginSend in a way that's symmetric to
// IsUserCall via direct comparison), so the IsUserCall tightening
// would only block legitimate `maketx run` flows (operator-authored
// scripts that legitimately set OriginCaller==operator) without
// adding identity-squat protection.
//
// Auth-list seeding: the profile's Authorizable owner is set to addr
// (NOT OriginCaller). At H>0 the squat guard makes them equal anyway;
// at H==0 the deployer pattern (one signer registers many operators)
// requires owner == addr so each operator can manage their own profile
// post-genesis without needing the deployer's auth.
func Register(cur realm, moniker string, description string, serverType string, addr address, pubKey string) {
	// Operator-slot squat guard.
	if runtime.ChainHeight() > 0 && runtime.OriginCaller() != addr {
		panic(ErrOperatorSquatGuard)
	}

	// Fee enforcement (read from sysparams; defaults to 0 until
	// governance raises it post-transfer-enablement).
	if fee := sysparams.GetValoperRegisterFee(); fee > 0 {
		minFee := chain.NewCoin("ugnot", int64(fee))
		sentCoins := banker.OriginSend()
		if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
			panic(ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom))
		}
	}

	// Check if the valoper is already registered.
	if isValoper(addr) {
		panic(ErrValoperExists)
	}

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
		auth:               authorizable.New(ownable.NewWithAddressByPrevious(addr)),
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

	// Refresh v3's cache for this operator.
	validators.NotifyValoperChanged(cross, addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// UpdateMoniker updates an existing valoper's moniker.
func UpdateMoniker(cur realm, addr address, moniker string) {
	// Check that the moniker is not empty.
	if err := validateMoniker(moniker); err != nil {
		panic(err)
	}

	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.Auth().AssertPreviousOnAuthList()

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
	v.Auth().AssertPreviousOnAuthList()

	// Update the description.
	v.Description = description

	// Save the valoper info.
	valopers.Set(addr.String(), v)
}

// UpdateKeepRunning updates an existing valoper's active status.
// Calls v3.NotifyValoperChanged because the cache stores KeepRunning.
func UpdateKeepRunning(cur realm, addr address, keepRunning bool) {
	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.Auth().AssertPreviousOnAuthList()

	// Update status.
	v.KeepRunning = keepRunning

	// Save the valoper info.
	valopers.Set(addr.String(), v)

	// Refresh v3's cache (KeepRunning is one of the cached fields).
	validators.NotifyValoperChanged(cross, addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// UpdateServerType updates an existing valoper's server type.
func UpdateServerType(cur realm, addr address, serverType string) {
	// Check that the server type is valid.
	if err := validateServerType(serverType); err != nil {
		panic(err)
	}

	v := GetByAddr(addr)

	// Check that the caller has permissions.
	v.Auth().AssertPreviousOnAuthList()

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
//   - fee: banker.OriginSend() >= rotationFee (mirrors Register's
//     fee-check pattern)
//
// Effect: profile's SigningPubKey/SigningAddress/LastRotationHeight
// updated; old registry entry marked retired (retiredAtHeight =
// ChainHeight()); new entry inserted into signingRegistry; v3 emits
// remove+add to sysparams via RotateValoperSigningKey; v3 cache
// refreshed via NotifyValoperChanged. Rotation lands in consensus
// at H+2.
//
// Atomicity: Gno tx atomicity rolls back all state if any step
// panics. If v3.RotateValoperSigningKey panics, the registry insert
// and profile mutation revert with it.
func UpdateSigningKey(cur realm, addr address, newPubKey string) {
	v := GetByAddr(addr)

	// Auth: caller must be on operator's auth list.
	v.Auth().AssertPreviousOnAuthList()

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
		minFee := chain.NewCoin("ugnot", int64(fee))
		sentCoins := banker.OriginSend()
		if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
			panic(ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom))
		}
	}

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

	// Remember the previous signing key for the v3 cross-call.
	oldPubKey := v.SigningPubKey
	oldSigningAddr := v.SigningAddress

	// Mark the old registry entry retired. The entry must exist —
	// it was inserted at Register time.
	rawOld, ok := signingRegistry.Get(oldSigningAddr.String())
	if !ok {
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

	// Apply to consensus via v3, then refresh v3's cache view of the
	// profile. Order matters only in that both must complete; tx
	// atomicity rolls back together on any panic.
	validators.RotateValoperSigningKey(cross, addr, oldPubKey, newPubKey)
	validators.NotifyValoperChanged(cross, addr, v.SigningPubKey, v.SigningAddress, v.KeepRunning)
}

// GetByAddr fetches the valoper using the operator address, if present.
func GetByAddr(addr address) Valoper {
	valoperRaw, exists := valopers.Get(addr.String())
	if !exists {
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
		valoperRaw, exists := valopers.Get(addr)
		if !exists {
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
	_, exists := valopers.Get(addr.String())

	return exists
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

// validateServerType checks if the server type is valid.
func validateServerType(serverType string) error {
	if serverType != ServerTypeCloud &&
		serverType != ServerTypeOnPrem &&
		serverType != ServerTypeDataCenter {
		return ErrInvalidServerType
	}

	return nil
}
