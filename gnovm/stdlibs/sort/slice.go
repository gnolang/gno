package sort

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_sliceLen(m *gno.Machine, x gno.TypedValue) int {
	assertSlice(m, x)
	return x.GetLength()
}

// X_sliceSwap swaps x[i] and x[j] in place in the backing array. It makes
// the readonly check of the statement x[i], x[j] = x[j], x[i], but moves the
// elements instead of copying them. A stored struct element keeps its
// object, and a pointer to one of its fields follows it to its new index.
func X_sliceSwap(m *gno.Machine, x gno.TypedValue, i, j int) {
	assertSlice(m, x)
	if m.IsReadonly(&x) {
		m.PanicString("cannot swap elements of readonly tainted slice")
	}
	pi := x.GetPointerAtIndexInt(m.Store, i)
	pj := x.GetPointerAtIndexInt(m.Store, j)
	if i == j {
		// The sort routines swap an element with itself. Returning after the
		// lookups keeps their index checks.
		return
	}
	if pi.TV.T == gno.DataByteType {
		vi, vj := pi.Deref(), pj.Deref()
		pi.TV.SetDataByte(vj.GetUint8())
		pj.TV.SetDataByte(vi.GetUint8())
	} else {
		*pi.TV, *pj.TV = *pj.TV, *pi.TV
	}
	// The elements stay in the same array, so the objects they hold keep
	// their owner and reference counts, and only the array needs saving.
	// DidUpdate marks it dirty and refuses an array of another realm.
	m.Realm.DidUpdate(m, pi.Base.(gno.Object), nil, nil)
}

func assertSlice(m *gno.Machine, x gno.TypedValue) {
	if x.T == nil {
		m.PanicString("sort: x is nil, not a slice")
	}
	if x.T.Kind() != gno.SliceKind {
		m.PanicString("sort: x is not a slice: " + x.T.String())
	}
}
