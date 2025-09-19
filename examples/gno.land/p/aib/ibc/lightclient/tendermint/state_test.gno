package tendermint_test

import (
	"strings"
	"testing"
	"time"

	"gno.land/p/aib/ibc/lightclient/tendermint"
	"gno.land/p/aib/ibc/types"
)

const (
	chainID = "gno.land"
	// Do not change the length of these variables
	fiftyCharChainID    = "12345678901234567890123456789012345678901234567890"
	fiftyOneCharChainID = "123456789012345678901234567890123456789012345678901"

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

var (
	invalidProof       = []byte("invalid proof")
	height             = types.NewHeight(0, 4)
	upgradePath        = []string{"upgrade", "upgradedIBCState"}
	invalidUpgradePath = []string{"upgrade", ""}
)

func TestClientStateValidate(t *testing.T) {
	testCases := []struct {
		name        string
		clientState *tendermint.ClientState
		expErr      string
	}{
		{
			name:        "valid client",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
		},
		{
			name:        "valid client with nil upgrade path",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), nil),
		},
		{
			name:        "invalid chainID",
			clientState: tendermint.NewClientState("  ", tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "chainID cannot be empty string",
		},
		{
			name:        "valid chainID - chainID validation did not fail for chainID of length 50!",
			clientState: tendermint.NewClientState(fiftyCharChainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
		},
		{
			// NOTE: if this test fails, the code must account for the change in chainID length across tendermint versions!
			// Do not only fix the test, fix the code!
			// https://github.com/cosmos/ibc-go/issues/177
			name:        "invalid chainID - chainID validation failed for chainID of length 51!",
			clientState: tendermint.NewClientState(fiftyOneCharChainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "chainID is too long",
		},
		{
			name:        "invalid trust level",
			clientState: tendermint.NewClientState(chainID, tendermint.NewFraction(0, 1), trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "trustLevel must be within [1/3, 1]",
		},
		{
			name:        "invalid zero trusting period",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, 0, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "trusting period must be greater than zero",
		},
		{
			name:        "invalid negative trusting period",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, -1, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "trusting period must be greater than zero",
		},
		{
			name:        "invalid zero unbonding period",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, 0, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "unbonding period must be greater than zero",
		},
		{
			name:        "invalid negative unbonding period",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, -1, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "unbonding period must be greater than zero",
		},
		{
			name:        "invalid zero max clock drift",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, 0, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "max clock drift must be greater than zero",
		},
		{
			name:        "invalid negative max clock drift",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, -1, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "max clock drift must be greater than zero",
		},
		{
			name:        "invalid revision number",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.NewHeight(1, 1), types.GetSDKProofSpecs(), upgradePath),
			expErr:      "latest height revision number must match chain id revision number (1 != 0)",
		},
		{
			name:        "invalid revision height",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, types.ZeroHeight(), types.GetSDKProofSpecs(), upgradePath),
			expErr:      "tendermint client's latest height revision height cannot be zero",
		},
		{
			name:        "trusting period not less than unbonding period",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, ubdPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), upgradePath),
			expErr:      "trusting period (504h0m0s) should be < unbonding period",
		},
		{
			name:        "proof specs is nil",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, nil, upgradePath),
			expErr:      "proof specs cannot be nil for tm client",
		},
		{
			name:        "proof specs contains nil",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, []*types.ProofSpec{types.TendermintSpec, nil}, upgradePath),
			expErr:      "proof spec cannot be nil at index: 1",
		},
		{
			name:        "invalid upgrade path",
			clientState: tendermint.NewClientState(chainID, tendermint.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, types.GetSDKProofSpecs(), invalidUpgradePath),
			expErr:      "key in upgrade path at index 1 cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.clientState.Validate()

			if tc.expErr == "" && err != nil {
				t.Errorf("expected no error, got %s", err)
				return
			}
			if tc.expErr != "" && !strings.Contains(err.Error(), tc.expErr) {
				t.Errorf("expected error %s, got %s", tc.expErr, err)
			}
		})
	}
}
