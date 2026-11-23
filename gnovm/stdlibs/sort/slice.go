package sort

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_sliceLen(m *gno.Machine, x gno.TypedValue) int {
	assertSlice(m, x)
	return x.GetLength()
}

// X_sliceSwap does what the statement x[i], x[j] = x[j], x[i] does in Gno:
// the same readonly check, and an Assign2 per element, so the realm marks
// the backing array dirty, rejects a write into another realm's array, and
// updates the reference counts of any objects the elements hold.
func X_sliceSwap(m *gno.Machine, x gno.TypedValue, i, j int) {
	assertSlice(m, x)
	if m.IsReadonly(&x) {
		m.PanicString("cannot swap elements of readonly tainted slice")
	}
	sv, ok := x.V.(*gno.SliceValue)
	if !ok {
		// A nil slice has no elements, so any index is out of range.
		m.PanicString("sort: slice index out of range")
	}
	et := x.T.Elem()
	pi := sv.GetElementPointer(m.Store, i, et)
	pj := sv.GetElementPointer(m.Store, j, et)
	vi, vj := pi.Deref(), pj.Deref()
	pi.Assign2(m, m.Alloc, m.Store, m.Realm, vj, false)
	pj.Assign2(m, m.Alloc, m.Store, m.Realm, vi, false)
}

func assertSlice(m *gno.Machine, x gno.TypedValue) {
	if x.T == nil {
		m.PanicString("sort: x is nil, not a slice")
	}
	if x.T.Kind() != gno.SliceKind {
		m.PanicString("sort: x is not a slice: " + x.T.String())
	}
}
