package types_test

import (
	"testing"

	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/urequire"
)

func TestValidateClientType(t *testing.T) {
	testCases := []struct {
		name       string
		clientType string
		expError   string
	}{
		{"valid", "tendermint", ""},
		{"valid solomachine", "solomachine-v1", ""},
		{"too large", "tenderminttenderminttenderminttenderminttendermintt", "invalid length: 72, must be between 4-64 characters"},
		{"too short", "t", "invalid length: 3, must be between 4-64 characters"},
		{"blank id", "               ", "client type cannot be blank"},
		{"empty id", "", "client type cannot be blank"},
		{"ends with dash", "tendermint-", "not a valid clientID: tendermint--0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateClientType(tc.clientType)

			if tc.expError == "" {
				urequire.NoError(t, err, tc.name)
			} else {
				urequire.ErrorContains(t, err, tc.expError)
			}
		})
	}
}
