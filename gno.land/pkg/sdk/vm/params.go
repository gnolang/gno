package vm

import (
	"fmt"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	sdkparams "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

const (
	sysNamesPkgDefault             = "gno.land/r/sys/names"
	sysCLAPkgDefault               = "gno.land/r/sys/cla"
	chainDomainDefault             = "gno.land"
	depositDefault                 = "600000000ugnot"
	storagePriceDefault            = "100ugnot" // cost per byte (1 gnot per 10KB) 1B GNOT == 10TB
	storageFeeCollectorNameDefault = "storage_fee_collector"
	// Depth floors calibrated for B+32 at 100M items with 10K cache, batched 1000 muts.
	minGetReadDepth100Default = int64(300) // 3.0 GET read ops
	minSetReadDepth100Default = int64(200) // 2.0 SET read ops
	minWriteDepth100Default   = int64(440) // 4.4 write ops (batched)
	// Iterator step flat cost; mirrors store.DefaultGasConfig().IterNextCostFlat.
	iterNextCostFlatDefault = int64(1_000)
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
}

// NewParams creates a new Params object
func NewParams(namesPkgPath, claPkgPath, chainDomain, defaultDeposit, storagePrice string, storageFeeCollector crypto.Address, minGetReadDepth100, minSetReadDepth100, minWriteDepth100, iterNextCostFlat int64) Params {
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
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysNamesPkgDefault, sysCLAPkgDefault, chainDomainDefault,
		depositDefault, storagePriceDefault, crypto.AddressFromPreimage([]byte(storageFeeCollectorNameDefault)),
		minGetReadDepth100Default, minSetReadDepth100Default, minWriteDepth100Default,
		iterNextCostFlatDefault)
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
	coins, err := std.ParseCoins(p.DefaultDeposit)
	if len(coins) == 0 || err != nil {
		return fmt.Errorf("invalid default storage deposit %q", p.DefaultDeposit)
	}
	coins, err = std.ParseCoins(p.StoragePrice)
	if len(coins) == 0 || err != nil {
		return fmt.Errorf("invalid storage price %q", p.StoragePrice)
	}
	if p.StorageFeeCollector.IsZero() {
		return fmt.Errorf("invalid storage fee collector, cannot be empty")
	}
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
	const maxIterNextCostFlat = int64(100_000)
	if p.IterNextCostFlat <= 0 {
		return fmt.Errorf("IterNextCostFlat must be positive, got %d", p.IterNextCostFlat)
	}
	if p.IterNextCostFlat > maxIterNextCostFlat {
		return fmt.Errorf("IterNextCostFlat must be <= %d, got %d", maxIterNextCostFlat, p.IterNextCostFlat)
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

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := Params{}
	vm.prmk.GetStruct(ctx, "vm:p", &params) // prmk is root.
	return params
}

const (
	sysUsersPkgParamPath = "vm:p:sysnames_pkgpath"
	sysCLAPkgParamPath   = "vm:p:syscla_pkgpath"
	chainDomainParamPath = "vm:p:chain_domain"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) string {
	chainDomain := chainDomainDefault // default
	vm.prmk.GetString(ctx, chainDomainParamPath, &chainDomain)
	return chainDomain
}

func (vm *VMKeeper) getSysNamesPkgParam(ctx sdk.Context) string {
	sysNamesPkg := sysNamesPkgDefault
	vm.prmk.GetString(ctx, sysUsersPkgParamPath, &sysNamesPkg)
	return sysNamesPkg
}

func (vm *VMKeeper) getSysCLAPkgParam(ctx sdk.Context) string {
	sysCLAPkg := sysCLAPkgDefault
	vm.prmk.GetString(ctx, sysCLAPkgParamPath, &sysCLAPkg)
	return sysCLAPkg
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
