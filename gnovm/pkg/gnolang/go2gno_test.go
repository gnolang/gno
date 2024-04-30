package gnolang

import (
	"fmt"
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
	n, err := ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")
	fmt.Printf("CODE:\n%s\n\n", gocode)
	fmt.Printf("AST:\n%#v\n\n", n)
	fmt.Printf("AST.String():\n%s\n", n.String())
}
