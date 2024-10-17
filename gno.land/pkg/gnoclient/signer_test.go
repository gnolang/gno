package gnoclient

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignerFromBip39 tests the SignerFromBip39 function.
func TestSignerFromBip39(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-id"
	passphrase := ""
	account := uint32(0)
	index := uint32(0)

	// Define test cases with mnemonic and expected outcomes.
	testcases := []struct {
		name          string
		mnemonic      string
		expectedError bool
	}{
		{
			name:          "Valid mnemonic",
			mnemonic:      "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply",
			expectedError: false,
		},
		{
			name:          "Invalid mnemonic",
			mnemonic:      "invalid mnemonic",
			expectedError: true,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a signer from mnemonic
			signer, err := SignerFromBip39(tc.mnemonic, chainID, passphrase, account, index)

			// Check if an error was expected
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, signer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, signer)

				// Validate the signer
				err = signer.Validate()
				assert.NoError(t, err)
			}
		})
	}
}

// TestSignerFromKeybase tests the SignerFromKeybase struct.
func TestSignerFromKeybase(t *testing.T) {
	t.Parallel()

	chainID := "test-chain-id"
	passphrase := ""
	account := uint32(0)
	index := uint32(0)

	mnemonic := "index brass unknown lecture autumn provide royal shrimp elegant wink now zebra discover swarm act ill you bullet entire outdoor tilt usage gap multiply"

	// Define test cases for different scenarios of the signer
	tests := []struct {
		name          string
		account       string
		password      string
		expectedError bool
		validateOnly  bool
	}{
		{
			name:          "Valid signer",
			account:       "default",
			password:      "",
			expectedError: false,
		},
		{
			name:          "Missing ChainID",
			account:       "default",
			password:      "",
			expectedError: true,
			validateOnly:  true,
		},
		{
			name:          "Incorrect password",
			account:       "default",
			password:      "wrong-password",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel

			// Initialize in-memory keybase and create account
			kb := keys.NewInMemory()
			name := "default"
			password := ""

			_, err := kb.CreateAccount(name, mnemonic, passphrase, password, account, index)
			require.NoError(t, err)

			// Create a signer from the keybase
			signer := SignerFromKeybase{
				Keybase:  kb,
				Account:  tc.account,
				Password: tc.password,
				ChainID:  chainID,
			}

			signerInfo, err := signer.Info()
			require.NoError(t, err)

			// Test for missing ChainID scenario
			if tc.validateOnly {
				signer.ChainID = ""
				err := signer.Validate()
				assert.Error(t, err)
				assert.Equal(t, "missing ChainID", err.Error())
			} else {
				// Prepare a sign configuration
				signCfg := SignCfg{
					Tx: std.Tx{
						Msgs: []std.Msg{
							vm.MsgCall{
								Caller: signerInfo.GetAddress(),
							},
						},
						Fee: std.NewFee(0, std.NewCoin("ugnot", 1000000)),
					},
				}

				// Try to sign the transaction
				signedTx, err := signer.Sign(signCfg)

				// Check if an error was expected
				if tc.expectedError {
					assert.Error(t, err)
					assert.Nil(t, signedTx)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, signedTx)
				}
			}
		})
	}
}
