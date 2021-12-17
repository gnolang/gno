package stdlibs

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gnolang/gno"
)

func InjectPackage(store gno.Store, pn *gno.PackageNode, pv *gno.PackageValue) {
	switch pv.PkgPath {
	case "strings":
		pn.DefineGoNativeFunc("ToLower", strings.ToLower)
		pn.DefineGoNativeFunc("ToLowerSpecial", strings.ToLowerSpecial)
		pn.DefineGoNativeFunc("ToUpper", strings.ToUpper)
		pn.DefineGoNativeFunc("ToUpperSpecial", strings.ToUpperSpecial)
		pn.DefineGoNativeFunc("ToTitle", strings.ToTitle)
		pn.DefineGoNativeFunc("ToTitleSpecial", strings.ToTitleSpecial)
		// NOTE: Split returns []string, which becomes
		// gonative{[]string}, which is confusing.
		// So, implement with DefineNative instead.
		// pn.DefineGoNativeFunc("Split", strings.Split)
		pn.DefineNative("Split",
			gno.Flds( // params
				"str", "string",
				"delim", "string",
			),
			gno.Flds( // results
				"parts", "[]string",
			),
			func(m *gno.Machine) {
				arg0, arg1 := m.LastBlock().GetParams2()
				str := arg0.TV.GetString()
				delim := arg1.TV.GetString()
				parts := strings.Split(str, delim)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(parts),
				)
				m.PushValue(res0)
			},
		)
		pn.PrepareNewValues(pv)
	case "strconv":
		pn.DefineGoNativeFunc("Itoa", strconv.Itoa)
		pn.DefineGoNativeFunc("Atoi", strconv.Atoi)
		pn.PrepareNewValues(pv)
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
		pn.PrepareNewValues(pv)
	}
}
