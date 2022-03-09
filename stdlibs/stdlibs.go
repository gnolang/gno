package stdlibs

import (
	"reflect"
	"strconv"
	"time"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/bech32"
	"github.com/gnolang/gno/pkgs/crypto"
)

func InjectPackage(store gno.Store, pn *gno.PackageNode) {
	switch pn.PkgPath {
	case "strconv":
		pn.DefineGoNativeValue("Itoa", strconv.Itoa)
		pn.DefineGoNativeValue("Atoi", strconv.Atoi)
		pn.DefineGoNativeValue("FormatInt", strconv.FormatInt)
		pn.DefineGoNativeValue("FormatUint", strconv.FormatUint)
		pn.DefineGoNativeValue("Quote", strconv.Quote)
		pn.DefineGoNativeValue("QuoteToASCII", strconv.QuoteToASCII)
		pn.DefineGoNativeValue("CanBackquote", strconv.CanBackquote)
		pn.DefineGoNativeValue("IntSize", strconv.IntSize)
	case "std":
		// NOTE: pkgs/sdk/vm/VMKeeper also
		// injects more like .Send, .GetContext.
		pn.DefineNative("Hash",
			gno.Flds( // params
				"bz", "[]byte",
			),
			gno.Flds( // results
				"hash", "[20]byte",
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				bz := []byte(nil)
				if arg0.V != nil {
					slice := arg0.V.(*gno.SliceValue)
					array := slice.GetBase(m.Store)
					bz = array.GetReadonlyBytes()
				}
				hash := gno.HashBytes(bz)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf([20]byte(hash)),
				)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("CurrentRealmPath",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "string",
			),
			func(m *gno.Machine) {
				realmPath := ""
				if m.Realm != nil {
					realmPath = m.Realm.Path
				}
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(realmPath),
				)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetChainID",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "string",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx.ChainID),
				)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetHeight",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "int64",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx.Height),
				)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetTxSendCoins",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Coins",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx.TxSend),
				)
				coinT := store.GetType(gno.DeclaredTypeID("std", "Coin"))
				coinsT := store.GetType(gno.DeclaredTypeID("std", "Coins"))
				res0.T = coinsT
				av := res0.V.(*gno.SliceValue).Base.(*gno.ArrayValue)
				for i := range av.List {
					av.List[i].T = coinT
				}
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetCaller",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Address",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx.Caller),
				)
				addrT := store.GetType(gno.DeclaredTypeID("std", "Address"))
				res0.T = addrT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetPkgAddr",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Address",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx.PkgAddr),
				)
				addrT := store.GetType(gno.DeclaredTypeID("std", "Address"))
				res0.T = addrT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetBanker",
			gno.Flds( // params
				"bankerType", "BankerType",
			),
			gno.Flds( // results
				"", "Banker",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				arg0 := m.LastBlock().GetParams1().TV
				bankerType := BankerType(arg0.GetUint8())
				banker := ctx.Banker
				switch bankerType {
				case BankerTypeReadonly:
					banker = NewReadonlyBanker(banker)
				case BankerTypeTxSend:
					banker = NewTxSendBanker(banker, ctx.PkgAddr, ctx.TxSend, ctx.TxSendSpent)
				case BankerTypeRealmSend:
					banker = NewRealmSendBanker(banker, ctx.PkgAddr)
				case BankerTypeRealmIssue:
					banker = banker
				default:
					panic("should not happen") // defensive
				}
				rv := reflect.ValueOf(banker)
				res0 := gno.Go2GnoNativeValue(rv)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetTimestamp",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "int64",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := typedInt64(ctx.Timestamp)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("FormatTimestamp",
			gno.Flds( // params
				"timestamp", "int64",
				"format", "string",
			),
			gno.Flds( // results
				"", "string",
			),
			func(m *gno.Machine) {
				arg0, arg1 := m.LastBlock().GetParams2()
				timestamp := arg0.TV.GetInt64()
				format := arg1.TV.GetString()
				t := time.Unix(timestamp, 0).Round(0).UTC()
				result := t.Format(format)
				res0 := typedString(m.Alloc.NewStringValue(result))
				m.PushValue(res0)
			},
		)
		pn.DefineNative("ToBech32",
			gno.Flds( // params
				"addr", "Address",
			),
			gno.Flds( // results
				"", "string",
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1()
				bz := arg0.TV.V.(*gno.ArrayValue).GetReadonlyBytes()
				if len(bz) != crypto.AddressSize {
					panic("should not happen")
				}
				b32, err := bech32.ConvertAndEncode("g", bz)
				if err != nil {
					panic(err)
				}
				res0 := typedString(m.Alloc.NewStringValue(b32))
				m.PushValue(res0)
			},
		)
	}
}

func typedInt64(i64 int64) gno.TypedValue {
	tv := gno.TypedValue{T: gno.Int64Type}
	tv.SetInt64(i64)
	return tv
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}
