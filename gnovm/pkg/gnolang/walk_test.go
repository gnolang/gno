package gnolang

import (
	"fmt"
	"go/parser"
	"go/token"
)

func ExampleInspect() {
	// src is the input for which we want to inspect the AST.
	src := `
package p
const c = 1.0
var X = f(3.14)*2 + c
`

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	gnon := Go2Gno(fset, f)

	// Inspect the AST and print all identifiers and literals.
	Inspect(gnon, func(n Node) bool {
		var s string
		switch x := n.(type) {
		case *BasicLitExpr:
			s = x.Value
		case *NameExpr:
			s = string(x.Name)
		case *FileNode:
			s = string(x.PkgName)
		}
		if s != "" {
			fmt.Printf("%v:\t%s\n", n.GetLine(), s)
		}
		return true
	}, nil)

	// Output:
	// 2:	p
	// 0:	c
	// 3:	1.0
	// 0:	X
	// 4:	f
	// 4:	3.14
	// 4:	2
	// 4:	c
}

func ExampleOutInspect() {
	// src is the input for which we want to inspect the AST.
	src := `
package p

var x = &Foo{
	left: &Foo{},
	right: &Foo{},
}
`

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	gnon := Go2Gno(fset, f)

	// Inspect the AST and set value paths and print them in the out closure.
	Inspect(gnon, nil, func(n Node) bool {
		switch x := n.(type) {
		case *CompositeLitExpr:
			fmt.Printf("%v:%v\n", x.Line, x.Type)
			for _, field := range x.Elts {
				fmt.Printf("%+v => %+v\n", field.Key, field.Value)
			}
			println()
		}
		return true
	})

	// Output:
	// 5:Foo<VPUverse(0)>
	// 6:Foo<VPUverse(0)>
	// 4:Foo<VPUverse(0)>
	// left<VPUverse(0)> => &(Foo<VPUverse(0)>{})
	// right<VPUverse(0)> => &(Foo<VPUverse(0)>{})
}
