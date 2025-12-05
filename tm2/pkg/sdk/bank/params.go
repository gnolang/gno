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
	if err := params.Validate(); err != nil {
		return err
	}
	bank.prmk.SetStruct(ctx, "p", params)
	return nil
}

func (bank BankKeeper) GetParams(ctx sdk.Context) Params {
	params := Params{}
	bank.prmk.GetStruct(ctx, "p", &params)
	return params
}

func (bank BankKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	switch key {
	case "p:restricted_denoms": // XXX test
		bank.WillSetRestrictedDenoms(ctx, value.([]string))
	default:
		// Allow setting non-existent key.
	}
}

func (bank BankKeeper) WillUpdateParam(ctx sdk.Context, key string, value any, add bool) {
	// XXX
}
