package vm

import (
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

const (
	sysUsersPkgDefault = "gno.land/r/sys/users"
	paramsKey          = "p"
)

// Params defines the parameters for the bank module.
type Params struct {
	SysUsersPkg string `json:"sysusers_pkgpath" yaml:"sysusers_pkgpath"`
}

// NewParams creates a new Params object
func NewParams(pkgPath string) Params {
	return Params{
		SysUsersPkg: pkgPath,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysUsersPkgDefault)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkg: %q\n", p.SysUsersPkg))
	return sb.String()
}

func (p Params) Validate() error {
	if !gno.ReRealmPath.MatchString(p.SysUsersPkg) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", p.SysUsersPkg, gno.ReRealmPath)
	}
	return nil
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
}

func (vm *VMKeeper) SetParams(ctx sdk.Context, params Params) error {
	if params.Equals(Params{}) {
		return nil
	}
	if err := params.Validate(); err != nil {
		return err
	}
	err := vm.prmk.SetParams(ctx, ModuleName, paramsKey, params)
	return err
}

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}

	_, err := vm.prmk.GetParams(ctx, ModuleName, paramsKey, params)
	if err != nil {
		panic(err.Error())
	}

	return *params
}

const (
	sysUsersPkgParamPath = "gno.land/r/sys/params.sys.users_pkgpath.string"
	chainDomainParamPath = "gno.land/r/sys/params.vm.chain_domain.string"
)

func (vm *VMKeeper) getChainDomainParam(ctx sdk.Context) string {
	chainDomain := "gno.land" // default
	vm.prmk.GetString(ctx, chainDomainParamPath, &chainDomain)
	return chainDomain
}

func (vm *VMKeeper) getSysUsersPkgParam(ctx sdk.Context) string {
	var sysUsersPkg string
	vm.prmk.GetString(ctx, sysUsersPkgParamPath, &sysUsersPkg)
	return sysUsersPkg
}

func (vm *VMKeeper) GetParamfulKey() string {
	return ModuleName
}

// WillSetParam checks if the key contains the module's parameter key prefix and updates the module parameter accordingly.
func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value interface{}) {
	// TODO: move parameters here.
}
