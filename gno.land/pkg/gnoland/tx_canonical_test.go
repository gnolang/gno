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

// A signed tx re-encoded non-canonically decodes to the same tx, so its
// signatures still verify, but it hashes differently; before the canonical
// check the chain included it under a hash the sender never computed. Every
// variant must now fail to decode in CheckTx and DeliverTx, at every nesting
// level, while the sender's own bytes still deliver. The written-out zero
// session address is the form every Go client produced before #6404.
func TestTxNonCanonicalEncodingRejected(t *testing.T) {
	t.Parallel()

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

	// Msgs[0] as an Any, Fee, Signatures[0].
	top := splitFields(t, canonical)
	require.Len(t, top, 3)
	require.Equal(t, canonical, joinFields(top, -1), "the wire helpers must round-trip the canonical bytes")
	withTop := func(i int, raw []byte) []byte {
		fields := slices.Clone(top)
		fields[i].raw = raw
		return joinFields(fields, -1)
	}
	msgAny := splitFields(t, top[0].raw) // type_url, value
	fee := splitFields(t, top[1].raw)    // gas_wanted, gas_fee
	sig := splitFields(t, top[2].raw)    // pub_key, signature
	require.Len(t, sig, 2, "a zero session address has no wire presence")
	writtenSessionAddr := wireField{3, amino.Typ3ByteLength, []byte(crypto.Address{}.String())}

	mutations := map[string][]byte{
		// Memo is field 4 and empty, so the encoder omitted it; write it out.
		"explicit empty memo":                         slices.Concat(canonical, []byte{0x22, 0x00}),
		"fee: gas_wanted value padded":                withTop(1, joinFields(fee, 0)),
		"message: type_url length padded":             withTop(0, joinFields(msgAny, 0)),
		"signature: zero session address written out": withTop(2, joinFields(append(slices.Clone(sig), writtenSessionAddr), -1)),
	}
	// The first tag as a two-byte varint.
	require.Less(t, canonical[0], byte(0x80))
	mutations["padded first tag"] = slices.Concat([]byte{canonical[0] | 0x80, 0x00}, canonical[1:])

	for name, bz := range mutations {
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

	deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: canonical})
	require.True(t, deliver.IsOK(), deliver.Log)
}
