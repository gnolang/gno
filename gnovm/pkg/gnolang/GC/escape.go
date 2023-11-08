package GC

import (
	// "fmt"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
)

type VarNode struct {
	Name 			string
	Escape			bool
	IsParameter		bool
	IsReturnValue 	bool
}

type VarGraph struct {
	Nodes map[string]*VarNode
	Edges map[string][]*VarNode
}

func NewVarGraph() *VarGraph {
	return &VarGraph{
		Nodes: make(map[string]*VarNode),
		Edges: make(map[string][]*VarNode),
	}
}

func (g *VarGraph) AddNode(name string) *VarNode {
	if _, exists := g.Nodes[name]; !exists {
        g.Nodes[name] = &VarNode{Name: name}
    }
    return g.Nodes[name]
}

func (g *VarGraph) AddEdge(fromName, toName string) {
    from := g.AddNode(fromName)
    to := g.AddNode(toName)

	// Initialize the edge list for 'from' node if it doesn't exist
    if _, exists := g.Edges[fromName]; !exists {
        g.Edges[fromName] = []*VarNode{}
    }

    // Add the 'from'/'to' node to the adjacency list of the 'from' node.
	g.Edges[fromName] = append(g.Edges[fromName], to)
	g.Edges[toName] = append(g.Edges[toName], from)
}

func (g *VarGraph) AnalyzeEscape() {
	visited := make(map[string]bool)
	// var path []string

	for name := range g.Nodes {
		visited[name] = false
	}

	var dfs func(node *VarNode)
    dfs = func(node *VarNode) {
		if node == nil {
			return
		}

		// Mark the current node as visited
        visited[node.Name] = true
		// path = append(path, node.Name)

        // If the current node escapes, mark all connected nodes as escaping
        if node.Escape {
			// fmt.Println("Escaping node:", node.Name)
            for _, edge := range g.Edges[node.Name] {
                edge.Escape = true
                if !visited[edge.Name] {
                    dfs(edge)
                }
            }
        }

		for _, edge := range g.Edges[node.Name] {
			if !visited[edge.Name] {
				dfs(edge)
			}
		}

		// Remove the current node from the path
		// path = path[:len(path)-1]
    }

	for _, node := range g.Nodes {
		if !visited[node.Name] {
			dfs(node)
		}
	}
}

func (g *VarGraph) GetEscapeVars() []string {
    var escapeVars []string
    for varName, varNode := range g.Nodes {
        if varNode.Escape {
            escapeVars = append(escapeVars, varName)
        }
    }
    return escapeVars
}

// TrackingEscapeVariables identifies variables in a function that should be allocated on the heap.
func TrackingEscapeVariables(f *ast.FuncDecl) []string {
	varGraph := NewVarGraph()

	// Add the function parameters and results to the graph.
	addFunctionParamsAndResultsToGraph(f.Type, varGraph)

	if f.Body == nil {
		return []string{}
	}

	// Inspect the AST of the function body to determine variable allocations.
	ast.Inspect(f.Body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch x := n.(type) {
		case *ast.GoStmt:
			// In a Go statement, arguments must be heap-allocated to be accessible in the new goroutine.
			for _, arg := range x.Call.Args {
				varName := getVariableName(arg)
				varGraph.AddNode(varName)
				varGraph.Nodes[varName].Escape = true
			}
		case *ast.CallExpr:
			for _, arg := range x.Args {
				varName := getVariableName(arg)
				if varName != "" && varName != "_" {
					// Check if the argument is a reference type or if it's address is taken.
					if isReference(arg) || isTakingAddress(arg) {
						varGraph.AddNode(varName)
						varGraph.Nodes[varName].Escape = true
					}
				}
			}
		case *ast.Ident:
			varName := x.Name
			if _, ok := varGraph.Nodes[varName]; !ok {
				varGraph.AddNode(varName)
			}
		case *ast.FuncLit:
			// Handle function literals, which can have their own set of heap-allocated variables.
			for _, param := range x.Type.Params.List {
                for _, paramName := range param.Names {
                    varGraph.AddNode(paramName.Name) // Add the parameter as a node.
					varGraph.Nodes[paramName.Name].Escape = true // mark as escaping to the heap.
                }
            }
			ast.Inspect(x.Body, func(n ast.Node) bool {
				if v, ok := n.(*ast.Ident); ok && (varGraph.Nodes[v.Name].Escape) {
					varGraph.AddNode(v.Name)
					varGraph.Nodes[v.Name].Escape = true
				}
				return true
			})
			// Parameters and return values of function literals may also be heap-allocated.
			checkParamsAndResultsForHeapAllocation(x.Type, varGraph)
		case *ast.AssignStmt:
			for _, expr := range x.Rhs {
				if isReference(expr) {
					// If the right-hand side is a reference, then the left-hand side should be on the heap.
					for _, lhs := range x.Lhs {
						if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
							varGraph.AddNode(ident.Name)
							varGraph.Nodes[ident.Name].Escape = true
						}
					}
				}
			}
		case *ast.ReturnStmt:
			// Returned values may escape to the heap.
			for _, res := range x.Results {
				varName := getVariableName(res)
				if varName != "" && varName != "_" {
					varNode := varGraph.AddNode(varName)
					if isReference(res) || isTakingAddress(res) {
						varNode.Escape = true
					} else {
						panic("unhandled return type")
					}
				}
			}
		case *ast.ValueSpec:
			// If a value specification is not a built-in type, it may be heap-allocated.
			if !isBuiltinType(x.Type) {
				for _, name := range x.Names {
					varGraph.AddNode(name.Name)
					varGraph.Nodes[name.Name].Escape = true
				}
			}
		}
		return true
	})

	// Parameters of the function itself may also be heap-allocated.
	checkParamsAndResultsForHeapAllocation(f.Type, varGraph)

	// Analyze the graph to determine which variables escape to the heap.
	varGraph.AnalyzeEscape()

	return varGraph.GetEscapeVars()
}

// checkParamsAndResultsForHeapAllocation checks the parameters and results of a function type for heap allocation.
func checkParamsAndResultsForHeapAllocation(ft *ast.FuncType, varGraph *VarGraph) {
	// Check parameters for references and built-in types.
	if ft.Params != nil {
		for _, param := range ft.Params.List {
			if isReference(param.Type) || isBuiltinType(param.Type) {
				for _, paramName := range param.Names {
					varNode := varGraph.AddNode(paramName.Name) // Add the parameter as a node.
					varNode.Escape = true // Parameters of functions may escape to the heap.
				}
			}
		}
	}
	// Check results for references.
	if ft.Results != nil {
		for _, result := range ft.Results.List {
			if isReference(result.Type) {
				for _, resultName := range result.Names {
					varNode := varGraph.AddNode(resultName.Name) // Add the result as a node.
					varNode.Escape = true // Returned values may escape to the heap.
				}
			}
		}
	}
}

// getVariableName extracts the name of a variable from an expression.
func getVariableName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return getVariableName(x.X)
	case *ast.UnaryExpr:
		return getVariableName(x.X)
	}
	return ""
}

// isReference determines if an expression is a reference type.
func isReference(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.StarExpr:
		// A star expression (*) or taking the address of (&) indicates a reference.
		return true
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			return true
		}
	case *ast.CallExpr:
		// A call expression may be a reference if it's a call to make or new.
		if id, ok := x.Fun.(*ast.Ident); ok {
			return id.Name == "new" || id.Name == "make"
		}
	}
	return false
}

// isBuiltinType checks if an expression is a built-in type.
func isBuiltinType(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.ArrayType, *ast.MapType, *ast.StructType, *ast.InterfaceType:
		return true
	case *ast.Ident:
		// TODO: This could be extended to include all built-in types.
		builtinTypes := map[string]bool{
			"string": true, "int": true, "float64": true,
			"bool": true, "byte": true, "rune": true,
			"int8": true, "int16": true, "int32": true,
			"int64": true, "uint": true, "uint8": true,
			"uint16": true, "uint32": true, "uint64": true,
		}
		if id, ok := expr.(*ast.Ident); ok {
			return builtinTypes[id.Name]
		}
	}
	return false
}

// isPointerDecl checks if the expression is a pointer declaration.
func isPointerDecl(expr ast.Expr) bool {
    _, ok := expr.(*ast.StarExpr)
    return ok
}

// isTakingAddress determines if an expression is taking the address of a variable.
func isTakingAddress(expr ast.Expr) bool {
    unaryExpr, ok := expr.(*ast.UnaryExpr)
    return ok && unaryExpr.Op == token.AND
}

func requiresEscapeAnalysis(expr ast.Expr, varGraph *VarGraph) bool {
	switch x := expr.(type) {
	case *ast.Ident:
		if isGlobal(x) {
			return true
		}
	case *ast.UnaryExpr:
		// If the expression takes the address of a variable, it may escape.
		if x.Op == token.AND {
			return true
		}
	case *ast.CallExpr:
		// If the expression is a call to a function that may cause escaping,
		// such as passing a variable to a goroutine, it may require escape analysis.
		errors.New(fmt.Sprintf("unhandled expression type: %T", expr))
	default:
		errors.New(fmt.Sprintf("unhandled expression type: %T", expr))
	}

	return false
}

func isGlobal(ident *ast.Ident) bool {
	if ident.Obj == nil {
		return false
	}

	if ident.Obj.Kind != ast.Var {
		return false
	}

	if isPackageLevel(ident.Obj) {
		return true
	}

	return false
}

// packageScopeEndPos would be determined during parsing and represents the end position
// of the package-level declarations.
//
// This should be set to the end position of the package-level scope.
var packageScopeEndPos token.Pos

// isPackageLevel determines if the object is declared at the package level.
func isPackageLevel(obj *ast.Object) bool {
	// XXX Placeholder for actual implementation.
	return obj.Decl.(ast.Node).Pos() <= packageScopeEndPos
}

func isParamOrReturnValue(ident *ast.Ident, varGraph *VarGraph) bool {
	for _, node := range varGraph.Nodes {
		if node.Name == ident.Name {
			if node.IsParameter || node.IsReturnValue {
				return true
			}
		}
	}

	return false
}

func addFunctionParamsAndResultsToGraph(f *ast.FuncType, varGraph *VarGraph) {
    for _, field := range f.Params.List {
        for _, paramName := range field.Names {
            node := varGraph.AddNode(paramName.Name)
            node.IsParameter = true
        }
    }

    if f.Results != nil {
        for _, field := range f.Results.List {
            for _, resultName := range field.Names {
                if resultName != nil {
                    node := varGraph.AddNode(resultName.Name)
                    node.IsReturnValue = true
                }
            }
        }
    }
}
