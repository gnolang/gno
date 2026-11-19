package gnolang

import "testing"

// Benchmarks the cached path of (*DeclaredType).TypeID() — i.e. the call after
// the typeid has already been computed and memoized. This is the overwhelmingly
// common case at runtime (map keys, typed ==, type assertions, per-commit
// assertTypeIsPublic all re-call TypeID() on already-sealed named types).
//
// The benchmark body is identical regardless of the implementation under test;
// only (*DeclaredType).TypeID() itself changes between the baseline and the fix.

var benchTypeIDSink TypeID

// Package/file-level declaration: ParentLoc is zero, so the cached-path
// assertion (baseline) costs one fmt.Sprintf.
func BenchmarkDeclaredTypeID_Cached_PackageLevel(b *testing.B) {
	dt := &DeclaredType{
		PkgPath: "gno.land/r/demo/boards",
		Name:    "BoardID",
	}
	benchTypeIDSink = dt.TypeID() // compute + memoize once
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchTypeIDSink = dt.TypeID()
	}
}

// Function-level declaration: ParentLoc is non-zero, so the cached-path
// assertion (baseline) costs three fmt.Sprintf (typeidf + Location.String +
// Span.String).
func BenchmarkDeclaredTypeID_Cached_FuncLevel(b *testing.B) {
	dt := &DeclaredType{
		PkgPath: "gno.land/r/demo/boards",
		Name:    "localType",
		ParentLoc: Location{
			PkgPath: "gno.land/r/demo/boards",
			File:    "boards.gno",
			Span: Span{
				Pos: Pos{Line: 42, Column: 3},
				End: Pos{Line: 50, Column: 1},
			},
		},
	}
	benchTypeIDSink = dt.TypeID() // compute + memoize once
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchTypeIDSink = dt.TypeID()
	}
}
