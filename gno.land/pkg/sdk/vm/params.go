package vm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	sdkparams "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// CodeSubmissionPolicy controls who may submit MsgAddPackage, and how submitted
// packages are processed. It has no bearing on MsgRun: that is gated by
// RunSubmitters under every policy value, because no policy value makes running
// arbitrary source safe. See the RunSubmitters field doc.
type CodeSubmissionPolicy string

const (
	// CodeSubmissionPolicyPermissionless allows any address to submit code (default).
	CodeSubmissionPolicyPermissionless CodeSubmissionPolicy = "permissionless"
	// CodeSubmissionPolicyPermissioned restricts code submission to addresses
	// listed in Params.CodeSubmitters.
	CodeSubmissionPolicyPermissioned CodeSubmissionPolicy = "permissioned"
	// CodeSubmissionPolicyInert accepts packages from any address but stores them
	// without typechecking or execution (inert state). Packages become callable
	// only after an approver sends MsgEnablePackage.
	CodeSubmissionPolicyInert CodeSubmissionPolicy = "inert"
)

const (
	sysNamesPkgDefault              = "gno.land/r/sys/names"
	sysCLAPkgDefault                = "gno.land/r/sys/cla"
	chainDomainDefault              = "gno.land"
	depositDefault                  = "100000000ugnot" // 1 MB of realm state at storagePriceDefault
	storagePriceDefault             = "100ugnot"       // cost per byte (1 gnot per 10KB) 1.333B GNOT == 13.33TB
	storageFeeCollectorNameDefault  = "storage_fee_collector"
	codeSubmissionPolicyDefault     = CodeSubmissionPolicyPermissionless
	inertChargeCollectorNameDefault = "inert_charge_collector"
	// maxInertSubmissionCharge caps InertSubmissionCharge. A charge governance
	// can raise without limit is a deploy freeze, which is the outcome the
	// charge exists to prevent. 1000 GNOT is far above any plausible setting
	// (a block of gas is worth ~3 GNOT at the initial price) and far below a
	// figure that would price deploys out.
	maxInertSubmissionCharge = int64(1_000_000_000_000) // 1000 GNOT in ugnot

	// Depth pins for the reference store: B+32 mounted with the fast index
	// (storebptree.FastStoreConstructor), calibrated at 100M items, 10K node
	// cache, batched 1000-mutation blocks (NewParams pins Fixed = Min, so
	// these are charged exactly, at every tree size). GET is one flat DB
	// read via the fast index, size-independent by construction.
	// SET-read/WRITE are pinned at the measured-with-cache costs rather than
	// priced by the store's live estimator (Fixed=0), which ignores the node
	// LRU and overcharges mid-range sizes ~2× — revisit once it is
	// cache-aware. Until then the pins drift gradually underpriced past the
	// calibration point (write pin ~-13% at 1.6G keys) and are re-tuned by
	// governance. Provenance: tm2/pkg/bptree/PERFORMANCE.md; rationale and
	// accepted imprecisions: gno.land/adr/pr5938_mount_bptree_store.md.
	//
	// Changing these defaults requires a new legacy fingerprint in
	// contribs/gnogenesis/internal/fork/generate.go (fork repricing).
	minGetReadDepth100Default = int64(100) // 1.0 flat read (fast-index hit)
	minSetReadDepth100Default = int64(200) // 2.0 SET read ops (descent, measured with 10K cache)
	minWriteDepth100Default   = int64(540) // 4.4 batched COW writes + 1.0 fast-index write
	// Iterator step flat cost; mirrors store.DefaultGasConfig().IterNextCostFlat.
	iterNextCostFlatDefault = 1_000
	// PreprocessGasPerByte: gas charged per .gno source byte at MsgAddPackage
	// and MsgRun for the native type-check + preprocess passes, which are
	// otherwise unmetered. Measured ~1250 gas/byte on realistic example
	// realms/packages (type-check ~920 + preprocess ~330, own compile only —
	// dependency re-type-checking excluded; 1 gas == 1ns on the reference
	// machine, and the host-machine calibration factor is the dominant
	// uncertainty).
	preprocessGasPerByteDefault = int64(1_250)
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the vm module.
type Params struct {
	SysNamesPkgPath      string         `json:"sysnames_pkgpath" yaml:"sysnames_pkgpath"`
	SysCLAPkgPath        string         `json:"syscla_pkgpath" yaml:"syscla_pkgpath"`
	ChainDomain          string         `json:"chain_domain" yaml:"chain_domain"`
	DefaultDeposit       string         `json:"default_deposit" yaml:"default_deposit"`
	StoragePrice         string         `json:"storage_price" yaml:"storage_price"`
	StorageFeeCollector  crypto.Address `json:"storage_fee_collector" yaml:"storage_fee_collector"`
	MinGetReadDepth100   int64          `json:"min_get_read_depth_100" yaml:"min_get_read_depth_100"`
	MinSetReadDepth100   int64          `json:"min_set_read_depth_100" yaml:"min_set_read_depth_100"`
	MinWriteDepth100     int64          `json:"min_write_depth_100" yaml:"min_write_depth_100"`
	FixedGetReadDepth100 int64          `json:"fixed_get_read_depth_100" yaml:"fixed_get_read_depth_100"`
	FixedSetReadDepth100 int64          `json:"fixed_set_read_depth_100" yaml:"fixed_set_read_depth_100"`
	FixedWriteDepth100   int64          `json:"fixed_write_depth_100" yaml:"fixed_write_depth_100"`
	// IterNextCostFlat must be > 0; Validate rejects zero. Asymmetric
	// with the six depth fields above (where 0 legitimately means
	// "no floor / use tree estimate") because zero iter-step cost
	// would effectively disable iteration gas charging.
	IterNextCostFlat int64 `json:"iter_next_cost_flat" yaml:"iter_next_cost_flat"`
	// PreprocessGasPerByte must be > 0; Validate rejects zero (like
	// IterNextCostFlat) because a zero per-byte cost would silently disable
	// the type-check/preprocess gas charge. Charged per .gno source byte at
	// MsgAddPackage and MsgRun; it is NOT part of store.GasConfig (charged
	// directly in the keeper), so it has no ApplyToGasConfig entry.
	//
	// Kept immediately after IterNextCostFlat (amino field 14) so its wire
	// number matches master; the code-submission fields below take 15/16/17.
	PreprocessGasPerByte int64 `json:"preprocess_gas_per_byte" yaml:"preprocess_gas_per_byte"`

	// CodeSubmissionPolicy controls who may submit MsgAddPackage and how
	// packages are processed on arrival. Defaults to "permissionless".
	// Not consulted for MsgRun; see RunSubmitters.
	CodeSubmissionPolicy CodeSubmissionPolicy `json:"code_submission_policy" yaml:"code_submission_policy"`
	// CodeSubmitters is the allowlist used when CodeSubmissionPolicy == "permissioned".
	CodeSubmitters []crypto.Address `json:"code_submitters" yaml:"code_submitters"`
	// PkgApprovers may call MsgEnablePackage / MsgDisablePackage.
	// Required when CodeSubmissionPolicy == "inert".
	PkgApprovers []crypto.Address `json:"pkg_approvers" yaml:"pkg_approvers"`
	// RunSubmitters may send MsgRun. Consulted under EVERY CodeSubmissionPolicy,
	// deliberately unlike CodeSubmitters: MsgRun type-checks and executes
	// arbitrary source immediately, under every policy including "inert", so the
	// policy value has no bearing on the hazard. It is also the only code-bearing
	// message with no other gate — MsgAddPackage clears a namespace check and a
	// CLA check, while MsgRun's path is forced to /e/<caller>/run and so has no
	// namespace to check against.
	//
	// EMPTY MEANS OFF: anyone may send MsgRun, which is the behaviour that
	// predates this field. Listing one address turns the gate on.
	//
	// The asymmetry with CodeSubmitters is deliberate. CodeSubmitters is read
	// only after CodeSubmissionPolicy has been explicitly moved to
	// "permissioned", so its empty state is a half-finished opt-in and refusing
	// is the safe reading. RunSubmitters has no such switch — it is read on
	// every MsgRun from the moment the field exists — so treating empty as
	// "nobody" would disable MsgRun on every chain that upgrades without editing
	// genesis. Because GovDAO proposal creation is MsgRun-only (a
	// ProposalRequest carries a func value, which MsgCall cannot marshal), that
	// would take governance with it and leave no in-band repair.
	//
	// Kept as its own list rather than reusing CodeSubmitters, because reuse
	// would make one list mean different things per policy — and an operator who
	// populated it just to gate MsgRun would silently grant deploy rights the
	// moment governance flipped the policy to "permissioned".
	RunSubmitters []crypto.Address `json:"run_submitters" yaml:"run_submitters"`

	// InertSubmissionCharge is taken from the creator on every MsgAddPackage
	// that parks under the "inert" policy, and paid to InertChargeCollector.
	//
	// Under "inert" the submitter chooses the work and an unattended approval
	// oracle pays for it: MsgEnablePackage runs their init() on ITS transaction
	// and ITS gas meter. Since fees are flat, the oracle's exposure is its fee
	// times the number of approvals it can be induced to make, and it stops
	// approving for everyone once its spend limit is reached. Submitting is
	// close to free today, so that is cheap to provoke. A charge at submit
	// prices it: the attacker pays per approval they induce, and the payer is
	// the party who chose the cost.
	//
	// It is a flat amount and NOT derived from the gas price, deliberately.
	// LastGasPrice is recomputed only in auth.EndBlocker and InitChain runs no
	// EndBlock, so a price-derived amount would collect something different
	// during a fork's replay than the source chain collected, and balances would
	// drift silently. A literal amount replays correctly because params are
	// store-backed: a governance transaction that changed it re-applies at its
	// own point in the replayed history.
	//
	// EMPTY MEANS OFF, and this is what makes replay correct without a carve-out.
	// applyLegacyDefaults must never fill it, and neither must `gnogenesis fork
	// generate` — a chain whose history predates the field then replays
	// charge-free, rather than being charged for submissions that never paid.
	// Governance turns it on going forward. A fork adopting it must do so with a
	// migration tx appended AFTER the history, never in the fork's genesis
	// params, for the same reason CodeSubmissionPolicy must (see AddPackage).
	//
	// Capped by maxInertSubmissionCharge. Without a ceiling, governance could
	// price deploys out of reach entirely, which is the outcome the charge
	// exists to prevent.
	InertSubmissionCharge string `json:"inert_submission_charge" yaml:"inert_submission_charge"`
	// InertChargeCollector receives InertSubmissionCharge. Validate does not look
	// at this field. Two other things keep a zero address away from the charge:
	// applyLegacyDefaults fills a derived placeholder, and the charge site skips
	// a zero collector rather than pay it. Both are load-bearing — neither is a
	// redundant check on the other — and together they make the "charge set,
	// collector unset" combination unrepresentable, rather than a cross-field
	// rule doing it. Cross-field validation on Params was tried and
	// reverted (see the ADR): WillSetParam re-validates the whole struct and
	// PANICS, and r/sys/params sets one key per proposal, so a rule spanning two
	// fields aborts a proposal that already passed its vote.
	//
	// The intended value is an approver treasury, so that submissions fund the
	// approvals they cause. Note that this is a governance convention and not a
	// property of the mechanism: nothing can tell a spendable treasury from a
	// derived address with no private key, and the default IS such an address.
	// Set it before turning the charge on, or the charge is burned.
	InertChargeCollector crypto.Address `json:"inert_charge_collector" yaml:"inert_charge_collector"`
}

// NewParams creates a new Params object
func NewParams(namesPkgPath, claPkgPath, chainDomain, defaultDeposit, storagePrice string, storageFeeCollector crypto.Address, minGetReadDepth100, minSetReadDepth100, minWriteDepth100, iterNextCostFlat, preprocessGasPerByte int64) Params {
	return Params{
		SysNamesPkgPath:      namesPkgPath,
		SysCLAPkgPath:        claPkgPath,
		ChainDomain:          chainDomain,
		DefaultDeposit:       defaultDeposit,
		StoragePrice:         storagePrice,
		StorageFeeCollector:  storageFeeCollector,
		MinGetReadDepth100:   minGetReadDepth100,
		MinSetReadDepth100:   minSetReadDepth100,
		MinWriteDepth100:     minWriteDepth100,
		FixedGetReadDepth100: minGetReadDepth100,
		FixedSetReadDepth100: minSetReadDepth100,
		FixedWriteDepth100:   minWriteDepth100,
		IterNextCostFlat:     iterNextCostFlat,
		CodeSubmissionPolicy: codeSubmissionPolicyDefault,
		PreprocessGasPerByte: preprocessGasPerByte,
		// InertSubmissionCharge is deliberately left empty: off by default.
		// The collector still gets a value, so that Validate can reject the zero
		// address unconditionally and the "charge set, collector unset"
		// combination never has to be expressed as a cross-field rule.
		InertChargeCollector: crypto.AddressFromPreimage([]byte(inertChargeCollectorNameDefault)),
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysNamesPkgDefault, sysCLAPkgDefault, chainDomainDefault,
		depositDefault, storagePriceDefault, crypto.AddressFromPreimage([]byte(storageFeeCollectorNameDefault)),
		minGetReadDepth100Default, minSetReadDepth100Default, minWriteDepth100Default,
		iterNextCostFlatDefault, preprocessGasPerByteDefault)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysNamesPkgPath))
	sb.WriteString(fmt.Sprintf("SysCLAPkgPath: %q\n", p.SysCLAPkgPath))
	sb.WriteString(fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain))
	sb.WriteString(fmt.Sprintf("DefaultDeposit: %q\n", p.DefaultDeposit))
	sb.WriteString(fmt.Sprintf("StoragePrice: %q\n", p.StoragePrice))
	sb.WriteString(fmt.Sprintf("StorageFeeCollector: %q\n", p.StorageFeeCollector.String()))
	sb.WriteString(fmt.Sprintf("MinGetReadDepth100: %d\n", p.MinGetReadDepth100))
	sb.WriteString(fmt.Sprintf("MinSetReadDepth100: %d\n", p.MinSetReadDepth100))
	sb.WriteString(fmt.Sprintf("MinWriteDepth100: %d\n", p.MinWriteDepth100))
	sb.WriteString(fmt.Sprintf("FixedGetReadDepth100: %d\n", p.FixedGetReadDepth100))
	sb.WriteString(fmt.Sprintf("FixedSetReadDepth100: %d\n", p.FixedSetReadDepth100))
	sb.WriteString(fmt.Sprintf("FixedWriteDepth100: %d\n", p.FixedWriteDepth100))
	sb.WriteString(fmt.Sprintf("IterNextCostFlat: %d\n", p.IterNextCostFlat))
	sb.WriteString(fmt.Sprintf("CodeSubmissionPolicy: %q\n", p.CodeSubmissionPolicy))
	sb.WriteString(fmt.Sprintf("CodeSubmitters: %v\n", p.CodeSubmitters))
	sb.WriteString(fmt.Sprintf("PkgApprovers: %v\n", p.PkgApprovers))
	sb.WriteString(fmt.Sprintf("RunSubmitters: %v\n", p.RunSubmitters))
	sb.WriteString(fmt.Sprintf("InertSubmissionCharge: %q\n", p.InertSubmissionCharge))
	sb.WriteString(fmt.Sprintf("InertChargeCollector: %q\n", p.InertChargeCollector.String()))
	sb.WriteString(fmt.Sprintf("PreprocessGasPerByte: %d\n", p.PreprocessGasPerByte))
	return sb.String()
}

func (p Params) Validate() error {
	if p.SysNamesPkgPath != "" && !gno.IsUserlib(p.SysNamesPkgPath) {
		return fmt.Errorf("invalid user package path %q", p.SysNamesPkgPath)
	}
	if p.SysCLAPkgPath != "" && !gno.IsUserlib(p.SysCLAPkgPath) {
		return fmt.Errorf("invalid CLA package path %q", p.SysCLAPkgPath)
	}
	if p.ChainDomain != "" && !ASCIIDomain.MatchString(p.ChainDomain) {
		return fmt.Errorf("invalid chain domain %q, failed to match %q", p.ChainDomain, ASCIIDomain)
	}
	// Both are read back with std.MustParseCoin, which takes one coin and panics
	// on anything else, and both are then used as ugnot amounts. So accepting a
	// set here would let a parameter change through that panics in the storage
	// deposit path, or that is silently spent as ugnot under another denom's name.
	coins, err := std.ParseCoins(p.DefaultDeposit)
	if err != nil || len(coins) != 1 || coins[0].Denom != ugnot.Denom {
		return fmt.Errorf("invalid default storage deposit %q, want a single %s amount",
			p.DefaultDeposit, ugnot.Denom)
	}
	coins, err = std.ParseCoins(p.StoragePrice)
	if err != nil || len(coins) != 1 || coins[0].Denom != ugnot.Denom {
		return fmt.Errorf("invalid storage price %q, want a single %s amount",
			p.StoragePrice, ugnot.Denom)
	}
	if p.StorageFeeCollector.IsZero() {
		return fmt.Errorf("invalid storage fee collector, cannot be empty")
	}
	// Empty is off, and is the only spelling of off: ParseCoins("") returns no
	// error and no coins, while "0ugnot" fails Coins.validate, so a zero amount
	// cannot be smuggled through as a second spelling.
	if p.InertSubmissionCharge != "" {
		coins, err = std.ParseCoins(p.InertSubmissionCharge)
		if err != nil || len(coins) != 1 || coins[0].Denom != ugnot.Denom {
			return fmt.Errorf(
				"invalid inert submission charge %q, want a single %s amount",
				p.InertSubmissionCharge, ugnot.Denom)
		}
		if coins[0].Amount > maxInertSubmissionCharge {
			return fmt.Errorf("inert submission charge must be <= %d%s, got %d",
				maxInertSubmissionCharge, ugnot.Denom, coins[0].Amount)
		}
	}
	// The collector is deliberately NOT validated here, in either form.
	//
	// A cross-field rule ("charge set implies collector set") would abort a
	// governance proposal that already passed its vote: WillSetParam
	// re-validates the whole struct and panics, and r/sys/params sets one key
	// per proposal, so "charge set, collector unset" is an unavoidable
	// intermediate state.
	//
	// An unconditional non-zero rule fails every path that builds Params without
	// going through applyLegacyDefaults -- `gnogenesis fork generate` does
	// exactly that, and would refuse to produce a fork genesis.
	//
	// applyLegacyDefaults supplies the collector on the one read path the keeper
	// uses, and AddPackage skips the charge entirely if it is somehow still zero,
	// so a misconfiguration costs nothing rather than burning coins at the zero
	// address.
	// Depth floors / overrides are 100x fixed-point. The cap is 10_000
	// (= 100 tree levels), well beyond any plausible B+tree / IAVL
	// depth. Upper bound prevents a governance proposal from setting
	// absurd values that would trip overflow.Mulp in the cache.Store
	// gas calculation and silently brick writes. Zero remains
	// legitimate (no floor / use tree estimate); negative is rejected
	// because downstream `> 0` guards would make it a silent no-op.
	const maxDepth100 = int64(10_000)
	for _, f := range []struct {
		name string
		v    int64
	}{
		{"MinGetReadDepth100", p.MinGetReadDepth100},
		{"MinSetReadDepth100", p.MinSetReadDepth100},
		{"MinWriteDepth100", p.MinWriteDepth100},
		{"FixedGetReadDepth100", p.FixedGetReadDepth100},
		{"FixedSetReadDepth100", p.FixedSetReadDepth100},
		{"FixedWriteDepth100", p.FixedWriteDepth100},
	} {
		if f.v < 0 {
			return fmt.Errorf("%s must be non-negative, got %d", f.name, f.v)
		}
		if f.v > maxDepth100 {
			return fmt.Errorf("%s must be <= %d, got %d", f.name, maxDepth100, f.v)
		}
	}
	// IterNextCostFlat is a raw gas amount per iterator step. Cap at
	// 100_000 — 100x the tm2 default of 1_000, well above any
	// realistic per-step cost while far below the block gas limit
	// (~3B) so that a single adversarial proposal can't make one
	// iterator step drain an entire block's gas budget.
	const maxIterNextCostFlat = 100_000
	if p.IterNextCostFlat <= 0 {
		return fmt.Errorf("IterNextCostFlat must be positive, got %d", p.IterNextCostFlat)
	}
	if p.IterNextCostFlat > maxIterNextCostFlat {
		return fmt.Errorf("IterNextCostFlat must be <= %d, got %d", maxIterNextCostFlat, p.IterNextCostFlat)
	}
	switch p.CodeSubmissionPolicy {
	case CodeSubmissionPolicyPermissionless, CodeSubmissionPolicyPermissioned,
		CodeSubmissionPolicyInert:
		// valid
	case "":
		// treat empty as permissionless (zero-value compat)
	default:
		return fmt.Errorf("invalid code_submission_policy %q", p.CodeSubmissionPolicy)
	}
	if err := validateAddressSlice("CodeSubmitters", p.CodeSubmitters); err != nil {
		return err
	}
	if err := validateAddressSlice("RunSubmitters", p.RunSubmitters); err != nil {
		return err
	}
	if err := validateAddressSlice("PkgApprovers", p.PkgApprovers); err != nil {
		return err
	}
	// Cap PreprocessGasPerByte at 100_000 (80x the default, far above the
	// measured cost) to give governance headroom while preventing an absurd
	// proposal from making deploys impossibly expensive.
	const maxPreprocessGasPerByte = 100_000
	if p.PreprocessGasPerByte <= 0 {
		return fmt.Errorf("PreprocessGasPerByte must be positive, got %d", p.PreprocessGasPerByte)
	}
	if p.PreprocessGasPerByte > maxPreprocessGasPerByte {
		return fmt.Errorf("PreprocessGasPerByte must be <= %d, got %d", maxPreprocessGasPerByte, p.PreprocessGasPerByte)
	}
	return nil
}

// mustParseAddressStrings decodes a repeated (string-array) governance param
// into typed addresses. Both code_submitters and pkg_approvers are set via the
// strings param path (params.NewSysParamStringsPropRequest / SetStrings); the
// keeper stores the raw string array and GetParams decodes it element-wise back
// into the typed []crypto.Address field. Validation is strict — no trimming, no
// skipping of empty entries — precisely so this matches what GetParams later
// decodes: any entry accepted here must round-trip, otherwise a value could
// pass validation yet make every subsequent GetParams panic. A comma-separated
// single string does NOT round-trip and is unsupported.
//
// This mirrors the convention introduced for code_submitters in Phase 1
// (#5885), so the two changes compose without divergence.
func mustParseAddressStrings(paramName string, value any) []crypto.Address {
	ss := sdkparams.MustParamStrings(paramName, value)
	addrs := make([]crypto.Address, 0, len(ss))
	for _, s := range ss {
		addr, err := crypto.AddressFromString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid %s address %q: %v", paramName, s, err))
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

// maxAddressListLen bounds every governance address list in Params.
//
// These lists are not read lazily. gno.land's ante closure calls
// vmk.GetParams(ctx) on EVERY transaction, before auth's SetGasMeter installs
// the per-tx meter — so in DeliverTx the decode is charged to a passthrough
// bounded by remaining BLOCK gas rather than the sender's GasWanted, and in
// CheckTx to an infinite meter with no block-gas bound at all, re-run on every
// mempool recheck. GetParams does one store read plus one amino JSON unmarshal
// per field, and unmarshalling []crypto.Address bech32-decodes every element.
//
// So an unbounded list is unmetered per-transaction work that the party who
// caused it does not pay for: at roughly 43 bytes per entry, 10k entries is
// ~430KB decoded and 10k bech32 decodes on every transaction, forever, for a
// one-off cost of about 43 GNOT in storage deposit. Gas is not the binding
// constraint — you would need hundreds of MB to exhaust a block.
//
// 1000 is chosen to be generous for real operation (~43KB, well under a
// millisecond) while keeping the ceiling finite. Do not raise it casually and
// do not lower it later: a chain whose genesis already exceeds a smaller cap
// would fail to import. Enforced here rather than in the realm because the
// realm is replaceable state, while this covers governance, genesis and
// gnogenesis in one place.
const maxAddressListLen = 1000

func validateAddressSlice(name string, addrs []crypto.Address) error {
	if len(addrs) > maxAddressListLen {
		return fmt.Errorf("%s has %d entries, exceeding the maximum of %d",
			name, len(addrs), maxAddressListLen)
	}
	seen := make(map[string]struct{}, len(addrs))
	for i, addr := range addrs {
		if addr.IsZero() {
			return fmt.Errorf("%s[%d] is a zero address", name, i)
		}
		key := addr.String()
		if _, dup := seen[key]; dup {
			return fmt.Errorf("%s contains duplicate address %s", name, key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
}

// ApplyToGasConfig overwrites the governed gas fields of cfg with
// the values in p. Shared by the anteHandler (tx path) and
// newGnoTransactionStore (query path) so the two never drift.
//
// Every write is unconditional. For the six depth fields that's safe
// because store.DefaultGasConfig() initializes them to 0 and zero is
// a legitimate value ("no floor / use tree estimate"); overwriting
// 0 with 0 is a no-op. IterNextCostFlat is required to be positive
// (Validate rejects zero), so a Params that reached this code is
// guaranteed to have a meaningful value for it.
func (p Params) ApplyToGasConfig(cfg *store.GasConfig) {
	cfg.MinGetReadDepth100 = p.MinGetReadDepth100
	cfg.MinSetReadDepth100 = p.MinSetReadDepth100
	cfg.MinWriteDepth100 = p.MinWriteDepth100
	cfg.FixedGetReadDepth100 = p.FixedGetReadDepth100
	cfg.FixedSetReadDepth100 = p.FixedSetReadDepth100
	cfg.FixedWriteDepth100 = p.FixedWriteDepth100
	cfg.IterNextCostFlat = p.IterNextCostFlat
}

func (vm *VMKeeper) SetParams(ctx sdk.Context, params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	vm.prmk.SetStruct(ctx, "vm:p", params) // prmk is root.
	return nil
}

// applyLegacyDefaults fills fields absent from a params blob written before
// they existed (GetStruct leaves an absent key zero; a genesis predating the
// field decodes it as zero too). A zero PreprocessGasPerByte is otherwise
// unrepresentable — Validate rejects it on every write path — so a zero here
// can only mean pre-field legacy state. Defaulting it keeps the type-check/
// preprocess charge active on legacy chains, keeps WillSetParam's whole-struct
// re-validation from rejecting an unrelated param update on such state, and
// lets a relaunch genesis that predates the field still pass ValidateGenesis.
// ApplyLegacyDefaults is the exported form, for tools that must judge a genesis
// the way the node will boot it rather than as written.
func (p Params) ApplyLegacyDefaults() Params { return p.applyLegacyDefaults() }

func (p Params) applyLegacyDefaults() Params {
	if p.PreprocessGasPerByte == 0 {
		p.PreprocessGasPerByte = preprocessGasPerByteDefault
	}
	// A params blob written before code_submission_policy existed has no value
	// for it, and so does a genesis that simply omits it. Defaulting here, on
	// the one read path every caller goes through, means no consumer carries its
	// own empty-string case — which would be the same rule restated in as many
	// files as there are readers.
	if p.CodeSubmissionPolicy == "" {
		p.CodeSubmissionPolicy = codeSubmissionPolicyDefault
	}
	// The COLLECTOR is defaulted here; the CHARGE deliberately is not.
	//
	// Validate rejects a zero collector unconditionally, so a params blob
	// written before these fields existed would otherwise fail validation on
	// every read — and WillSetParam re-validates the whole struct and panics,
	// which would make every unrelated governance param update abort on a
	// legacy chain.
	//
	// Filling the charge would be the opposite of harmless: it would levy a
	// charge on replayed history that never paid one, so balances would drift
	// from the source chain. Empty means off, and that is what makes replay
	// correct without a carve-out. Do not add a case for it here, and do not
	// add one to the legacy fill in `gnogenesis fork generate` either.
	if p.InertChargeCollector.IsZero() {
		p.InertChargeCollector = crypto.AddressFromPreimage([]byte(inertChargeCollectorNameDefault))
	}
	// These three used to be reachable two ways: through the struct, and
	// through a small accessor that seeded its own default before reading the
	// same key. The two disagreed whenever the key was absent, which is how
	// AddPackage and EnablePackage came to apply the same domain rule
	// differently. Defaulting them here leaves one read path and no second
	// answer to keep in step.
	if p.ChainDomain == "" {
		p.ChainDomain = chainDomainDefault
	}
	if p.SysNamesPkgPath == "" {
		p.SysNamesPkgPath = sysNamesPkgDefault
	}
	if p.SysCLAPkgPath == "" {
		p.SysCLAPkgPath = sysCLAPkgDefault
	}
	return p
}

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := Params{}
	vm.prmk.GetStruct(ctx, "vm:p", &params) // prmk is root.
	return params.applyLegacyDefaults()
}

func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	params := vm.GetParams(ctx)
	switch key {
	case "p:sysnames_pkgpath":
		params.SysNamesPkgPath = sdkparams.MustParamString("sysnames_pkgpath", value)
	case "p:syscla_pkgpath":
		params.SysCLAPkgPath = sdkparams.MustParamString("syscla_pkgpath", value)
	case "p:chain_domain":
		params.ChainDomain = sdkparams.MustParamString("chain_domain", value)
	case "p:default_deposit":
		params.DefaultDeposit = sdkparams.MustParamString("default_deposit", value)
	case "p:storage_price":
		params.StoragePrice = sdkparams.MustParamString("storage_price", value)
	case "p:storage_fee_collector":
		s := sdkparams.MustParamString("storage_fee_collector", value)
		addr, err := crypto.AddressFromString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid storage_fee_collector address: %v", err))
		}
		params.StorageFeeCollector = addr
	case "p:min_get_read_depth_100":
		params.MinGetReadDepth100 = sdkparams.MustParamInt64("min_get_read_depth_100", value)
	case "p:min_set_read_depth_100":
		params.MinSetReadDepth100 = sdkparams.MustParamInt64("min_set_read_depth_100", value)
	case "p:min_write_depth_100":
		params.MinWriteDepth100 = sdkparams.MustParamInt64("min_write_depth_100", value)
	case "p:fixed_get_read_depth_100":
		params.FixedGetReadDepth100 = sdkparams.MustParamInt64("fixed_get_read_depth_100", value)
	case "p:fixed_set_read_depth_100":
		params.FixedSetReadDepth100 = sdkparams.MustParamInt64("fixed_set_read_depth_100", value)
	case "p:fixed_write_depth_100":
		params.FixedWriteDepth100 = sdkparams.MustParamInt64("fixed_write_depth_100", value)
	case "p:iter_next_cost_flat":
		params.IterNextCostFlat = sdkparams.MustParamInt64("iter_next_cost_flat", value)
	case "p:code_submission_policy":
		params.CodeSubmissionPolicy = CodeSubmissionPolicy(sdkparams.MustParamString("code_submission_policy", value))
	case "p:code_submitters":
		params.CodeSubmitters = mustParseAddressStrings("code_submitters", value)
	case "p:pkg_approvers":
		params.PkgApprovers = mustParseAddressStrings("pkg_approvers", value)
	case "p:run_submitters":
		params.RunSubmitters = mustParseAddressStrings("run_submitters", value)
	case "p:preprocess_gas_per_byte":
		params.PreprocessGasPerByte = sdkparams.MustParamInt64("preprocess_gas_per_byte", value)
	case "p:inert_submission_charge":
		params.InertSubmissionCharge = sdkparams.MustParamString("inert_submission_charge", value)
	case "p:inert_charge_collector":
		s := sdkparams.MustParamString("inert_charge_collector", value)
		addr, err := crypto.AddressFromString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid inert_charge_collector address: %v", err))
		}
		params.InertChargeCollector = addr
	default:
		if strings.HasPrefix(key, "p:") {
			panic(fmt.Sprintf("unknown vm param key: %q", key))
		}
		// Allow realm-scoped params through without validation.
		return
	}
	if err := params.Validate(); err != nil {
		panic("invalid param: " + err.Error())
	}
}
