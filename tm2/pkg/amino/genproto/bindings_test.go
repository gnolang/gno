package genproto

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateProtoBindings(t *testing.T) {
	t.Parallel()

	file, err := GenerateProtoBindingsForTypes(tests.Package, tests.Package.ReflectTypes()...)
	assert.NoError(t, err)
	t.Logf("%v", file)

	// Print the function body into buffer buf.
	// The file set is provided to the printer so that it knows
	// about the original source formatting and can add additional
	// line breaks where they were present in the source.
	var buf bytes.Buffer
	fset := token.NewFileSet()
	printer.Fprint(&buf, fset, file)
	t.Log(buf.String())
}

// TestGenerateProtoBindings_SkipZero: for an amino:"skip_zero" field,
// ToPBMessage leaves the pb field unset whenever the Go value is zero, so
// proto3 leaves it out as amino does, FromPBMessage reads an unset pb field
// as the zero Go value without converting it, and Is*ReprEmpty counts a
// zero one as empty. Untagged fields convert as before.
func TestGenerateProtoBindings_SkipZero(t *testing.T) {
	t.Parallel()

	file, err := GenerateProtoBindingsForTypes(tests.Package, tests.Package.ReflectTypes()...)
	require.NoError(t, err)
	var buf bytes.Buffer
	require.NoError(t, printer.Fprint(&buf, token.NewFileSet(), file))
	src := buf.String()
	method := func(signature string) string {
		start := strings.Index(src, signature)
		require.GreaterOrEqual(t, start, 0, "%s not generated:\n%s", signature, src)
		body := src[start:]
		if end := strings.Index(body, "\n}\n"); end >= 0 {
			body = body[:end]
		}
		return body
	}
	toPB := method("func (goo FuzzSkipZero) ToPBMessage(")
	fromPB := method("func (goo *FuzzSkipZero) FromPBMessage(")

	for _, f := range []string{"Addr", "Bytes", "Ints", "Inners", "Inner", "List", "Score", "Fixed", "Name", "Nums"} {
		assert.Contains(t, toPB, fmt.Sprintf("if !reflect.ValueOf(&goo.%s).Elem().IsZero() {", f))
		assert.Contains(t, fromPB, fmt.Sprintf("if !reflect.ValueOf(&pbo.%s).Elem().IsZero() {", f))
		assert.Contains(t, fromPB, fmt.Sprintf("reflect.ValueOf(&(*goo).%s).Elem().SetZero()", f))
	}
	for _, f := range []string{"PlainAddr", "PlainInner"} {
		assert.NotContains(t, toPB, fmt.Sprintf("reflect.ValueOf(&goo.%s)", f))
		assert.NotContains(t, fromPB, fmt.Sprintf("reflect.ValueOf(&pbo.%s)", f))
	}

	// SkipZeroOnly's one field is tagged, so at zero amino leaves out the
	// whole struct, and a field of its type gets no pb message.
	isEmpty := method("func IsSkipZeroOnlyReprEmpty(")
	assert.Contains(t, isEmpty, "if !reflect.ValueOf(&goor.Addr).Elem().IsZero() {")
}
