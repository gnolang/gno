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
)

const (
	sysNamesPkgDefault = "gno.land/r/sys/names"
	chainDomainDefault = "gno.land"
	ValsetRealmDefault = "gno.land/r/sys/validators/v3"
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the bank module.
type Params struct {
	SysNamesPkgPath string `json:"sysnames_pkgpath" yaml:"sysnames_pkgpath"`
	ChainDomain     string `json:"chain_domain" yaml:"chain_domain"`
	ValsetRealmPath string `json:"valset_realm_path" yaml:"valset_realm_path"`
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		SysNamesPkgPath: sysNamesPkgDefault,
		ChainDomain:     chainDomainDefault,
		ValsetRealmPath: ValsetRealmDefault,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysNamesPkgPath))
	sb.WriteString(fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain))
	sb.WriteString(fmt.Sprintf("ValsetRealmPath: %q\n", p.ValsetRealmPath))
	return sb.String()
}

func (p Params) Validate() error {
	if p.SysNamesPkgPath != "" && !gno.ReRealmPath.MatchString(p.SysNamesPkgPath) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", p.SysNamesPkgPath, gno.ReRealmPath)
	}
	if p.ChainDomain != "" && !ASCIIDomain.MatchString(p.ChainDomain) {
		return fmt.Errorf("invalid chain domain %q, failed to match %q", p.ChainDomain, ASCIIDomain)
	}
	if p.ValsetRealmPath != "" && !gno.ReRealmPath.MatchString(p.ValsetRealmPath) {
		return fmt.Errorf("invalid valset realm path %q, failed to match %q", p.ValsetRealmPath, gno.ReRealmPath)
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
	chainDomainParamPath = moduleParamPrefix + ":p:chain_domain"
	ValsetRealmParamPath = moduleParamPrefix + ":p:valset_realmpath"
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

func (vm *VMKeeper) getValsetRealmParam(ctx sdk.Context) string {
	// Fetch the latest valset realm path
	valsetRealm := ValsetRealmDefault
	vm.prmk.GetString(ctx, ValsetRealmParamPath, &valsetRealm)

	return valsetRealm
}

func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	switch {
	// vm:<valset-realm-path>:valset_new
	case strings.HasPrefix(key, vm.getValsetRealmParam(ctx)+":"+"valset_new"):
		// Validate the proposed valset changes
		changes, ok := value.([]string)
		if !ok {
			// This panic is explicit, because the issue is otherwise
			// unclearly propagated in the GnoVM to the caller (cast error)
			panic(
				fmt.Sprintf(
					"Value for VM param %s update is an invalid type (%T)",
					key,
					value,
				),
			)
		}

		// Sanity check the param update
		if err := validateValsetUpdate(changes); err != nil {
			panic(err)
		}
	default:
		// Allow setting arbitrary key-vals
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
				"valset update is not in the format <address>:<pub-key>:<voting-power>, but %s",
				change,
			)
		}

		// Grab the address
		address, err := crypto.AddressFromBech32(changeParts[0])
		if err != nil {
			return fmt.Errorf("invalid validator address: %w", err)
		}

		// Grab the public key
		pubKey, err := crypto.PubKeyFromBech32(changeParts[1])
		if err != nil {
			return fmt.Errorf("invalid validator pubkey: %w", err)
		}

		// Make sure the address matches the public key
		if pubKey.Address().Compare(address) != 0 {
			return fmt.Errorf(
				"address (%s) does not match public key address (%s)",
				address,
				pubKey.Address(),
			)
		}

		// Validate the voting power
		_, err = strconv.ParseUint(changeParts[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid voting power: %w", err)
		}
	}

	return nil
}
