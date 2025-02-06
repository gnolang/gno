package bank

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const paramsKey = "p"

type BankParamsContextKey struct{}

// Params defines the parameters for the bank module.
type Params struct {
	RestrictedDenoms []string `json:"restricted_denoms" yaml:"restricted_denoms"`
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
	err := bank.paramk.SetParams(ctx, ModuleName, paramsKey, params)

	return err
}

func (bank BankKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}
	_, err := bank.paramk.GetParams(ctx, ModuleName, paramsKey, params)
	if err != nil {
		panic(err.Error())
	}
	return *params
}

func (bank BankKeeper) GetParamfulKey() string {
	return ModuleName
}

// WillSetParam checks if the key contains the module's parameter key and updates the module parameter accordingly.
func (bank BankKeeper) WillSetParam(ctx sdk.Context, key string, value interface{}) {
	if key == lockTransferKey {
		if value != "" { // lock sending denoms
			bank.AddRestrictedDenoms(ctx, value.(string))
		} else { // unlock sending ugnot
			bank.DelAllRestrictedDenoms(ctx)
		}
	}
}
