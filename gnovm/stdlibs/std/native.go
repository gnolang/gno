package std

import (
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !IsOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func IsOriginCall(m *gno.Machine) bool {
	return len(m.Frames) == 2
}

func CurrentRealmPath(m *gno.Machine) string {
	if m.Realm != nil {
		return m.Realm.Path
	}
	return ""
}

func GetChainID(m *gno.Machine) string {
	return m.Context.(ExecContext).ChainID
}

func GetHeight(m *gno.Machine) int64 {
	return m.Context.(ExecContext).Height
}

func GetOrigSend(m *gno.Machine) std.Coins {
	return m.Context.(ExecContext).OrigSend
}

func GetOrigCaller(m *gno.Machine) crypto.Bech32Address {
	return m.Context.(ExecContext).OrigCaller
}

func CurrentRealm(m *gno.Machine) Realm {
	var (
		ctx = m.Context.(ExecContext)
		// Default lastCaller is OrigCaller, the signer of the tx
		lastCaller  = ctx.OrigCaller
		lastPkgPath = ""
	)

	for i := m.NumFrames() - 1; i > 0; i-- {
		fr := m.Frames[i]
		if fr.LastPackage != nil && fr.LastPackage.IsRealm() {
			lastCaller = fr.LastPackage.GetPkgAddr().Bech32()
			lastPkgPath = fr.LastPackage.PkgPath
			break
		}
	}

	return Realm{
		addr:    lastCaller,
		pkgPath: lastPkgPath,
	}
}

func PrevRealm(m *gno.Machine) Realm {
	var (
		ctx = m.Context.(ExecContext)
		// Default lastCaller is OrigCaller, the signer of the tx
		lastCaller  = ctx.OrigCaller
		lastPkgPath = ""
	)

	for i := m.NumFrames() - 1; i > 0; i-- {
		fr := m.Frames[i]
		if fr.LastPackage == nil || !fr.LastPackage.IsRealm() {
			// Ignore non-realm frame
			continue
		}
		pkgPath := fr.LastPackage.PkgPath
		// The first realm we encounter will be the one calling
		// this function; to get the calling realm determine the first frame
		// where fr.LastPackage changes.
		if lastPkgPath == "" {
			lastPkgPath = pkgPath
		} else if lastPkgPath == pkgPath {
			continue
		} else {
			lastCaller = fr.LastPackage.GetPkgAddr().Bech32()
			lastPkgPath = pkgPath
			break
		}
	}

	// Empty the pkgPath if we return a user
	if ctx.OrigCaller == lastCaller {
		lastPkgPath = ""
	}

	return Realm{
		addr:    lastCaller,
		pkgPath: lastPkgPath,
	}
}

func GetOrigPkgAddr(m *gno.Machine) crypto.Bech32Address {
	return m.Context.(ExecContext).OrigPkgAddr
}

func GetCallerAt(m *gno.Machine, n int) crypto.Bech32Address {
	if n <= 0 {
		m.Panic(typedString("GetCallerAt requires positive arg"))
		return ""
	}
	if n > m.NumFrames() {
		// NOTE: the last frame's LastPackage
		// is set to the original non-frame
		// package, so need this check.
		m.Panic(typedString("frame not found"))
		return ""
	}
	if n == m.NumFrames() {
		// This makes it consistent with GetOrigCaller.
		ctx := m.Context.(ExecContext)
		return ctx.OrigCaller
	}
	return m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32()
}

func GetBanker(m *gno.Machine, bankerType BankerType) gno.TypedValue {
	ctx := m.Context.(ExecContext)
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
	m.Alloc.AllocateStruct()         // defensive; native space not allocated.
	m.Alloc.AllocateStructFields(10) // defensive 10; native space not allocated.

	// make gno bankAdapter{rv}
	btv := gno.Go2GnoNativeValue(m.Alloc, reflect.ValueOf(banker))
	bsv := m.Alloc.NewStructWithFields(btv)
	bankAdapterType := m.Store.GetType(gno.DeclaredTypeID("std", "bankAdapter"))
	res0 := gno.TypedValue{T: bankAdapterType, V: bsv}

	return res0
}

func EncodeBech32(prefix string, bytes [20]byte) crypto.Bech32Address {
	b32, err := bech32.ConvertAndEncode(prefix, bytes[:])
	if err != nil {
		panic(err) // should not happen
	}
	return crypto.Bech32Address(b32)
}

func DerivePkgAddr(pkgPath string) crypto.Bech32Address {
	return gno.DerivePkgAddr(pkgPath).Bech32()
}

func DecodeBech32(addr crypto.Bech32Address) (prefix string, bytes [20]byte, ok bool) {
	prefix, bz, err := bech32.Decode(string(addr))
	if err != nil || len(bz) != 20 {
		return "", [20]byte{}, false
	}
	return prefix, [20]byte(bz), true
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}
