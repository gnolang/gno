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
			name:          "header version block invalid",
			malleate:      func() { msg.Header.Version.Block-- },
			expectedError: "block protocol is incorrect",
		},
		{
			name: "header chainId too long",
			malleate: func() {
				msg.Header.ChainID = fiftyOneCharChainID
			},
			expectedError: "chainID is too long",
		},
		{
			name:          "header height is zero",
			malleate:      func() { msg.Header.Height = 0 },
			expectedError: "zero Height",
		},
		{
			name: "header lastBlockID hash invalid",
			malleate: func() {
				msg.Header.LastBlockID.Hash = invalidHash
			},
			expectedError: "wrong blockID Hash",
		},
		{
			name: "header lastBlockID ParSetHeader hash invalid",
			malleate: func() {
				msg.Header.LastBlockID.PartSetHeader.Hash = invalidHash
			},
			expectedError: "wrong parSetHeader Hash",
		},
		// TODO add more cases from Header.ValidateBasic
		{
			name:          "commit is nil",
			malleate:      func() { msg.Commit = nil },
			expectedError: "missing commit",
		},
		// TODO add more cases from Commit.ValidateBasic
		{
			name:          "commit and header height mismatch",
			malleate:      func() { msg.Commit.Height-- },
			expectedError: "header and commit height mismatch: 3 vs 2",
		},
		{
			name: "trusted height is equal to header height",
			malleate: func() {
				msg.TrustedHeight = msg.GetHeight()
			},
			expectedError: "trustedHeight 0/3 must be less than header height 0/3",
		},
		{
			name: "trusted height is lower to header height",
			malleate: func() {
				msg.TrustedHeight = msg.GetHeight()
				msg.TrustedHeight.RevisionHeight++
			},
			expectedError: "trustedHeight 0/4 must be less than header height 0/3",
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
