//go:build debugAssert

package gnolang

import "testing"

func TestDeclaredTypeTypeIDCacheHitChecksIdentity(t *testing.T) {
	dt := &DeclaredType{PkgPath: "gno.land/r/demo", Name: "Counter"}
	_ = dt.TypeID()
	dt.Name = "Changed"

	defer func() {
		if recover() == nil {
			t.Fatal("cached TypeID accepted changed identity")
		}
	}()
	_ = dt.TypeID()
}
