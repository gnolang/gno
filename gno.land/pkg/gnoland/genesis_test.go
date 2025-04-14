package gnoland

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/stretchr/testify/assert"
)

func TestGenesis_Verify(t *testing.T) {
	tests := []struct {
		name      string
		genesis   GnoGenesisState
		expectErr bool
	}{
		{"default GenesisState", DefaultGenState(), false},
		{
			"invalid GenesisState Auth",
			GnoGenesisState{
				Auth: auth.GenesisState{},
				Bank: bank.DefaultGenesisState(),
				VM:   vm.DefaultGenesisState(),
			},
			true,
		},
		{
			"invalid GenesisState Bank",
			GnoGenesisState{
				Auth: auth.DefaultGenesisState(),
				Bank: bank.GenesisState{
					Params: bank.Params{
						RestrictedDenoms: []string{"INVALID!!!"},
					},
				},
				VM: vm.DefaultGenesisState(),
			},
			true,
		},
		{
			"invalid GenesisState VM",
			GnoGenesisState{
				Auth: auth.DefaultGenesisState(),
				Bank: bank.DefaultGenesisState(),
				VM:   vm.GenesisState{},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGenState(tc.genesis)
			if tc.expectErr {
				assert.Error(t, err, fmt.Sprintf("TestGenesis_Verify: %s", tc.name))
			} else {
				assert.NoError(t, err, fmt.Sprintf("TestGenesis_Verify: %s", tc.name))
			}
		})
	}
}
