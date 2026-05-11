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

// MockMsgCall mimics vm.MsgCall for testing session AllowPaths.
type MockMsgCall struct {
	Caller  crypto.Address
	PkgPath string
	Send    std.Coins
}

var _ std.Msg = MockMsgCall{}

func (msg MockMsgCall) Route() string        { return "vm" }
func (msg MockMsgCall) Type() string         { return "exec" }
func (msg MockMsgCall) ValidateBasic() error { return nil }
func (msg MockMsgCall) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}
func (msg MockMsgCall) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}
func (msg MockMsgCall) GetPkgPath() string { return msg.PkgPath }
func (msg MockMsgCall) GetReceived() std.Coins {
	return msg.Send
}

// SpendForSigner implements std.SpendEstimator for the session ante pre-check.
func (msg MockMsgCall) SpendForSigner(signer crypto.Address) std.Coins {
	if signer != msg.Caller {
		return nil
	}
	return msg.Send
}

// MockMsgRun mimics vm.MsgRun for testing session allowlist behavior.
// Intentionally does NOT implement GetPkgPath — MsgRun has no realm path
// at the protocol level (the path is auto-derived from Caller), so the
// session AllowPaths check should treat it as path-less.
type MockMsgRun struct {
	Caller crypto.Address
	Send   std.Coins
}

var _ std.Msg = MockMsgRun{}

func (msg MockMsgRun) Route() string        { return "vm" }
func (msg MockMsgRun) Type() string         { return "run" }
func (msg MockMsgRun) ValidateBasic() error { return nil }
func (msg MockMsgRun) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}
func (msg MockMsgRun) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Caller}
}

// SpendForSigner implements std.SpendEstimator for the session ante pre-check.
func (msg MockMsgRun) SpendForSigner(signer crypto.Address) std.Coins {
	if signer != msg.Caller {
		return nil
	}
	return msg.Send
}

// MockMsgSend mimics bank.MsgSend for testing session allowlist behavior.
type MockMsgSend struct {
	From   crypto.Address
	To     crypto.Address
	Amount std.Coins
}

var _ std.Msg = MockMsgSend{}

func (msg MockMsgSend) Route() string        { return "bank" }
func (msg MockMsgSend) Type() string         { return "send" }
func (msg MockMsgSend) ValidateBasic() error { return nil }
func (msg MockMsgSend) GetSignBytes() []byte {
	return std.MustSortJSON(amino.MustMarshalJSON(msg))
}
func (msg MockMsgSend) GetSigners() []crypto.Address {
	return []crypto.Address{msg.From}
}

// SpendForSigner implements std.SpendEstimator for the session ante pre-check.
func (msg MockMsgSend) SpendForSigner(signer crypto.Address) std.Coins {
	if signer != msg.From {
		return nil
	}
	return msg.Amount
}

// NewSessionTestTx creates a tx signed by a session key with SessionAddr set.
func NewSessionTestTx(
	t *testing.T,
	chainID string,
	msgs []std.Msg,
	sessionPriv crypto.PrivKey,
	sessionAddr crypto.Address,
	accNum uint64,
	seq uint64,
	fee std.Fee,
) std.Tx {
	t.Helper()

	signBytes, err := std.GetSignaturePayload(std.SignDoc{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      seq,
		Fee:           fee,
		Msgs:          msgs,
	})
	require.NoError(t, err)

	sig, err := sessionPriv.Sign(signBytes)
	require.NoError(t, err)

	sigs := []std.Signature{{
		// PubKey omitted — stored on session account at creation.
		SessionAddr: sessionAddr,
		Signature:   sig,
	}}

	return std.NewTx(msgs, fee, sigs, "")
}
