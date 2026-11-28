package gnoland

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const utf8ProbePath = "gno.land/r/demo/utf8probe"

const utf8ProbeBody = `package utf8probe

import (
	"strconv"
	"unicode/utf8"
)

func Probe(cur realm, s string) string {
	return "len=" + strconv.Itoa(len(s)) + " valid=" + strconv.FormatBool(utf8.ValidString(s))
}
`

// A relayer can rewrite a signed tx without breaking its signature in two
// ways: swap a signed U+FFFD for an invalid byte, since the signature covers a
// JSON rendering in which every invalid UTF-8 byte becomes U+FFFD, or strip a
// signature's public key once the account has stored one. The app must refuse
// each rewritten copy before it costs the signer anything, and keep running the
// original.
func TestRelayedRewritesAreRefused(t *testing.T) {
	const chainID = "test"
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)
	signer, deployer := getDummyKey(t), getDummyKey(t)
	addr := signer.PubKey().Address()

	gen := gnoGenesisState(t)
	gen.Balances = []Balance{
		{Address: addr, Amount: std.NewCoins(std.NewCoin("ugnot", 10_000_000_000))},
		{Address: deployer.PubKey().Address(), Amount: std.NewCoins(std.NewCoin("ugnot", 10_000_000_000))},
	}
	deploy := vm.MsgAddPackage{Creator: deployer.PubKey().Address(), Package: &std.MemPackage{
		Name: "utf8probe", Path: utf8ProbePath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(utf8ProbePath)},
			{Name: "probe.gno", Body: utf8ProbeBody},
		},
	}}
	gen.Txs = []TxWithMetadata{{Tx: createAndSignTx(t, []std.Msg{deploy}, chainID, deployer)}}
	require.True(t, bapp.InitChain(abci.RequestInitChain{
		ChainID: chainID, AppState: gen,
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
	}).IsOK())
	bapp.Commit()
	// One empty block, so the ante handler checks signatures against the
	// signer's real account number and sequence.
	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()

	acc := queryAccount(t, bapp, addr)
	accNum, seq := acc.GetAccountNumber(), acc.GetSequence()

	sign := func(seq uint64, msg std.Msg) []byte {
		bz, err := amino.Marshal(createAndSignTxWithAccSeq(t, []std.Msg{msg}, chainID, signer, accNum, seq))
		require.NoError(t, err)
		return bz
	}
	// What a relayer does: decode, change a string, re-encode. The JSON the
	// signature covers does not change, so the copied signature would still
	// verify. (Built by hand: GetSignBytes now refuses invalid UTF-8.)
	relay := func(bz []byte, seq uint64, edit func(*std.Tx)) []byte {
		var tx std.Tx
		require.NoError(t, amino.Unmarshal(bz, &tx))
		signedJSON := func() []byte {
			return std.MustSortJSON(amino.MustMarshalJSON(tx.SignDoc(chainID, accNum, seq)))
		}
		before := signedJSON()
		edit(&tx)
		require.Equal(t, before, signedJSON())
		out, err := amino.Marshal(tx)
		require.NoError(t, err)
		return out
	}
	requireRefused := func(bz []byte, wantErr abci.Error, wantLog string) {
		t.Helper()
		for name, res := range map[string]abci.ResponseBase{
			"CheckTx":   bapp.CheckTx(abci.RequestCheckTx{Tx: bz}).ResponseBase,
			"DeliverTx": bapp.DeliverTx(abci.RequestDeliverTx{Tx: bz}).ResponseBase,
		} {
			require.IsType(t, wantErr, res.Error, name)
			require.Contains(t, res.Log, wantLog, name)
		}
	}
	call := func(arg string) vm.MsgCall {
		return vm.MsgCall{Caller: addr, PkgPath: utf8ProbePath, Func: "Probe", Args: []string{arg}}
	}

	signedCall := sign(seq, call("a\uFFFDb"))
	const docPath = "gno.land/r/demo/utf8doc"
	signedDeploy := sign(seq+1, vm.MsgAddPackage{Creator: addr, Package: &std.MemPackage{
		Name: "utf8doc", Path: docPath,
		Files: []*std.MemFile{
			{Name: "README.md", Body: "hello \uFFFD"},
			{Name: "doc.gno", Body: "package utf8doc\n"},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(docPath)},
		},
	}})

	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 2}})

	// The relayed copies are refused, in CheckTx and in DeliverTx alike.
	requireRefused(relay(signedCall, seq, func(tx *std.Tx) { tx.Msgs[0] = call("a\xffb") }),
		std.TxDecodeError{}, "invalid UTF-8 in msg[0].args[0]")
	requireRefused(relay(signedDeploy, seq+1, func(tx *std.Tx) {
		tx.Msgs[0].(vm.MsgAddPackage).Package.Files[0].Body = "hello \xff"
	}), std.TxDecodeError{}, "invalid UTF-8 in msg[0].package.files[0].body")

	// They took no sequence, so the signed originals still run, as signed.
	res := bapp.DeliverTx(abci.RequestDeliverTx{Tx: signedCall})
	require.True(t, res.IsOK(), res.Log)
	require.Contains(t, string(res.Data), "len=5 valid=true")
	res = bapp.DeliverTx(abci.RequestDeliverTx{Tx: signedDeploy})
	require.True(t, res.IsOK(), res.Log)
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()

	acc = queryAccount(t, bapp, addr)
	require.Equal(t, seq+2, acc.GetSequence())
	readme := bapp.Query(abci.RequestQuery{Path: "vm/qfile", Data: []byte(docPath + "/README.md")})
	require.True(t, readme.IsOK(), readme.Log)
	require.Equal(t, "hello \uFFFD", string(readme.Data))

	// The account has stored its key, so a signature may now leave it out,
	// which a relayer could use to restate the tx under a new hash.
	signedPlain := sign(seq+2, call("plain"))
	stripped := relay(signedPlain, seq+2, func(tx *std.Tx) { tx.Signatures[0].PubKey = nil })
	require.NotEqual(t, signedPlain, stripped)
	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 3}})
	requireRefused(stripped, std.InvalidPubKeyError{}, "signature 0 has no public key")
	res = bapp.DeliverTx(abci.RequestDeliverTx{Tx: signedPlain})
	require.True(t, res.IsOK(), res.Log)
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()
	acc = queryAccount(t, bapp, addr)
	require.Equal(t, seq+3, acc.GetSequence())
}

// utf8TestMap builds a GnoVM map value holding vals.
func utf8TestMap(vals ...gnolang.TypedValue) *gnolang.MapValue {
	ml := &gnolang.MapList{}
	for i, v := range vals {
		item := &gnolang.MapListItem{
			Key:   gnolang.TypedValue{T: gnolang.StringType, V: gnolang.StringValue(strconv.Itoa(i))},
			Value: v,
		}
		if ml.Tail == nil {
			ml.Head = item
		} else {
			item.Prev = ml.Tail
			ml.Tail.Next = item
		}
		ml.Tail = item
		ml.Size++
	}
	return &gnolang.MapValue{List: ml}
}

// utf8TestDecode round-trips mv through the binary codec, so its list is
// linked by MapList.UnmarshalAmino, as a decoded value's would be.
func utf8TestDecode(t *testing.T, mv *gnolang.MapValue) *gnolang.MapValue {
	t.Helper()
	bz, err := amino.Marshal(mv)
	require.NoError(t, err)
	out := new(gnolang.MapValue)
	require.NoError(t, amino.Unmarshal(bz, out))
	return out
}

// ValidateUTF8 walks what amino encodes. A MapList encodes every item, but
// its Go value reaches the middle ones only through Next, which amino skips,
// so a walk over the Go fields missed them.
func TestValidateUTF8ReachesEveryMapItem(t *testing.T) {
	str := func(s string) gnolang.TypedValue {
		return gnolang.TypedValue{T: gnolang.StringType, V: gnolang.StringValue(s)}
	}
	mv := utf8TestDecode(t, utf8TestMap(str("a"), str("\xff"), str("c")))
	require.EqualError(t, std.ValidateUTF8(mv), "invalid UTF-8 in List.List[1].Value.V")
}

// A decoded one-item MapList has Head == Tail. Walking Go fields visited the
// item twice per level, so nested one-item maps cost 2^depth; walking the
// representation is linear.
func TestValidateUTF8NestedMapsAreLinear(t *testing.T) {
	tv := gnolang.TypedValue{T: gnolang.StringType, V: gnolang.StringValue("leaf")}
	for range 30 {
		tv = gnolang.TypedValue{
			T: &gnolang.MapType{Key: gnolang.StringType, Value: gnolang.StringType},
			V: utf8TestMap(tv),
		}
	}
	mv := utf8TestDecode(t, tv.V.(*gnolang.MapValue))

	done := make(chan error, 1)
	go func() { done <- std.ValidateUTF8(mv) }()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("ValidateUTF8 did not finish: the walk is not linear in the nesting")
	}
}

// MemPackage.Info is typed `any`, so it decodes into any registered type, and
// rendering some of them panics, as a zero params.Param does, or costs far more
// than decoding them. ValidateBasic refuses a non-nil Info before the UTF-8
// walk would render it.
func TestUTF8WalkSkipsRefusedInfo(t *testing.T) {
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)
	require.True(t, bapp.InitChain(abci.RequestInitChain{
		ChainID: "test", AppState: DefaultGenState(),
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
	}).IsOK())

	// "/gno.FuncType" and "/params.Param" are both 13 bytes, and a zero
	// FuncType encodes to no value, so the swap decodes to a zero Param.
	tx := std.Tx{Msgs: []std.Msg{vm.MsgAddPackage{
		Creator: crypto.AddressFromPreimage([]byte("creator")),
		Package: &std.MemPackage{
			Name: "p", Path: "gno.land/r/demo/p",
			Files: []*std.MemFile{{Name: "p.gno", Body: "package p"}},
			Info:  &gnolang.FuncType{},
		},
	}}, Fee: std.NewFee(1_000_000, std.MustParseCoin("1ugnot"))}
	bz := amino.MustMarshal(tx)
	require.Equal(t, 1, bytes.Count(bz, []byte("/gno.FuncType")))
	bz = bytes.ReplaceAll(bz, []byte("/gno.FuncType"), []byte("/params.Param"))
	var decoded std.Tx
	require.NoError(t, amino.Unmarshal(bz, &decoded))
	require.IsType(t, params.Param{}, decoded.Msgs[0].(vm.MsgAddPackage).Package.Info)

	res := bapp.CheckTx(abci.RequestCheckTx{Tx: bz})
	_, ok := res.Error.(vm.InvalidPackageError)
	require.True(t, ok, "%v", res.Error)
}
