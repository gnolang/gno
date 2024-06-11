package testutils

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// msg type for testing
type TestMsg struct {
	Signers []crypto.Address
}

var _ std.Msg = &TestMsg{}

func NewTestMsg(addrs ...crypto.Address) *TestMsg {
	return &TestMsg{
		Signers: addrs,
	}
}

func (msg *TestMsg) Route() string { return "TestMsg" }
func (msg *TestMsg) Type() string  { return "Test message" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := amino.MarshalJSON(msg.Signers)
	if err != nil {
		panic(err)
	}
	return std.MustSortJSON(bz)
}
func (msg *TestMsg) ValidateBasic() error { return nil }
func (msg *TestMsg) GetSigners() []crypto.Address {
	return msg.Signers
}

// ----------------------------------------
// Utility Methods

func NewTestFee() std.Fee {
	return std.NewFee(50000, std.NewCoin("atom", 150))
}

// coins to more than cover the fee
func NewTestCoins() std.Coins {
	return std.Coins{std.NewCoin("atom", 10000000)}
}

func KeyTestPubAddr() (crypto.PrivKey, crypto.PubKey, crypto.Address) {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := pub.Address()
	return key, pub, addr
}

func NewTestTx(
	t *testing.T,
	chainID string,
	msgs []std.Msg,
	privs []crypto.PrivKey,
	accNums []uint64,
	seqs []uint64,
	fee std.Fee,
) std.Tx {
	t.Helper()

	sigs := make([]std.Signature, len(privs))
	for i, priv := range privs {
		signBytes, err := std.GetSignaturePayload(std.SignDoc{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      seqs[i],
			Fee:           fee,
			Msgs:          msgs,
		})
		require.NoError(t, err)

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = std.Signature{PubKey: priv.PubKey(), Signature: sig}
	}

	tx := std.NewTx(msgs, fee, sigs, "")
	return tx
}

func NewTestTxWithMemo(
	t *testing.T,
	chainID string,
	msgs []std.Msg,
	privs []crypto.PrivKey,
	accNums []uint64,
	seqs []uint64,
	fee std.Fee,
	memo string,
) std.Tx {
	t.Helper()

	sigs := make([]std.Signature, len(privs))
	for i, priv := range privs {
		signBytes, err := std.GetSignaturePayload(std.SignDoc{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      seqs[i],
			Fee:           fee,
			Msgs:          msgs,
			Memo:          memo,
		})
		require.NoError(t, err)

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = std.Signature{PubKey: priv.PubKey(), Signature: sig}
	}

	tx := std.NewTx(msgs, fee, sigs, memo)
	return tx
}

func NewTestTxWithSignBytes(msgs []std.Msg, privs []crypto.PrivKey, fee std.Fee, signBytes []byte, memo string) std.Tx {
	sigs := make([]std.Signature, len(privs))
	for i, priv := range privs {
		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = std.Signature{PubKey: priv.PubKey(), Signature: sig}
	}

	tx := std.NewTx(msgs, fee, sigs, memo)
	return tx
}

func TestAddress(name string) crypto.Address {
	if len(name) > crypto.AddressSize {
		panic("address name cannot be greater than crypto.AddressSize bytes")
	}
	addr := crypto.Address{}
	// TODO: use strings.RepeatString or similar.
	// NOTE: I miss python's "".Join().
	blanks := "____________________"
	copy(addr[:], []byte(blanks))
	copy(addr[:], []byte(name))
	return addr
}

func TestBech32Address(name string) crypto.Bech32Address {
	return TestAddress(name).Bech32()
}
