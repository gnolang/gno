package gnolang

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
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
	n, err := ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	fmt.Printf("CODE:\n%s\n\n", gocode)
	fmt.Printf("AST:\n%#v\n\n", n)
	fmt.Printf("AST.String():\n%s\n", n.String())
}

func TestTypeCheckMemPackage_MultiError(t *testing.T) {
	const src = `package main
func main() {
	_, _ = 11
	return 88, 88
}`
	err := TypeCheckMemPackage(&std.MemPackage{
		Name: "main",
		Path: "gno.land/p/demo/x",
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: src,
			},
		},
	}, nil)
	errs := multierr.Errors(err)
	if assert.Len(t, errs, 2, "should contain two errors") {
		assert.ErrorContains(t, errs[0], "assignment mismatch")
		assert.ErrorContains(t, errs[1], "too many return values")
	}
}
