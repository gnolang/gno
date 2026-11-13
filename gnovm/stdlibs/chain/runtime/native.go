package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !isOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func isOriginCall(m *gno.Machine) bool {
	n := m.NumFrames()
	if n == 0 {
		return false
	}
	firstPkg := m.Frames[0].LastPackage
	isMsgCall := firstPkg != nil && firstPkg.PkgPath == ""
	if !isMsgCall {
		return false
	}
	// The code that runs has to belong to the package the message named.
	// A message names a path, the VM credits its coins to that path, and
	// then invokes whatever value the named symbol holds: an exported
	// `var Dep = other.Deposit` makes Frames[0] *other's* own function
	// while the envelope was credited to the aliasing realm. Counting
	// frames cannot see that, because there is only ever one realm's code
	// on the stack. OriginSendRecipientPath is the named path, set for
	// every MsgCall whether or not it carries coins, and it is empty in
	// contexts that have no message at all, which fails closed here.
	entry := m.Frames[0].Func
	if entry == nil || entry.PkgPath != execctx.GetContext(m).OriginSendRecipientPath {
		return false
	}
	// Count the frames that are a call boundary: named calls, plus func
	// literals reached by crossing into another realm. Control-flow basic
	// frames (for/range/switch) and same-realm closures stay transparent.
	return m.NumCallBoundaryFrames() <= 2
}

func ChainID(m *gno.Machine) string {
	return execctx.GetContext(m).ChainID
}

func ChainDomain(m *gno.Machine) string {
	return execctx.GetContext(m).ChainDomain
}

func ChainHeight(m *gno.Machine) int64 {
	return execctx.GetContext(m).Height
}

// pathRestricted is satisfied by GnoSessionAccount without importing gno.land.
// Entries use the typed grammar "*" or "<route>/<type>[:<path>]" — see
// gno.land/pkg/gnoland/allow_paths.go. AllowPaths is required at create-time;
// only session account types that don't implement this interface can return nil.
type pathRestricted interface{ GetAllowPaths() []string }

func X_getSessionInfo(m *gno.Machine) (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool) {
	ctx := execctx.GetContext(m)
	if ctx.SessionAccount == nil {
		return "", 0, nil, false
	}
	da := ctx.SessionAccount
	addr := da.(std.Account).GetAddress()
	var paths []string
	if pr, ok := da.(pathRestricted); ok {
		paths = pr.GetAllowPaths()
	}
	return addr.String(), da.GetExpiresAt(), paths, true
}

func typedString(s string) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(gno.StringValue(s))
	return tv
}

// ObjectID implements chain/runtime.ObjectID.
//
// The Gno side declares any; the Go side takes a gno.TypedValue, which genstd
// links straight through with no Go2Gno conversion (see
// mapping.signaturesMatch). That is what lets this read the object a value
// refers to rather than a converted copy of it, and it needs no genstd change.
//
// It is a pure read: nothing here allocates, persists, or advances a realm
// clock.
func ObjectID(m *gno.Machine, v gno.TypedValue) string {
	pv, ok := v.V.(gno.PointerValue)
	if !ok {
		// Not a pointer: a scalar, a string, a struct by value, a slice, a
		// nil interface. None of these name an object in their own right.
		return ""
	}

	// GetBase resolves a RefValue base through the store. A nil base, which is
	// what a nil pointer carries, comes back as nil and fails the assertion.
	// Only a standalone heap item qualifies: a *StructValue, *ArrayValue or
	// *Block base means a pointer into a container (&x.Field, &arr[i], a
	// pointer to a package-level var), and the container's identity is not
	// the field's.
	hiv, ok := pv.GetBase(m.Store).(*gno.HeapItemValue)
	if !ok {
		return ""
	}

	oid := hiv.GetObjectID()
	if !oid.IsFinalized() {
		// The allocator stamped the realm half at creation, but the owning
		// realm has not persisted the object yet, so the clock half is still
		// zero. Attaching to the ownership tree is what mints it; no native
		// can conjure one without changing the object lifecycle, since
		// GetIsReal() is ObjectID.IsFinalized().
		return ""
	}

	if !oid.PkgID.IsRealmPkg() {
		// Persisted by a stdlib or a p/ package rather than by a realm. Those
		// are immutable packages whose objects nobody can create twice under
		// one name, so they are outside what #6026 is about, and whether they
		// should be identifiable has not been settled. Reporting "" leaves
		// that open; widening later is compatible.
		return ""
	}

	return oid.String()
}
