package vm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	sdkparams "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	sysNamesPkgDefault             = "gno.land/r/sys/names"
	sysCLAPkgDefault               = "gno.land/r/sys/cla"
	chainDomainDefault             = "gno.land"
	depositDefault                 = "600000000ugnot"
	storagePriceDefault            = "100ugnot" // cost per byte (1 gnot per 10KB) 1B GNOT == 10TB
	storageFeeCollectorNameDefault = "storage_fee_collector"

	// ValsetRealmDefault is the default realm path for on-chain validator set management.
	// Keep in sync with examples/gno.land/r/sys/validators/v3/poc.gno
	ValsetRealmDefault = "gno.land/r/sys/validators/v3"
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the vm module.
type Params struct {
	SysNamesPkgPath     string         `json:"sysnames_pkgpath" yaml:"sysnames_pkgpath"`
	SysCLAPkgPath       string         `json:"syscla_pkgpath" yaml:"syscla_pkgpath"`
	ChainDomain         string         `json:"chain_domain" yaml:"chain_domain"`
	DefaultDeposit      string         `json:"default_deposit" yaml:"default_deposit"`
	StoragePrice        string         `json:"storage_price" yaml:"storage_price"`
	StorageFeeCollector crypto.Address `json:"storage_fee_collector" yaml:"storage_fee_collector"`
	ValsetRealmPath     string         `json:"valset_realm_path" yaml:"valset_realm_path"`
}

// NewParams creates a new Params object
func NewParams(namesPkgPath, claPkgPath, chainDomain, defaultDeposit, storagePrice string, storageFeeCollector crypto.Address) Params {
	return Params{
		SysNamesPkgPath:     namesPkgPath,
		SysCLAPkgPath:       claPkgPath,
		ChainDomain:         chainDomain,
		DefaultDeposit:      defaultDeposit,
		StoragePrice:        storagePrice,
		StorageFeeCollector: storageFeeCollector,
		ValsetRealmPath:     ValsetRealmDefault,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysNamesPkgDefault, sysCLAPkgDefault, chainDomainDefault,
		depositDefault, storagePriceDefault, crypto.AddressFromPreimage([]byte(storageFeeCollectorNameDefault)))
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
	sb.WriteString(fmt.Sprintf("ValsetRealmPath: %q\n", p.ValsetRealmPath))
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
	if p.ValsetRealmPath != "" && !gno.IsRealmPath(p.ValsetRealmPath) {
		return fmt.Errorf("invalid valset realm path %q", p.ValsetRealmPath)
	}
	return nil
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
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
	moduleParamPrefix = "vm"

	sysUsersPkgParamPath = moduleParamPrefix + ":p:sysnames_pkgpath"
	sysCLAPkgParamPath   = moduleParamPrefix + ":p:syscla_pkgpath"
	chainDomainParamPath = moduleParamPrefix + ":p:chain_domain"

	// ValsetRealmParamPath is the param key that stores the path of the
	// realm responsible for on-chain validator set management.
	ValsetRealmParamPath = moduleParamPrefix + ":p:valset_realm_path"
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

func (vm *VMKeeper) getValsetRealmParam(ctx sdk.Context) string {
	valsetRealm := ValsetRealmDefault
	vm.prmk.GetString(ctx, ValsetRealmParamPath, &valsetRealm)
	return valsetRealm
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
	case "p:valset_realm_path":
		params.ValsetRealmPath = sdkparams.MustParamString("valset_realm_path", value)
	default:
		if strings.HasPrefix(key, "p:") {
			panic(fmt.Sprintf("unknown vm param key: %q", key))
		}
		// Validate valset updates if the key targets the valset realm's valset_new param.
		valsetRealm := vm.getValsetRealmParam(ctx)
		if strings.HasPrefix(key, valsetRealm+":valset_new") {
			changes, ok := value.([]string)
			if !ok {
				panic(fmt.Sprintf(
					"value for VM param %s update is an invalid type (%T)",
					key,
					value,
				))
			}
			if err := validateValsetUpdate(changes); err != nil {
				panic(err)
			}
		}
		// Allow realm-scoped params through without validation.
		return
	}
	if err := params.Validate(); err != nil {
		panic("invalid param: " + err.Error())
	}
}

// validateValsetUpdate validates the validator set updates,
// which are serialized in the form:
//   - <address>:<pub-key>:<voting-power>
//   - voting power == 0 => validator removal
//   - voting power != 0 => validator power update / validator addition
func validateValsetUpdate(changes []string) error {
	for _, change := range changes {
		changeParts := strings.Split(change, ":")
		if len(changeParts) != 3 {
			return fmt.Errorf(
				"valset update is not in the format <address>:<pub-key>:<voting-power>, but %q",
				change,
			)
		}

		address, err := crypto.AddressFromBech32(changeParts[0])
		if err != nil {
			return fmt.Errorf("invalid validator address: %w", err)
		}

		pubKey, err := crypto.PubKeyFromBech32(changeParts[1])
		if err != nil {
			return fmt.Errorf("invalid validator pubkey: %w", err)
		}

		if pubKey.Address().Compare(address) != 0 {
			return fmt.Errorf(
				"address (%s) does not match public key address (%s)",
				address,
				pubKey.Address(),
			)
		}

		if _, err = strconv.ParseUint(changeParts[2], 10, 64); err != nil {
			return fmt.Errorf("invalid voting power: %w", err)
		}
	}

	return nil
}
