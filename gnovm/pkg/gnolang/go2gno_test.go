package gnolang

import (
	"fmt"
	"testing"

	"github.com/jaekwon/testify/assert"
)

func TestParseForLoop(t *testing.T) {
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

func TestParseSetRoots(t *testing.T) {
	gocode := `package main

func f() {
    i := 1
    func() {
        _ = i
    }()
	b := 4
	c := 5
	foo(&b)
}

func ff() {
    i := 1
	b := 4
	c := 5
	foo(&b)
}

func main() {
	f()
}
`
	n, err := ParseFile("main.go", gocode)
	assert.NoError(t, err, "ParseFile error")
	assert.NotNil(t, n, "ParseFile error")

	for _, decl := range n.Decls {
		fn := decl.(*FuncDecl)
		if fn.Name == "f" {
			i := fn.Body[0].(*AssignStmt).Lhs[0].(*NameExpr)
			b := fn.Body[2].(*AssignStmt).Lhs[0].(*NameExpr)
			c := fn.Body[3].(*AssignStmt).Lhs[0].(*NameExpr)

			assert.True(t, i.IsRoot)
			assert.True(t, b.IsRoot)
			assert.False(t, c.IsRoot)
		} else if fn.Name == "ff" {
			i := fn.Body[0].(*AssignStmt).Lhs[0].(*NameExpr)
			b := fn.Body[1].(*AssignStmt).Lhs[0].(*NameExpr)
			c := fn.Body[2].(*AssignStmt).Lhs[0].(*NameExpr)

			assert.False(t, i.IsRoot)
			assert.True(t, b.IsRoot)
			assert.False(t, c.IsRoot)
		}
	}
}
