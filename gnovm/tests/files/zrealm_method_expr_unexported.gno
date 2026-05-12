// PKGPATH: gno.land/r/test/methodexprunexported
package methodexprunexported

// Regression: the method-expression resolver must enforce export
// visibility, matching the bound-method path's check (types.go:1685).
// Pre-fix, `(*tests.TestRealmObject2).clearField` resolved successfully
// from outside the tests package, yielding a callable *FuncValue. A
// confused-deputy callback in tests could then invoke clearField on
// tests' own state. Post-fix, preprocess.go's *TypeType branch panics
// with "cannot access unexported method ... from ...".

import (
	"gno.land/p/demo/tests"
)

func main() {
	_ = (*tests.TestRealmObject2).clearField
}

// Error:
// gno.land/r/test/methodexprunexported/zrealm_method_expr_unexported.gno:17:6-42: cannot access unexported method gno.land/p/demo/tests.clearField from gno.land/r/test/methodexprunexported

// TypeCheckError:
// gno.land/r/test/methodexprunexported/zrealm_method_expr_unexported.gno:17:32: (*tests.TestRealmObject2).clearField undefined (cannot refer to unexported method clearField)
