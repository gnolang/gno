package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseForLoop(t *testing.T) {
	t.Parallel()

	gocode := `package main
func main(){
	for i:=0; i<10; i++ {
		if i == -1 {
			return
		}
	}
}`
	var m *Machine
	n, err := m.ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	t.Logf("CODE:\n%s\n\n", gocode)
	t.Logf("AST:\n%#v\n\n", n)
	t.Logf("AST.String():\n%s\n", n.String())
}

// TestCommentsDoNotAffectParsingGas verifies that comment tokens are excluded
// from gas metering during source code parsing. The parser callback skips
// token.COMMENT so that documentation does not penalize runtime gas costs.
func TestCommentsDoNotAffectParsingGas(t *testing.T) {
	bodyNoComments := `package foo

func Add(a, b int) int {
	return a + b
}
`
	bodyWithComments := `package foo

// Add adds two integers together and returns the result.
// This is a very detailed comment explaining the function.
// It spans multiple lines to simulate real-world documentation.
func Add(a, b int) int {
	// perform addition
	return a + b // returns the sum
}
`

	parseWithGas := func(t *testing.T, body string) int64 {
		t.Helper()

		gasMeter := stypes.NewGasMeter(10_000_000)

		db := memdb.NewMemDB()
		baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
		iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), stypes.StoreOptions{})
		st := NewStore(nil, baseStore, iavlStore)

		m := NewMachineWithOptions(MachineOptions{
			PkgPath:  "gno.land/p/test/foo",
			Store:    st,
			Output:   io.Discard,
			GasMeter: gasMeter,
		})
		defer m.Release()

		fn, err := m.ParseFile("foo.gno", body)
		require.NoError(t, err)
		require.NotNil(t, fn)

		return gasMeter.GasConsumed()
	}

	gasNoComments := parseWithGas(t, bodyNoComments)
	gasWithComments := parseWithGas(t, bodyWithComments)

	t.Logf("parsing gas without comments: %d", gasNoComments)
	t.Logf("parsing gas with comments:    %d", gasWithComments)

	assert.Equal(t, gasNoComments, gasWithComments,
		"comments should not affect parsing gas")
}
