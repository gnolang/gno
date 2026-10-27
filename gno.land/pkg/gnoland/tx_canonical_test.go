package gnoland

import (
	"encoding/binary"
	"slices"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wireField is one field as it sits on the amino wire: the content of a
// length-delimited field, or the bytes of a varint one.
type wireField struct {
	num uint32
	typ amino.Typ3
	raw []byte
}

func splitFields(t *testing.T, bz []byte) []wireField {
	t.Helper()

	var fields []wireField
	for len(bz) > 0 {
		num, typ, n, err := amino.DecodeFieldNumberAndTyp3(bz)
		require.NoError(t, err)
		bz = bz[n:]
		switch typ {
		case amino.Typ3ByteLength:
			content, n, err := amino.DecodeByteSlice(bz)
			require.NoError(t, err)
			fields = append(fields, wireField{num, typ, content})
			bz = bz[n:]
		case amino.Typ3Varint:
			_, n, err := amino.DecodeUvarint(bz)
			require.NoError(t, err)
			fields = append(fields, wireField{num, typ, bz[:n]})
			bz = bz[n:]
		default:
			t.Fatalf("unexpected wire type %v", typ)
		}
	}
	return fields
}

// padded rewrites a varint with one redundant continuation byte.
func padded(v []byte) []byte {
	return slices.Concat(v[:len(v)-1], []byte{v[len(v)-1] | 0x80, 0x00})
}

// joinFields re-encodes fields minimally, except the field at index pad,
// whose value varint, or length prefix for a length-delimited field, is
// written padded. Re-joining a parent recomputes the enclosing lengths.
func joinFields(fields []wireField, pad int) []byte {
	var out []byte
	for i, f := range fields {
		out = binary.AppendUvarint(out, uint64(f.num)<<3|uint64(f.typ))
		switch f.typ {
		case amino.Typ3ByteLength:
			l := binary.AppendUvarint(nil, uint64(len(f.raw)))
			if i == pad {
				l = padded(l)
			}
			out = append(out, l...)
			out = append(out, f.raw...)
		case amino.Typ3Varint:
			v := f.raw
			if i == pad {
				v = padded(v)
			}
			out = append(out, v...)
		}
	}
	return out
}

// canonicalSendApp boots an app at height 1 with a funded signer and returns
// it with the canonical bytes of one signed bank send, split into its three
// top-level fields: Msgs[0] as an Any, Fee, Signatures[0].
func canonicalSendApp(t *testing.T) (abci.Application, []byte, []wireField) {
	t.Helper()

	const chainID = "test-chain"
	key := getDummyKey(t)
	to := crypto.AddressFromPreimage([]byte("to"))

	gs := DefaultGenState()
	gs.Balances = []Balance{{
		Address: key.PubKey().Address(),
		Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
	}}
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	app.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
		AppState: gs,
	})
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})

	msg := bank.NewMsgSend(key.PubKey().Address(), to, std.NewCoins(std.NewCoin("ugnot", 1000)))
	tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)
	require.Empty(t, tx.Memo)
	canonical, err := amino.Marshal(tx)
	require.NoError(t, err)

	top := splitFields(t, canonical)
	require.Len(t, top, 3)
	require.Equal(t, canonical, joinFields(top, -1), "the wire helpers must round-trip the canonical bytes")
	return app, canonical, top
}

// requireSameTxRejected checks that bz decodes to the very tx canonical
// encodes, then that both CheckTx and DeliverTx refuse it as non-canonical.
func requireSameTxRejected(t *testing.T, app abci.Application, canonical, bz []byte, name string) {
	t.Helper()

	require.NotEqual(t, canonical, bz, name)
	var decoded std.Tx
	require.NoError(t, amino.Unmarshal(bz, &decoded), name)
	reencoded, err := amino.Marshal(decoded)
	require.NoError(t, err)
	require.Equal(t, canonical, reencoded, "%s: mutation must decode to the same signed tx", name)

	check := app.CheckTx(abci.RequestCheckTx{Tx: bz})
	require.False(t, check.IsOK(), name)
	assert.Equal(t, std.TxDecodeError{}, check.Error, name)
	assert.Contains(t, check.Log, "non-canonical encoding", name)

	deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
	require.False(t, deliver.IsOK(), name)
	assert.Equal(t, std.TxDecodeError{}, deliver.Error, name)
	assert.Contains(t, deliver.Log, "non-canonical encoding", name)
}

// A signed tx re-encoded non-canonically decodes to the same tx, so its
// signatures still verify, but it hashes differently; before the canonical
// check the chain included it under a hash the sender never computed. Every
// variant must now fail to decode in CheckTx and DeliverTx, at every nesting
// level, while the sender's own bytes still deliver.
func TestTxNonCanonicalEncodingRejected(t *testing.T) {
	t.Parallel()

	app, canonical, top := canonicalSendApp(t)
	withTop := func(i int, raw []byte) []byte {
		fields := slices.Clone(top)
		fields[i].raw = raw
		return joinFields(fields, -1)
	}
	msgAny := splitFields(t, top[0].raw) // type_url, value
	fee := splitFields(t, top[1].raw)    // gas_wanted, gas_fee
	require.Less(t, canonical[0], byte(0x80))

	mutations := map[string][]byte{
		// Memo is field 4 and empty, so the encoder omitted it; write it out.
		"explicit empty memo": slices.Concat(canonical, []byte{0x22, 0x00}),
		// The first tag as a two-byte varint.
		"fee: gas_wanted value padded":    withTop(1, joinFields(fee, 0)),
		"message: type_url length padded": withTop(0, joinFields(msgAny, 0)),
	}
	mutations["padded first tag"] = slices.Concat([]byte{canonical[0] | 0x80, 0x00}, canonical[1:])

	for name, bz := range mutations {
		requireSameTxRejected(t, app, canonical, bz, name)
	}

	deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: canonical})
	require.True(t, deliver.IsOK(), deliver.Log)
}

// TestTxLegacyZeroSessionAddrRejected: before #6399, every Go client wrote a
// master-key signature's zero SessionAddr as its bech32 string, 42 bytes in
// field 3. #6399 leaves it out, which makes those bytes another encoding of
// the same signed tx, under another hash, and so refused here. Until #6399
// is in this tree the written-out form is still what the encoder produces,
// so there is nothing to reject and the test stands down.
func TestTxLegacyZeroSessionAddrRejected(t *testing.T) {
	t.Parallel()

	app, canonical, top := canonicalSendApp(t)
	sig := splitFields(t, top[2].raw) // pub_key, signature[, session_addr before #6399]
	if len(sig) == 3 {
		t.Skip("the encoder still writes the zero session address (before #6399): the legacy form is the canonical one")
	}
	require.Len(t, sig, 2, "a zero session address has no wire presence")

	written := wireField{3, amino.Typ3ByteLength, []byte(crypto.Address{}.String())}
	fields := slices.Clone(top)
	fields[2].raw = joinFields(append(slices.Clone(sig), written), -1)
	requireSameTxRejected(t, app, canonical, joinFields(fields, -1), "legacy written-out zero session address")
}
