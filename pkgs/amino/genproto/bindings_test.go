package genproto

import (
	"bytes"
	"testing"

	"go/printer"
	"go/token"

	"github.com/gnolang/gno/pkgs/amino/tests"
	"github.com/jaekwon/testify/assert"
)

func TestGenerateProtoBindings(t *testing.T) {
	file, err := GenerateProtoBindingsForTypes(tests.Package, tests.Package.ReflectTypes()...)
	assert.NoError(t, err)
	t.Logf("%v", file)

	// Print the function body into buffer buf.
	// The file set is provided to the printer so that it knows
	// about the original source formatting and can add additional
	// line breaks where they were present in the source.
	var buf bytes.Buffer
	var fset = token.NewFileSet()
	printer.Fprint(&buf, fset, file)
	t.Log(buf.String())
}
