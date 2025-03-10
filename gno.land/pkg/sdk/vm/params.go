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
	sysNamesPkgDefault = "gno.land/r/sys/names"
	chainDomainDefault = "gno.land"
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the bank module.
type Params struct {
	SysNamesPkgPath string `json:"sysnames_pkgpath" yaml:"sysnames_pkgpath"`
	ChainDomain     string `json:"chain_domain" yaml:"chain_domain"`
}

// NewParams creates a new Params object
func NewParams(namesPkgPath, chainDomain string) Params {
	return Params{
		SysNamesPkgPath: namesPkgPath,
		ChainDomain:     chainDomain,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysNamesPkgDefault, chainDomainDefault)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysNamesPkgPath))
	sb.WriteString(fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain))
	return sb.String()
}

func (p Params) Validate() error {
	if p.SysNamesPkgPath != "" && !gno.ReRealmPath.MatchString(p.SysNamesPkgPath) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", p.SysNamesPkgPath, gno.ReRealmPath)
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

func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value interface{}) {
	// XXX validate input?
}
