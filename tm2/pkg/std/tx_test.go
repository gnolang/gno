package std

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations of Msg interfaces
type mockMsg struct {
	caller  crypto.Address
	msgType string
}

func (m mockMsg) ValidateBasic() error {
	return nil
}

func (m mockMsg) GetSignBytes() []byte {
	return nil
}

func (m mockMsg) GetSigners() []crypto.Address {
	return []crypto.Address{m.caller}
}

func (m mockMsg) Route() string {
	return ""
}

func (m mockMsg) Type() string {
	return m.msgType
}

func TestNewTx(t *testing.T) {
	t.Parallel()

	addr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	msgs := []Msg{
		mockMsg{
			caller: addr,
		},
	}

	fee := NewFee(1000, Coin{Denom: "atom", Amount: 10})

	sigs := []Signature{
		{
			Signature: []byte{0x00},
		},
	}

	memo := "test memo"

	tx := NewTx(msgs, fee, sigs, memo)
	require.Equal(t, msgs, tx.GetMsgs())
	require.Equal(t, fee, tx.Fee)
	require.Equal(t, sigs, tx.GetSignatures())
	require.Equal(t, memo, tx.GetMemo())
}

func Test_ValidateBasic(t *testing.T) {
	t.Parallel()

	addr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	msgs := []Msg{
		mockMsg{
			caller: addr,
		},
	}

	testCases := []struct {
		name          string
		tx            Tx
		expectedError string
	}{
		{
			name:          "Valid case",
			tx:            NewTx(msgs, NewFee(maxGasWanted, Coin{Denom: "atom", Amount: 10}), []Signature{{Signature: []byte{0x00}}}, "test memo"),
			expectedError: "",
		},
		{
			name:          "Invalid gas case",
			tx:            NewTx(msgs, NewFee(maxGasWanted+1, Coin{Denom: "atom", Amount: 10}), []Signature{{Signature: []byte{0x00}}}, "test memo"),
			expectedError: "expected gas overflow error",
		},
		{
			name:          "Invalid fee case",
			tx:            NewTx(msgs, NewFee(1000, Coin{Denom: "atom", Amount: -10}), []Signature{{Signature: []byte{0x00}}}, "test memo"),
			expectedError: "expected insufficient fee error",
		},
		{
			name:          "No signatures case",
			tx:            NewTx(msgs, NewFee(maxGasWanted, Coin{Denom: "atom", Amount: 10}), []Signature{}, "test memo"),
			expectedError: "expected no signatures error",
		},
		{
			name:          "Wrong number of signers case",
			tx:            NewTx(msgs, NewFee(maxGasWanted, Coin{Denom: "atom", Amount: 10}), []Signature{{Signature: []byte{0x00}}, {Signature: []byte{0x01}}}, "test memo"),
			expectedError: "expected wrong number of signers error",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.tx.ValidateBasic()
			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, tc.expectedError)
			}
		})
	}
}

func TestCountSubKeys(t *testing.T) {
	t.Parallel()

	// Single key case
	pubKey := ed25519.GenPrivKey().PubKey()
	require.Equal(t, 1, CountSubKeys(pubKey))

	// Multi-sig case
	pubKeys := []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}
	multisigPubKey := multisig.NewPubKeyMultisigThreshold(2, pubKeys)
	require.Equal(t, len(pubKeys), CountSubKeys(multisigPubKey))
}

func Test_GetSigners(t *testing.T) {
	t.Parallel()

	addr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	addr2, _ := crypto.AddressFromBech32("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	testCases := []struct {
		name     string
		msgs     []Msg
		expected []crypto.Address
	}{
		{
			name: "Single signer case",
			msgs: []Msg{
				mockMsg{
					caller:  addr,
					msgType: "call",
				},
			},
			expected: []crypto.Address{addr},
		},
		{
			name: "Duplicate signers case",
			msgs: []Msg{
				mockMsg{
					caller:  addr,
					msgType: "send",
				},
				mockMsg{
					caller:  addr,
					msgType: "send",
				},
			},
			expected: []crypto.Address{addr},
		},
		{
			name: "Multiple unique signers case",
			msgs: []Msg{
				mockMsg{
					caller:  addr,
					msgType: "call",
				},
				mockMsg{
					caller:  addr2,
					msgType: "run",
				},
			},
			expected: []crypto.Address{
				addr,
				addr2,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx := NewTx(tc.msgs, Fee{}, []Signature{}, "")
			require.Equal(t, tc.expected, tx.GetSigners())
		})
	}
}

func Test_GetSignBytes(t *testing.T) {
	t.Parallel()

	msgs := []Msg{}
	fee := NewFee(1000, Coin{Denom: "ugnot", Amount: 10})
	sigs := []Signature{}
	tx := NewTx(msgs, fee, sigs, "test memo")
	chainID := "test-chain"
	accountNumber := uint64(1)
	sequence := uint64(1)

	// Generate the signBytes
	signBytes, err := tx.GetSignBytes(chainID, accountNumber, sequence)
	require.NoError(t, err)

	expectedResult := "{\"account_number\":\"1\",\"chain_id\":\"test-chain\",\"fee\":{\"gas_fee\":\"10ugnot\",\"gas_wanted\":\"1000\"},\"memo\":\"test memo\",\"msgs\":[],\"sequence\":\"1\"}"

	assert.Equal(t, expectedResult, string(signBytes), "signBytes did not match the expected output")
}
