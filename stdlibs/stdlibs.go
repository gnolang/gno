package stdlibs

import (
	"reflect"
	"strconv"
	"time"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/bech32"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

func InjectNativeMappings(store gno.Store) {
	store.AddGo2GnoMapping(reflect.TypeOf(crypto.Bech32Address("")), "std", "Address")
	store.AddGo2GnoMapping(reflect.TypeOf(std.Coins{}), "std", "Coins")
	store.AddGo2GnoMapping(reflect.TypeOf(std.Coin{}), "std", "Coin")
}

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
		// NOTE: some of these are overridden in tests/imports_test.go
		// Also see stdlibs/InjectPackage.
		pn.DefineNative("AssertOriginCall",
			gno.Flds( // params
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				isOrigin := len(m.Frames) == 2
				if !isOrigin {
					panic("invalid non-origin call")
				}
			},
		)
		pn.DefineNative("IsOriginCall",
			gno.Flds( // params
			),
			gno.Flds( // results
				"isOrigin", "bool",
			),
			func(m *gno.Machine) {
				isOrigin := len(m.Frames) == 2
				res0 := gno.TypedValue{T: gno.BoolType}
				res0.SetBool(isOrigin)
				m.PushValue(res0)
			},
		)
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
					m.Alloc,
					m.Store,
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
					m.Alloc,
					m.Store,
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
					m.Alloc,
					m.Store,
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
					m.Alloc,
					m.Store,
					reflect.ValueOf(ctx.Height),
				)
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetOrigSend",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Coins",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					m.Alloc,
					m.Store,
					reflect.ValueOf(ctx.OrigSend),
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
		pn.DefineNative("GetOrigCaller",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Address",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					m.Alloc,
					m.Store,
					reflect.ValueOf(ctx.OrigCaller),
				)
				addrT := store.GetType(gno.DeclaredTypeID("std", "Address"))
				res0.T = addrT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetOrigPkgAddr",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Address",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := gno.Go2GnoValue(
					m.Alloc,
					m.Store,
					reflect.ValueOf(ctx.OrigPkgAddr),
				)
				addrT := store.GetType(gno.DeclaredTypeID("std", "Address"))
				res0.T = addrT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetCallerAt",
			gno.Flds( // params
				"n", "int",
			),
			gno.Flds( // results
				"", "Address",
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				n := arg0.GetInt()
				if n <= 0 {
					panic("GetCallerAt requires positive arg")
				}
				if n > m.NumFrames() {
					// NOTE: the last frame's LastPackage
					// is set to the original non-frame
					// package, so need this check.
					panic("frame not found")
				}
				var pkgAddr string
				if n == m.NumFrames() {
					// This makes it consistent with GetOrigCaller.
					ctx := m.Context.(ExecContext)
					pkgAddr = string(ctx.OrigCaller)
				} else {
					pkgAddr = string(m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
				}
				res0 := gno.Go2GnoValue(
					m.Alloc,
					m.Store,
					reflect.ValueOf(pkgAddr),
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
				case BankerTypeOrigSend:
					banker = NewOrigSendBanker(banker, ctx.OrigPkgAddr, ctx.OrigSend, ctx.OrigSendSpent)
				case BankerTypeRealmSend:
					banker = NewRealmSendBanker(banker, ctx.OrigPkgAddr)
				case BankerTypeRealmIssue:
					banker = banker
				default:
					panic("should not happen") // defensive
				}
				rv := reflect.ValueOf(banker)
				m.Alloc.AllocateStruct()         // defensive; native space not allocated.
				m.Alloc.AllocateStructFields(10) // defensive 10; native space not allocated.

				// make gno bankAdapter{rv}
				btv := gno.Go2GnoNativeValue(m.Alloc, rv)
				bsv := m.Alloc.NewStructWithFields(btv)
				bankAdapterType := store.GetType(gno.DeclaredTypeID("std", "bankAdapter"))
				res0 := gno.TypedValue{T: bankAdapterType, V: bsv}
				m.PushValue(res0)
			},
		)
		pn.DefineNative("GetTimestamp",
			gno.Flds( // params
			),
			gno.Flds( // results
				"", "Time",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(ExecContext)
				res0 := typedInt64(ctx.Timestamp)
				timeT := store.GetType(gno.DeclaredTypeID("std", "Time"))
				res0.T = timeT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("FormatTimestamp",
			gno.Flds( // params
				"timestamp", "Time",
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
				res0 := typedString(m.Alloc.NewString(result))
				m.PushValue(res0)
			},
		)
		pn.DefineNative("EncodeBech32",
			gno.Flds( // params
				"prefix", "string",
				"bytes", "[20]byte",
			),
			gno.Flds( // results
				"addr", "string",
			),
			func(m *gno.Machine) {
				arg0, arg1 := m.LastBlock().GetParams2()
				prefix := arg0.TV.GetString()
				bz := arg1.TV.V.(*gno.ArrayValue).GetReadonlyBytes()
				if len(bz) != crypto.AddressSize {
					panic("should not happen")
				}
				b32, err := bech32.ConvertAndEncode(prefix, bz)
				if err != nil {
					panic(err)
				}
				res0 := typedString(m.Alloc.NewString(b32))
				m.PushValue(res0)
			},
		)
		pn.DefineNative("DecodeBech32",
			gno.Flds( // params
				"addr", "Address",
			),
			gno.Flds( // results
				"prefix", "string",
				"bytes", "[20]byte",
				"ok", "bool",
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1()
				addr := arg0.TV.GetString()
				prefix, bz, err := bech32.Decode(addr)
				if err != nil || len(bz) != 20 {
					m.PushValue(typedString(m.Alloc.NewString("")))
					m.PushValue(typedByteArray(20, m.Alloc.NewDataArray(20)))
					m.PushValue(typedBool(false))
				} else {
					m.PushValue(typedString(m.Alloc.NewString(prefix)))
					m.PushValue(typedByteArray(20, m.Alloc.NewArrayFromData(bz)))
					m.PushValue(typedBool(true))
				}
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

func typedBool(b bool) gno.TypedValue {
	tv := gno.TypedValue{T: gno.BoolType}
	tv.SetBool(b)
	return tv
}

func typedByteArray(ln int, bz *gno.ArrayValue) gno.TypedValue {
	if bz != nil && bz.GetLength() != ln {
		panic("array length mismatch")
	}
	tv := gno.TypedValue{T: &gno.ArrayType{Len: ln, Elt: gno.Uint8Type}, V: bz}
	return tv
}

func typedByteSlice(bz *gno.SliceValue) gno.TypedValue {
	tv := gno.TypedValue{T: &gno.SliceType{Elt: gno.Uint8Type}, V: bz}
	return tv
}

func typedNil(t gno.Type) gno.TypedValue {
	tv := gno.TypedValue{T: t, V: nil}
	return tv
}
