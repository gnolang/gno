package vm

import (
	"fmt"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

const (
	sysUsersPkgDefault = "gno.land/r/sys/users"
	chainDomainDefault = "gno.land"
	paramsKey          = "p"
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the bank module.
type Params struct {
	SysUsersPkgPath string `json:"sysusers_pkgpath" yaml:"sysusers_pkgpath"`
	ChainDomain     string `json:"chain_domain" yaml:"chain_domain"`
}

// NewParams creates a new Params object
func NewParams(userPkgPath, chainDomain string) Params {
	return Params{
		SysUsersPkgPath: userPkgPath,
		ChainDomain:     chainDomain,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysUsersPkgDefault, chainDomainDefault)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysUsersPkgPath))
	sb.WriteString(fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain))
	return sb.String()
}

func (p Params) Validate() error {
	if p.SysUsersPkgPath != "" && !gno.ReRealmPath.MatchString(p.SysUsersPkgPath) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", p.SysUsersPkgPath, gno.ReRealmPath)
	}
	if p.ChainDomain != "" && !ASCIIDomain.MatchString(p.ChainDomain) {
		return fmt.Errorf("invalid chain domain %q, failed to match %q", p.ChainDomain, ASCIIDomain)
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
	err := vm.prmk.SetParams(ctx, ModuleName, paramsKey, params)
	return err
}

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}
	// NOTE: important to not use local cached fields unless they are synchronously stored to the underlying store.
	// this optimization generally only belongs in paramk.GetParams(), not here. users of paramk.GetParams() generally
	// should not cache anything and instead rely on the efficiency of paramk.GetParams().
	_, err := vm.prmk.GetParams(ctx, ModuleName, paramsKey, params)
	if err != nil {
		panic(err)
	}

	return *params
}

const (
	sysUsersPkgParamPath = "vm:users_pkgpath.string"
	chainDomainParamPath = "vm:chain_domain.string"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) string {
	chainDomain := chainDomainDefault // default
	vm.prmk.GetString(ctx, chainDomainParamPath, &chainDomain)
	return chainDomain
}

func (vm *VMKeeper) getSysUsersPkgParam(ctx sdk.Context) string {
	sysUsersPkg := sysUsersPkgDefault
	vm.prmk.GetString(ctx, sysUsersPkgParamPath, &sysUsersPkg)
	return sysUsersPkg
}

func (vm *VMKeeper) GetParamfulKey() string {
	return ModuleName
}

// WillSetParam checks if the key contains the module's parameter key prefix and updates the
// module parameter accordingly.The key is in the format (<realm>.)?<key>. If <realm> is present,
// the key is an arbitrary key; otherwise, the key is a module key and needs to be checked against
// the module's parameter keys.
func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value interface{}) {
	params := vm.GetParams(ctx)
	realm := gno.ReRealmPath.FindString(key)
	var err error
	if realm == "" { // module parameters
		switch key {
		case "sysusers_pkgpath.string":
			params.SysUsersPkgPath = value.(string)
			err = vm.SetParams(ctx, params)
		case "chain_domain.string":
			params.ChainDomain = value.(string)
			err = vm.SetParams(ctx, params)
		default:
			panic(fmt.Sprintf("unknown parameter key: %s\n", key))
		}
	}
	if err != nil {
		panic(err)
	}
	// TODO: Check for duplicate parameter key names between individual fields and the fields
	// of the Params struct.
	err = vm.prmk.SetParams(ctx, ModuleName, key, value)
	if err != nil {
		panic(err)
	}
}
