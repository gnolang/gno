package bank

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type BankParamsContextKey struct{}

// Params defines the parameters for the bank module.
type Params struct {
	RestrictedDenoms []string `json:"restriced_denoms" yaml:"restriced_denoms"`
}

// NewParams creates a new Params object
func NewParams(restDenoms []string) Params {
	return Params{
		RestrictedDenoms: restDenoms,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams([]string{})
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("RestrictedDenom: %q\n", p.RestrictedDenoms))
	return sb.String()
}

func (p *Params) Validate() error {
	for _, denom := range p.RestrictedDenoms {
		err := std.ValidateDenom(denom)
		if err != nil {
			return fmt.Errorf("invalid restricted denom: %s", denom)
		}
	}
	return nil
}

func (bank BankKeeper) SetParams(ctx sdk.Context, params Params) error {
	if len(params.RestrictedDenoms) == 0 {
		return nil
	}
	if err := params.Validate(); err != nil {
		return err
	}
	bank.params = params
	bank.paramk.SetParams(ctx, ModuleName, "p", params)

	return nil
}

func (bank BankKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}

	ok, err := bank.paramk.GetParams(ctx, ModuleName, "p", params)

	if !ok {
		panic("params key " + ModuleName + " does not exist")
	}
	if err != nil {
		panic(err.Error())
	}
	return *params
}
