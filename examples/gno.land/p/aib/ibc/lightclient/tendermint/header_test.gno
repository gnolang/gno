package tendermint_test

import (
	"bytes"
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
		{
			name: "header lastCommitHash invalid",
			malleate: func() {
				msg.Header.LastCommitHash = invalidHash
			},
			expectedError: "wrong LastCommitHash",
		},
		{
			name: "header dataHash invalid",
			malleate: func() {
				msg.Header.DataHash = invalidHash
			},
			expectedError: "wrong DataHash",
		},
		{
			name: "header evidenceHash invalid",
			malleate: func() {
				msg.Header.EvidenceHash = invalidHash
			},
			expectedError: "wrong EvidenceHash",
		},
		{
			name: "header proposerAddress invalid",
			malleate: func() {
				msg.Header.ProposerAddress = []byte("adr_too_short")
			},
			expectedError: "invalid ProposerAddress",
		},
		{
			name: "header validatorHash invalid",
			malleate: func() {
				msg.Header.ValidatorsHash = invalidHash
			},
			expectedError: "wrong ValidatorsHash",
		},
		{
			name: "header nextValidatorHash invalid",
			malleate: func() {
				msg.Header.NextValidatorsHash = invalidHash
			},
			expectedError: "wrong NextValidatorsHash",
		},
		{
			name: "header consensusHash invalid",
			malleate: func() {
				msg.Header.ConsensusHash = invalidHash
			},
			expectedError: "wrong ConsensusHash",
		},
		{
			name: "header lastResultHash invalid",
			malleate: func() {
				msg.Header.LastResultsHash = invalidHash
			},
			expectedError: "wrong LastResultsHash",
		},
		{
			name:          "commit is nil",
			malleate:      func() { msg.Commit = nil },
			expectedError: "missing commit",
		},
		{
			name:          "commit round negative",
			malleate:      func() { msg.Commit.Round = -1 },
			expectedError: "negative Round",
		},
		{
			name: "commit blockId is zero",
			malleate: func() {
				msg.Commit.BlockID.Hash = nil
				msg.Commit.BlockID.PartSetHeader.Total = 0
				msg.Commit.BlockID.PartSetHeader.Hash = nil
			},
			expectedError: "commit cannot be for nil block",
		},
		{
			name: "commit signatures are empty",
			malleate: func() {
				msg.Commit.Signatures = nil
			},
			expectedError: "no signatures in commit",
		},
		{
			name: "commit signature blockIdFlag is invalid",
			malleate: func() {
				msg.Commit.Signatures[0].BlockIdFlag = tendermint.BlockIDFlagUnknown
			},
			expectedError: "unknown BlockIDFlag",
		},
		{
			name: "commit signature blockIdAbsent validatorAddress is not empty",
			malleate: func() {
				msg.Commit.Signatures[0].BlockIdFlag = tendermint.BlockIDFlagAbsent
			},
			expectedError: "validator address is present",
		},
		{
			name: "commit signature blockIdAbsent timestamp is not empty",
			malleate: func() {
				msg.Commit.Signatures[0].BlockIdFlag = tendermint.BlockIDFlagAbsent
				msg.Commit.Signatures[0].ValidatorAddress = nil
			},
			expectedError: "time is present",
		},
		{
			name: "commit signature blockIdAbsent signature is not empty",
			malleate: func() {
				msg.Commit.Signatures[0].BlockIdFlag = tendermint.BlockIDFlagAbsent
				msg.Commit.Signatures[0].ValidatorAddress = nil
				msg.Commit.Signatures[0].Timestamp = time.Time{}
			},
			expectedError: "signature is present",
		},
		{
			name: "commit signature validatorAddress is invalid",
			malleate: func() {
				msg.Commit.Signatures[0].ValidatorAddress = []byte("addr_too_short")
			},
			expectedError: "expected ValidatorAddress size to be",
		},
		{
			name: "commit signature is missing",
			malleate: func() {
				msg.Commit.Signatures[0].Signature = nil
			},
			expectedError: "signature is missing",
		},
		{
			name: "commit signature is invalid",
			malleate: func() {
				msg.Commit.Signatures[0].Signature = bytes.Repeat([]byte{1}, tendermint.MaxSignatureSize+1)
			},
			expectedError: "signature is too big",
		},
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
			name:          "trusted validator set nil",
			malleate:      func() { msg.TrustedValidators = nil },
			expectedError: "trusted validator set is nil",
		},
		{
			name: "validator set hash mismatch",
			malleate: func() {
				msg.ValidatorSet.Validators[0].PubKey = []byte{}
			},
			expectedError: "validator set does not match hash",
		},
		{
			name: "trusted validator set hash mismatch",
			malleate: func() {
				msg.TrustedValidators.Validators[0].PubKey = []byte{}
			},
			expectedError: "trusted validator set does not match hash",
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
				trustedValset = ibctesting.GenValset()
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
