package genproto

import (
	"bytes"
	"go/printer"
	"go/token"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino/tests"
	"github.com/stretchr/testify/assert"
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
