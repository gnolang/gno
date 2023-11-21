package escape_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/gc/escape"
	"github.com/stretchr/testify/assert"
)

type escapeTest struct {
	testName     string
	code         string
	declaration  func(*ast.File) *ast.FuncDecl
	expectedVars []string
}

func TestTrackingEscapeVariables(t *testing.T) {
	tests := []escapeTest{
		{
			testName: "TestGoRoutineParameterEscape",
			code: `
	package main
	
	func foo() {
		var x int
		go bar(x)
	}
	
	func bar(y int) {
		println(y)
	}
	`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "foo" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"x"},
		},
		{
			testName: "TestReturnEscape",
			code: `
	package main

	func foo() *int {
		x := 0
		return &x
	}
	`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "foo" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"x"},
		},
		{
			testName: "TestAssignEscape",
			code: `
	package main
	
	func foo() {
		var x int
		var y *int
		y = &x
		bar(y)
	}
	
	func bar(z *int) {
		println(z)
	}`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "foo" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"y"},
		},
		{
			testName: "TestCycleFunctionCallEscape",
			code: `
	package main

	func foo() {
		x := 10
		y := &x
		z := &y
		bar(z)
	}
	
	func bar(z **int) {
		foo()
	}`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "foo" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"y", "z"},
		},
		{
			testName: "TestAssignEscape3",
			code: `
	package main

	import "fmt"

	func main() {
		a := 10
		b := &a
		fmt.Println(*b)
	}`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "main" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"b"},
		},
		{
			testName: "TestGlobalEscape",
			code: `
		package main

		import "fmt"

		var global *int

		func escape() *int {
			a := 10
			return &a
		}

		func main() {
			global = escape()
			fmt.Println(*global)
		}`,
			declaration: func(f *ast.File) *ast.FuncDecl {
				for _, decl := range f.Decls {
					if fn, isFunc := decl.(*ast.FuncDecl); isFunc && fn.Name.Name == "main" {
						return fn
					}
				}
				return nil
			},
			expectedVars: []string{"global"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "", tt.code, 0)
			if err != nil {
				t.Fatalf("ParseFile() error = %v", err)
			}

			fn := tt.declaration(f)
			if fn == nil {
				t.Fatal("declaration() did not return a *ast.FuncDecl")
			}

			// Perform escape analysis.
			escapingVars := escape.TrackingEscapeVariables(fn)

			// Assert that the escaping variables match the expected ones.
			assert.ElementsMatch(t, tt.expectedVars, escapingVars)
		})
	}
}

func TestAddNode(t *testing.T) {
	graph := escape.NewVarGraph()

	// Test adding a new node.
	nodeName := "a"
	node := graph.AddNode(nodeName)
	assert.NotNil(t, node, "The node should not be nil.")
	assert.Equal(t, nodeName, node.Name, "The node name should be set correctly.")
	assert.Contains(t, graph.Nodes, nodeName, "The graph should contain the new node.")

	// Test adding a duplicate node.
	dupNode := graph.AddNode(nodeName)
	assert.Equal(t, node, dupNode, "Adding a node with the same name should return the existing node.")
	assert.Len(t, graph.Nodes, 1, "The graph should not contain duplicate nodes.")
}

func TestAnalyzeEscape(t *testing.T) {
	varGraph := escape.NewVarGraph()

	varGraph.AddNode("a").Escape = true
	varGraph.AddNode("b").Escape = true
	varGraph.AddNode("c")

	varGraph.AddEdge("a", "c")

	varGraph.AnalyzeEscape()

	assert.True(t, varGraph.Nodes["a"].Escape, "Node 'a' should escape")
	assert.True(t, varGraph.Nodes["b"].Escape, "Node 'b' should escape")

	assert.True(t, varGraph.Nodes["c"].Escape, "Node 'c' should escape because it's connected to 'a'")
}

func TestAddEdge(t *testing.T) {
	graph := escape.NewVarGraph()

	// Test adding an edge between two nodes.
	fromNode := "a"
	toNode := "b"
	graph.AddNode(fromNode)
	graph.AddNode(toNode)
	graph.AddEdge(fromNode, toNode)

	assert.Contains(t, graph.Edges, fromNode, "The graph should contain the 'from' node as a key in the edges.")
	assert.Contains(t, graph.Edges[fromNode], graph.Nodes[toNode], "The 'to' node should be in the edges list of the 'from' node.")

	// Test adding an edge with a new 'to' node.
	newToNode := "c"
	graph.AddEdge(fromNode, newToNode)
	assert.Contains(t, graph.Nodes, newToNode, "The graph should automatically add a new 'to' node.")
	assert.Contains(t, graph.Edges[fromNode], graph.Nodes[newToNode], "The new 'to' node should be in the edges list of the 'from' node.")
}