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
	vm.prmk.SetStruct(ctx, "vm:p", params) // prmk is root.
	return nil
}

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := Params{}
	vm.prmk.GetStruct(ctx, "vm:p", &params) // prmk is root.
	return params
}

const (
	sysUsersPkgParamPath = "vm:p:sysusers_pkgpath"
	chainDomainParamPath = "vm:p:chain_domain"
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

func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value interface{}) {
	// XXX validate input?
}
