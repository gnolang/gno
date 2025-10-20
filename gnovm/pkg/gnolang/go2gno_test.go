package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
