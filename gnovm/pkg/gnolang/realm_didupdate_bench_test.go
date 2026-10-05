package gnolang

import "testing"

// realm_didupdate_bench_test.go: microbenchmarks for Realm.DidUpdate,
// the per-write ownership hook called on every assignment to a
// realm-owned object. Scenarios cover the steady-state paths:
//
//   - NilRealm:      rlm == nil (read paths, non-realm packages).
//   - Unreal:        po not yet real → early return (temp objects).
//   - RealPrimitive: po real, xo/co nil (primitive field write on
//     persisted object; po already dirty → MarkDirty early-returns).
//   - RealAttach:    po real, co real (reference write; both dirty).
//   - RealSwap:      po real, xo and co real (replace a reference).

// benchFixture holds a realm plus three of its real, already-dirty
// objects for the DidUpdate scenarios. Returned as a struct so each
// benchmark reads only the fields it needs (avoids blank-identifier
// noise).
type benchFixture struct {
	rlm        *Realm
	po, xo, co *StructValue
}

func benchRealmAndObjects() benchFixture {
	rlm := NewRealm("gno.land/r/bench")
	rlm.Time = 100
	mkReal := func(t uint64) *StructValue {
		sv := &StructValue{}
		sv.SetPkgID(rlm.ID)
		sv.SetNewTime(t)
		sv.SetIsDirty(true, rlm.Time) // steady state: already marked dirty
		sv.IncRefCount()
		return sv
	}
	return benchFixture{rlm: rlm, po: mkReal(2), xo: mkReal(3), co: mkReal(4)}
}

func BenchmarkDidUpdate_NilRealm(b *testing.B) {
	f := benchRealmAndObjects()
	m := benchMachine()
	defer m.Release()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nilRealm.DidUpdate(m, f.po, nil, nil)
	}
}

func BenchmarkDidUpdate_Unreal(b *testing.B) {
	f := benchRealmAndObjects()
	m := benchMachine()
	defer m.Release()
	po := &StructValue{}
	po.SetPkgID(f.rlm.ID) // allocated but not finalized → not real
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.rlm.DidUpdate(m, po, nil, nil)
	}
}

func BenchmarkDidUpdate_RealPrimitive(b *testing.B) {
	f := benchRealmAndObjects()
	m := benchMachine()
	defer m.Release()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.rlm.DidUpdate(m, f.po, nil, nil)
	}
}

func BenchmarkDidUpdate_RealAttach(b *testing.B) {
	f := benchRealmAndObjects()
	m := benchMachine()
	defer m.Release()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.rlm.DidUpdate(m, f.po, nil, f.co)
	}
}

func BenchmarkDidUpdate_RealSwap(b *testing.B) {
	f := benchRealmAndObjects()
	m := benchMachine()
	defer m.Release()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.rlm.DidUpdate(m, f.po, f.xo, f.co)
	}
}
