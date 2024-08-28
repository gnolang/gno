package std

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
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
	addr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	msgs := []Msg{
		mockMsg{
			caller: addr,
		},
	}

	fee := NewFee(maxGasWanted, Coin{Denom: "atom", Amount: 10})
	sigs := []Signature{
		{
			Signature: []byte{0x00},
		},
	}

	tx := NewTx(msgs, fee, sigs, "test memo")

	// Valid case
	require.NoError(t, tx.ValidateBasic())

	// Invalid gas case
	invalidFee := NewFee(maxGasWanted+1, Coin{Denom: "atom", Amount: 10})
	txInvalidGas := NewTx(msgs, invalidFee, sigs, "test memo")
	require.Error(t, txInvalidGas.ValidateBasic(), "expected gas overflow error")

	// Invalid fee case
	invalidFeeAmount := NewFee(1000, Coin{Denom: "atom", Amount: -10})
	txInvalidFee := NewTx(msgs, invalidFeeAmount, sigs, "test memo")
	require.Error(t, txInvalidFee.ValidateBasic(), "expected insufficient fee error")

	// No signatures case
	txNoSigs := NewTx(msgs, fee, []Signature{}, "test memo")
	require.Error(t, txNoSigs.ValidateBasic(), "expected no signatures error")

	// Wrong number of signers case
	wrongSigs := []Signature{
		{
			Signature: []byte{0x00},
		},
		{
			Signature: []byte{0x01},
		},
	}
	txWrongSigs := NewTx(msgs, fee, wrongSigs, "test memo")
	require.Error(t, txWrongSigs.ValidateBasic(), "expected wrong number of signers error")
}

func Test_CountSubKeys(t *testing.T) {
	// Single key case
	pubKey := ed25519.GenPrivKey().PubKey()
	require.Equal(t, 1, CountSubKeys(pubKey))

	// Multi-sig case
	// Assuming multisig.PubKeyMultisigThreshold is correctly implemented for testing purposes
	pubKeys := []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}
	multisigPubKey := multisig.NewPubKeyMultisigThreshold(2, pubKeys)
	require.Equal(t, len(pubKeys), CountSubKeys(multisigPubKey))
}

func Test_GetSigners(t *testing.T) {
	// Single signer case
	addr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	msgs := []Msg{
		mockMsg{
			caller:  addr,
			msgType: "call",
		},
	}
	tx := NewTx(msgs, Fee{}, []Signature{}, "")
	require.Equal(t, []crypto.Address{addr}, tx.GetSigners())

	// Duplicate signers case
	msgs = []Msg{
		mockMsg{
			caller:  addr,
			msgType: "send",
		},
		mockMsg{
			caller:  addr,
			msgType: "send",
		},
	}

	tx = NewTx(msgs, Fee{}, []Signature{}, "")
	require.Equal(t, []crypto.Address{addr}, tx.GetSigners())

	// Multiple unique signers case
	addr2, _ := crypto.AddressFromBech32("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	msgs = []Msg{
		mockMsg{
			caller:  addr,
			msgType: "call",
		},
		mockMsg{
			caller:  addr2,
			msgType: "run",
		},
	}
	tx = NewTx(msgs, Fee{}, []Signature{}, "")
	require.Equal(t, []crypto.Address{addr, addr2}, tx.GetSigners())
}

func Test_GetSignBytes(t *testing.T) {
	msgs := []Msg{}
	fee := NewFee(1000, Coin{Denom: "atom", Amount: 10})
	sigs := []Signature{}
	tx := NewTx(msgs, fee, sigs, "test memo")
	chainID := "test-chain"
	accountNumber := uint64(1)
	sequence := uint64(1)

	signBytes, err := tx.GetSignBytes(chainID, accountNumber, sequence)
	require.NoError(t, err)
	require.NotEmpty(t, signBytes)
}
