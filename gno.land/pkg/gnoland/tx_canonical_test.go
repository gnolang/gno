package gnoland

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"slices"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
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
		"explicit empty memo":             slices.Concat(canonical, []byte{0x22, 0x00}),
		"fee: gas_wanted value padded":    withTop(1, joinFields(fee, 0)),
		"message: type_url length padded": withTop(0, joinFields(msgAny, 0)),
	}
	// The first tag as a two-byte varint.
	mutations["padded first tag"] = slices.Concat([]byte{canonical[0] | 0x80, 0x00}, canonical[1:])

	for name, bz := range mutations {
		requireSameTxRejected(t, app, canonical, bz, name)
	}

	deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: canonical})
	require.True(t, deliver.IsOK(), deliver.Log)
}

// MemPackage.Type and Info are `any`, and some registered types cost far more
// to re-encode than to decode, so ValidateBasic must refuse them before runTx's
// canonical check. These bytes are non-canonical (1e000009 re-encodes as
// 1000000000), so a decode error here would mean the re-encode ran first.
func TestTxPackageAnyFieldsRefusedBeforeReencode(t *testing.T) {
	t.Parallel()

	app, _, _ := canonicalSendApp(t)
	creator := crypto.AddressFromPreimage([]byte("creator"))
	placeholder := gno.BigdecValue{V: big.NewRat(12345678, 1)}
	for name, pkg := range map[string]*std.MemPackage{
		"Info": {Name: "p", Path: "gno.land/r/x/p", Info: placeholder},
		"Type": {Name: "p", Path: "gno.land/r/x/p", Type: placeholder},
	} {
		bz, err := amino.Marshal(std.Tx{Msgs: []std.Msg{vm.MsgAddPackage{Creator: creator, Package: pkg}}})
		require.NoError(t, err)
		require.Equal(t, 1, bytes.Count(bz, []byte("12345678")), name)
		bz = bytes.Replace(bz, []byte("12345678"), []byte("1e000009"), 1)
		require.ErrorContains(t, amino.UnmarshalCanonical(bz, new(std.Tx)), "non-canonical encoding", name)

		check := app.CheckTx(abci.RequestCheckTx{Tx: bz})
		assert.Equal(t, vm.InvalidPackageError{}, check.Error, name)
		deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
		assert.Equal(t, vm.InvalidPackageError{}, deliver.Error, name)
	}
}

// A nil *std.MemPackage inside MsgAddPackage or MsgRun is a non-nil message
// in its canonical encoding, so it passes both the canonical check and the
// nil-message guard in runTx; ValidateBasic then dereferenced it, and the
// recovered panic cost about thirteen times a clean rejection. Both must fail
// as a bad request in both modes. The written-out zero session address, the
// other legacy form, is covered by TestDeliverTxZeroSessionAddr.
func TestTxNilPackageRejected(t *testing.T) {
	t.Parallel()

	app, _, _ := canonicalSendApp(t)
	caller := crypto.AddressFromPreimage([]byte("caller"))
	for name, msg := range map[string]std.Msg{
		"MsgAddPackage": vm.MsgAddPackage{Creator: caller},
		"MsgRun":        vm.MsgRun{Caller: caller},
	} {
		bz, err := amino.Marshal(std.Tx{Msgs: []std.Msg{msg}})
		require.NoError(t, err)
		var decoded std.Tx
		require.NoError(t, amino.UnmarshalCanonical(bz, &decoded), name)

		check := app.CheckTx(abci.RequestCheckTx{Tx: bz})
		require.False(t, check.IsOK(), name)
		assert.Equal(t, vm.InvalidPkgPathError{}, check.Error, name)
		assert.NotContains(t, check.Log, "recovered", name)

		deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
		require.False(t, deliver.IsOK(), name)
		assert.Equal(t, vm.InvalidPkgPathError{}, deliver.Error, name)
		assert.NotContains(t, deliver.Log, "recovered", name)
	}
}
