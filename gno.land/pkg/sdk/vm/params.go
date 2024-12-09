package vm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

const (
	sysUsersPkgDefault = "gno.land/r/sys/users"

	paramsKey = "p"
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
	// XXX: This line is copied from gnovm/memfile.go. Should we instead export rePkgOrRlmPath with a getter method from gnovm?
	rePkgOrRlmPath := regexp.MustCompile(`^gno\.land\/(?:p|r)(?:\/_?[a-z]+[a-z0-9_]*)+$`)
	if !rePkgOrRlmPath.MatchString(p.SysUsersPkg) {
		return fmt.Errorf("invalid package/realm path %q, failed to match %q", p.SysUsersPkg, rePkgOrRlmPath)
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
	vm.params = params
	err := vm.prmk.SetParams(ctx, ModuleName, paramsKey, params)
	return err
}

func (vm *VMKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}

	ok, err := vm.prmk.GetParams(ctx, ModuleName, paramsKey, params)

	if !ok {
		panic("params key " + ModuleName + " does not exist")
	}
	if err != nil {
		panic(err.Error())
	}
	return *params
}
