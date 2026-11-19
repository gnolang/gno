package gnolang

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// newBuiltin stands for a builtin added after the shadowing rule was relaxed:
// in the uverse, not in protectedUverseNames.
const newBuiltin Name = "newbuiltin"

// Defined in init, before any test runs, so no test reads the uverse while it
// changes; the uverse package value is rebuilt from its node, as
// makeUverseNode does, so the new builtin has a value too.
func init() {
	UverseNode().Define2(true, newBuiltin, gTypeType, asValue(IntType), NameSource{})
	uverseValue = uverseNode.NewPackage(nil)
}

// runShadowTest runs the files of package gno.land/p/test/shadow, after saving
// a package gno.land/p/test/dep for them to import, and returns what it
// printed.
func runShadowTest(t *testing.T, body string) string {
	t.Helper()
	db := memdb.NewMemDB()
	store := NewStore(nil, dbadapter.StoreConstructor(db, stypes.StoreOptions{}), iavl.StoreConstructor(db, stypes.StoreOptions{}))
	var buf bytes.Buffer
	m := NewMachineWithOptions(MachineOptions{Output: &buf, Store: store})
	defer m.Release()
	m.RunMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "dep",
		Path:  "gno.land/p/test/dep",
		Files: []*std.MemFile{{Name: "dep.gno", Body: "package dep\n\nfunc Hello() string { return \"hello\" }\n"}},
	}, true)
	m2 := NewMachineWithOptions(MachineOptions{Output: &buf, Store: store})
	defer m2.Release()
	m2.RunMemPackage(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "shadow",
		Path:  "gno.land/p/test/shadow",
		Files: []*std.MemFile{{Name: "shadow.gno", Body: body}},
	}, false)
	m2.RunMain()
	return buf.String()
}

// TestShadowNewBuiltin pins that a builtin added after the shadowing rule was
// relaxed can be shadowed by every kind of declaration, as a package already
// using its name must keep compiling once it exists, while the builtins that
// existed before cannot.
func TestShadowNewBuiltin(t *testing.T) {
	t.Run("import alias", func(t *testing.T) {
		require.Equal(t, "hello\n", runShadowTest(t, `package shadow

import newbuiltin "gno.land/p/test/dep"

func main() { println(newbuiltin.Hello()) }
`))
	})
	t.Run("package-level declaration", func(t *testing.T) {
		require.Equal(t, "shadowed\n", runShadowTest(t, `package shadow

var newbuiltin = "shadowed"

func main() { println(newbuiltin) }
`))
	})
	t.Run("local definition and parameter", func(t *testing.T) {
		require.Equal(t, "3\n4\n", runShadowTest(t, `package shadow

func f(newbuiltin int) int { return newbuiltin + 1 }

func main() {
	newbuiltin := 3
	println(newbuiltin)
	println(f(newbuiltin))
}
`))
	})
	t.Run("still usable where not shadowed", func(t *testing.T) {
		require.Equal(t, "7\n", runShadowTest(t, `package shadow

func main() {
	var x newbuiltin = 7
	println(x)
}
`))
	})
	t.Run("protected builtins stay protected", func(t *testing.T) {
		defer func() {
			require.Contains(t, fmt.Sprint(recover()), "builtin identifiers cannot be shadowed: len")
		}()
		runShadowTest(t, `package shadow

var len = 1

func main() {}
`)
		t.Fatal("shadowing len did not panic")
	})
}

// TestProtectedUverseNames pins that every protected name is a builtin, so
// the list cannot silently drift from the uverse.
func TestProtectedUverseNames(t *testing.T) {
	for n := range protectedUverseNames {
		require.True(t, isUverseName(n), "%s is not a builtin", n)
	}
}
