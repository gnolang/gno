package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// newTryEvalStaticTestPkg returns a store and the *PackageNode of a package
// declaring `func Foo(cur realm)`, ready for static evaluation. pkgName must
// be the last segment of pkgPath.
func newTryEvalStaticTestPkg(t *testing.T, pkgName, pkgPath string) (*defaultStore, *PackageNode) {
	t.Helper()
	db := memdb.NewMemDB()
	tm2 := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	st := NewStore(nil, tm2, tm2)
	m := NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: st})
	defer m.Release()
	pn, _ := m.RunMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  pkgName,
		Path:  pkgPath,
		Files: []*std.MemFile{{Name: "a.gno", Body: "package " + pkgName + "\n\nfunc Foo(cur realm) {}\n"}},
	}, true)
	return st, pn
}

// TestTryEvalStatic_ErrNilOnSuccess pins tryEvalStatic's error contract: err
// must be nil when the expression resolves.
//
// The deferred recover() used to assign err unconditionally, so a successful
// evaluation still returned err = "recovered panic with: <nil>". That made err
// unusable as a success signal: err == nil was reachable only via the
// *ConstExpr early return.
//
// This is not observable from Gno source, because the sole caller's branch was
// inverted in the matching direction and the two bugs cancelled out. This test
// pins the contract directly so neither can be reintroduced alone.
func TestTryEvalStatic_ErrNilOnSuccess(t *testing.T) {
	t.Parallel()
	st, pn := newTryEvalStaticTestPkg(t, "tryeval", "gno.land/r/test/tryeval")

	// A NameExpr for a package-level func, path-resolved against pn. Not a
	// *ConstExpr, so it exercises the throwaway-machine path rather than the
	// early return.
	nx := Preprocess(st, pn, Nx("Foo")).(Expr)
	_, isConst := nx.(*ConstExpr)
	require.False(t, isConst, "must not be a *ConstExpr; that path early-returns and would not exercise the recover")

	tv, err := tryEvalStatic(st, pn, pn, nx)

	require.NoError(t, err, "err must be nil when the expression resolves")
	require.NotNil(t, tv.V, "a resolved func must carry its value")
	require.NotNil(t, tv.GetUnboundFunc(), "expected a *FuncValue for Foo")
}

// TestTryEvalStatic_ErrSetOnFailure is the negative half: an expression that
// cannot be resolved statically must report a non-nil err.
func TestTryEvalStatic_ErrSetOnFailure(t *testing.T) {
	t.Parallel()
	st, pn := newTryEvalStaticTestPkg(t, "tryevalfail", "gno.land/r/test/tryevalfail")

	// An undeclared name cannot be resolved.
	_, err := tryEvalStatic(st, pn, pn, Nx("Undeclared"))

	require.Error(t, err, "err must be non-nil when the expression cannot resolve")
}
