package vm

import (
	"fmt"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	sysNamesPkgDefault  = "gno.land/r/sys/names"
	chainDomainDefault  = "gno.land"
	depositDefault      = "600000000ugnot"
	storagePriceDefault = "100ugnot" // cost per byte (1 gnot per 10KB) 1B GNOT == 10TB
)

var ASCIIDomain = regexp.MustCompile(`^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]{2,}$`)

// Params defines the parameters for the bank module.
type Params struct {
	SysNamesPkgPath string `json:"sysnames_pkgpath" yaml:"sysnames_pkgpath"`
	ChainDomain     string `json:"chain_domain" yaml:"chain_domain"`
	DefaultDeposit  string `json:"default_deposit" yaml:"default_deposit"`
	StoragePrice    string `json:"storage_price" yaml:"storage_price"`
}

// NewParams creates a new Params object
func NewParams(namesPkgPath, chainDomain, defaultDeposit, storagePrice string) Params {
	return Params{
		SysNamesPkgPath: namesPkgPath,
		ChainDomain:     chainDomain,
		DefaultDeposit:  defaultDeposit,
		StoragePrice:    storagePrice,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(sysNamesPkgDefault, chainDomainDefault,
		depositDefault, storagePriceDefault)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("SysUsersPkgPath: %q\n", p.SysNamesPkgPath))
	sb.WriteString(fmt.Sprintf("ChainDomain: %q\n", p.ChainDomain))
	sb.WriteString(fmt.Sprintf("DefaultDeposit: %q\n", p.DefaultDeposit))
	sb.WriteString(fmt.Sprintf("StoragePrice: %q\n", p.StoragePrice))
	return sb.String()
}

func (p Params) Validate() error {
	if p.SysNamesPkgPath != "" && !gno.IsUserlib(p.SysNamesPkgPath) {
		return fmt.Errorf("invalid user package path %q", p.SysNamesPkgPath)
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

func (vm *VMKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	// XXX validate input?
}
