package sdk

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// preSkipZeroSignature is std.Signature as Go clients encoded it before
// gnolang/gno#6399: without amino:"skip_zero", so a zero SessionAddr is
// written out as the zero address's bech32 string.
type preSkipZeroSignature struct {
	PubKey      crypto.PubKey  `json:"pub_key" yaml:"pub_key"`
	Signature   []byte         `json:"signature" yaml:"signature"`
	SessionAddr crypto.Address `json:"session_addr,omitempty" yaml:"session_addr,omitempty"`
}

type preSkipZeroTx struct {
	Msgs       []std.Msg              `json:"msg" yaml:"msg"`
	Fee        std.Fee                `json:"fee" yaml:"fee"`
	Signatures []preSkipZeroSignature `json:"signatures" yaml:"signatures"`
	Memo       string                 `json:"memo" yaml:"memo"`
}

func preSkipZeroBytes(t *testing.T, tx std.Tx) []byte {
	t.Helper()

	old := preSkipZeroTx{Msgs: tx.Msgs, Fee: tx.Fee, Memo: tx.Memo}
	for _, sig := range tx.Signatures {
		old.Signatures = append(old.Signatures, preSkipZeroSignature(sig))
	}
	bz, err := amino.Marshal(old)
	require.NoError(t, err)
	return bz
}

func TestWritesZeroSessionAddr(t *testing.T) {
	t.Parallel()

	sigBytes := make([]byte, 64)
	secp := std.Signature{PubKey: secp256k1.GenPrivKeySecp256k1([]byte("a")).PubKey(), Signature: sigBytes}
	ed := std.Signature{PubKey: ed25519.GenPrivKeyFromSecret([]byte("b")).PubKey(), Signature: sigBytes}
	fee := std.NewFee(100_000, std.NewCoin("ugnot", 1_000))
	zeroField := append([]byte{0x1a, 0x28}, crypto.AddressToBech32(crypto.Address{})...)

	oneSig := std.Tx{Fee: fee, Signatures: []std.Signature{secp}}
	twoSigs := std.Tx{Fee: fee, Signatures: []std.Signature{secp, ed}}
	session := std.Tx{Fee: fee, Signatures: []std.Signature{{
		PubKey: secp.PubKey, Signature: sigBytes, SessionAddr: crypto.AddressFromPreimage([]byte("session")),
	}}}
	// The memo carries the field's bytes once more than there are signatures.
	memoEcho := std.Tx{Fee: fee, Signatures: []std.Signature{secp}, Memo: string(zeroField)}

	for _, tc := range []struct {
		name string
		bz   []byte
		sigs []std.Signature
		want bool
	}{
		{"one signature, written out", preSkipZeroBytes(t, oneSig), oneSig.Signatures, true},
		{"two signatures, written out", preSkipZeroBytes(t, twoSigs), twoSigs.Signatures, true},
		{"one signature, canonical", amino.MustMarshal(oneSig), oneSig.Signatures, false},
		{"two signatures, canonical", amino.MustMarshal(twoSigs), twoSigs.Signatures, false},
		{"session address set", preSkipZeroBytes(t, session), session.Signatures, false},
		{"memo repeats the field", preSkipZeroBytes(t, memoEcho), memoEcho.Signatures, false},
		{"no signatures", preSkipZeroBytes(t, std.Tx{Fee: fee}), nil, false},
	} {
		assert.Equal(t, tc.want, writesZeroSessionAddr(tc.bz, tc.sigs), tc.name)
	}

	// What the hint stands for: the written-out bytes decode to the same tx,
	// one 42-byte field per signature, plus a byte wherever a signature's
	// length prefix outgrows one byte.
	for _, tx := range []std.Tx{oneSig, twoSigs} {
		old := preSkipZeroBytes(t, tx)
		var decoded std.Tx
		require.NoError(t, amino.Unmarshal(old, &decoded))
		require.Equal(t, amino.MustMarshal(tx), amino.MustMarshal(decoded))
		grew := len(old) - len(amino.MustMarshal(tx))
		assert.GreaterOrEqual(t, grew, 42*len(tx.Signatures))
		assert.LessOrEqual(t, grew, 43*len(tx.Signatures))
	}
}
