package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A K-of-N multisignature carrying more than K signatures still verifies with
// any K of them, since the sign bytes do not say which constituents sign. So a
// relayer could drop the spares and land the tx under a new hash. The app takes
// exactly K: a tx signed by more is refused before it costs anything, and so is
// a relayed copy of an exactly-K tx with a signature dropped, or with its
// multisignature re-encoded, while the exactly-K original runs.
func TestRelayedMultisigSparesAreRefused(t *testing.T) {
	const chainID = "test"
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	privs := getDummyKeys(t, 3)
	pubs := []crypto.PubKey{privs[0].PubKey(), privs[1].PubKey(), privs[2].PubKey()}
	msPub := multisig.NewPubKeyMultisigThreshold(2, pubs)
	addr := msPub.Address()

	gen := gnoGenesisState(t)
	gen.Balances = []Balance{{Address: addr, Amount: std.NewCoins(std.NewCoin("ugnot", 10_000_000_000))}}
	require.True(t, bapp.InitChain(abci.RequestInitChain{
		ChainID: chainID, AppState: gen,
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
	}).IsOK())
	bapp.Commit()
	// One empty block, so the ante handler checks signatures against the
	// account's real number and sequence.
	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()
	acc := queryAccount(t, bapp, addr)

	tx := std.Tx{
		Msgs: []std.Msg{bank.NewMsgSend(addr, crypto.AddressFromPreimage([]byte("to")), std.NewCoins(std.NewCoin("ugnot", 1000)))},
		Fee:  std.Fee{GasFee: std.NewCoin("ugnot", 2_000_000), GasWanted: 10_000_000},
	}
	signBytes, err := tx.GetSignBytes(chainID, acc.GetAccountNumber(), acc.GetSequence())
	require.NoError(t, err)
	// multisign builds a multisignature from the constituents at idx.
	multisign := func(idx ...int) []byte {
		ms := multisig.NewMultisig(len(pubs))
		for _, i := range idx {
			sig, err := privs[i].Sign(signBytes)
			require.NoError(t, err)
			require.NoError(t, ms.AddSignatureFromPubKey(sig, pubs[i], pubs))
		}
		return ms.Marshal()
	}
	encode := func(ms []byte) []byte {
		signed := tx
		signed.Signatures = []std.Signature{{PubKey: msPub, Signature: ms}}
		return amino.MustMarshal(signed)
	}
	signedBy := func(idx ...int) []byte { return encode(multisign(idx...)) }
	requireRefused := func(bz []byte, wantLog string) {
		t.Helper()
		for name, res := range map[string]abci.ResponseBase{
			"CheckTx":   bapp.CheckTx(abci.RequestCheckTx{Tx: bz}).ResponseBase,
			"DeliverTx": bapp.DeliverTx(abci.RequestDeliverTx{Tx: bz}).ResponseBase,
		} {
			require.IsType(t, std.UnauthorizedError{}, res.Error, name)
			require.Contains(t, res.Log, wantLog, name)
		}
	}

	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 2}})
	// All three signed: any two of the signatures would also verify.
	requireRefused(signedBy(0, 1, 2), "multisignature has 3 signatures, want exactly the threshold 2")
	// What a relayer could make of the exactly-K original by dropping one.
	signed := signedBy(1, 2)
	requireRefused(signedBy(2), "multisignature has 1 signatures, want exactly the threshold 2")
	// Or by padding a varint in its multisignature: the tx stays canonical,
	// since its own check sees the multisignature as opaque bytes.
	repadded := encode(padBitArrayVarint(t, multisign(1, 2)))
	require.NoError(t, amino.UnmarshalCanonical(repadded, &std.Tx{}))
	requireRefused(repadded, "signature is not a valid multisignature")
	// None took the sequence, so the original still runs.
	res := bapp.DeliverTx(abci.RequestDeliverTx{Tx: signed})
	require.True(t, res.IsOK(), res.Log)
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()
	after := queryAccount(t, bapp, addr)
	require.Equal(t, acc.GetSequence()+1, after.GetSequence())
}

// padBitArrayVarint writes a multisignature's ExtraBitsStored as a two-byte
// varint, so its BitArray field grows by one byte: 0a L 08 E becomes
// 0a L+1 08 E|80 00.
func padBitArrayVarint(t *testing.T, bz []byte) []byte {
	t.Helper()
	require.True(t, bz[0] == 0x0a && bz[1] < 0x7f && bz[2] == 0x08 && bz[3] < 0x80, "% x", bz[:4])
	return append([]byte{0x0a, bz[1] + 1, 0x08, bz[3] | 0x80, 0x00}, bz[4:]...)
}
