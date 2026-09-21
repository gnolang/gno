package gnolang

import (
	"testing"

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

func TestParseTilde(t *testing.T) {
	t.Parallel()

	var m *Machine
	_, err := m.ParseFile("tilde.gno", "package A\nvar x = ~0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tilde operator is not permitted")
}
