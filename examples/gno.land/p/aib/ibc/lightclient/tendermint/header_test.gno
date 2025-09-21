package tendermint_test

import (
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/lightclient/tendermint"
	ibctesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"
)

func TestMsgHeaderValidate(t *testing.T) {
	var msg *tendermint.MsgHeader
	testCases := []struct {
		name          string
		malleate      func()
		expectedError string
	}{
		{
			name:          "valid",
			malleate:      func() {},
			expectedError: "",
		},
		{
			name:          "header is nil",
			malleate:      func() { msg.Header = nil },
			expectedError: "missing header",
		},
		{
			name:          "commit is nil",
			malleate:      func() { msg.Commit = nil },
			expectedError: "missing commit",
		},
		{
			name: "trusted height is equal to header height",
			malleate: func() {
				msg.TrustedHeight = msg.GetHeight()
			},
			expectedError: "trustedHeight 0/3 must be less than header height 0/3",
		},
		{
			name:          "validator set nil",
			malleate:      func() { msg.ValidatorSet = nil },
			expectedError: "validator set is nil",
		},
		{
			name: "validator set hash mismatch",
			malleate: func() {
				msg.ValidatorSet.Validators[0].PubKey = []byte{}
			},
			expectedError: "validator set does not match hash",
		},
		// TODO add more cases
	}

	if lightclient.Tendermint != msg.ClientType() {
		t.Fatal("client type is not tendermint")
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				newHeight     = uint64(3)
				newTimestamp  = time.Now()
				valset        = ibctesting.GenValset()
				trustedValset = valset
				trustedHeight = types.NewHeight(0, 1)
				apphash       = sha256.Sum256([]byte("apphash"))
			)
			msg = ibctesting.GenMsgHeader(
				chainID, newTimestamp, apphash[:],
				newHeight, trustedHeight, valset, trustedValset,
			)
			tc.malleate()

			err := msg.Validate()

			if tc.expectedError == "" && err != nil {
				t.Errorf("expected no error got %v", err)
				return
			}
			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error %s, got %s", tc.expectedError, err)
				}
			}
		})
	}
}
