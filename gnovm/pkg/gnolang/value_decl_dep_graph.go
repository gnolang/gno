package gnolang

import (
	"fmt"
	"slices"
	"strings"
)

// sortValueDeps creates a new topologically sorted
// decl slice ready for processing in order
func sortValueDeps(decls Decls) (Decls, error) {
	graph := &graph{
		edges:    make(map[string][]string),
		vertices: make([]string, 0),
	}

	otherDecls := make(Decls, 0)

	for i := 0; i < len(decls); i++ {
		d := decls[i]
		vd, ok := d.(*ValueDecl)

		if !ok {
			otherDecls = append(otherDecls, d)
			continue
		}

		if isTuple(vd) {
			_, ok := vd.Values[0].(*CallExpr)
			if ok {
				graph.addVertex(vd.NameExprs.String())
				continue
			}
		}

		for j := 0; j < len(vd.NameExprs); j++ {
			graph.addVertex(string(vd.NameExprs[j].Name))
		}
	}

	for i := 0; i < len(decls); i++ {
		d := decls[i]
		vd, ok := d.(*ValueDecl)

		if !ok {
			continue
		}

		if isTuple(vd) {
			ce, ok := vd.Values[0].(*CallExpr)
			if ok {
				addDepFromExpr(graph, vd.NameExprs.String(), ce)
				continue
			}
		}

		for j := 0; j < len(vd.NameExprs); j++ {
			if len(vd.Values) > j {
				addDepFromExpr(graph, string(vd.NameExprs[j].Name), vd.Values[j])
			}
		}
	}

	sorted := make(Decls, 0)

	for _, node := range graph.topologicalSort() {
		var dd Decl

		for _, decl := range decls {
			vd, ok := decl.(*ValueDecl)

			if !ok {
				continue
			}

			if isCompoundNode(node) {
				dd = processCompound(node, vd, decl)
				break
			}

			for i, nameExpr := range vd.NameExprs {
				if string(nameExpr.Name) == node {
					if len(vd.Values) > i {
						dd = &ValueDecl{
							Attributes: vd.Attributes.Copy(),
							NameExprs:  []NameExpr{nameExpr},
							Type:       vd.Type,
							Values:     []Expr{vd.Values[i]},
							Const:      vd.Const,
						}
						break
					} else {
						dd = vd
						break
					}
				}
			}
		}

		if dd == nil {
			continue
		}

		sorted = append(sorted, dd)
	}

	slices.Reverse(sorted)

	otherDecls = append(otherDecls, sorted...)

	return otherDecls, nil
}

func addDepFromExpr(dg *graph, fromNode string, expr Expr) {
	switch e := expr.(type) {
	case *FuncLitExpr:
		for _, stmt := range e.Body {
			addDepFromExprStmt(dg, fromNode, stmt)
		}
	case *CallExpr:
		addDepFromExpr(dg, fromNode, e.Func)

		for _, arg := range e.Args {
			addDepFromExpr(dg, fromNode, arg)
		}
	case *NameExpr:
		if isUverseName(e.Name) {
			break
		}

		toNode := string(e.Name)
		dg.addEdge(fromNode, toNode)
	}
}

func addDepFromExprStmt(dg *graph, fromNode string, stmt Stmt) {
	switch e := stmt.(type) {
	case *ExprStmt:
		addDepFromExpr(dg, fromNode, e.X)
	case *IfStmt:
		addDepFromExprStmt(dg, fromNode, e.Init)
		addDepFromExpr(dg, fromNode, e.Cond)

		for _, stm := range e.Then.Body {
			addDepFromExprStmt(dg, fromNode, stm)
		}
		for _, stm := range e.Else.Body {
			addDepFromExprStmt(dg, fromNode, stm)
		}
	case *ReturnStmt:
		for _, stm := range e.Results {
			addDepFromExpr(dg, fromNode, stm)
		}
	case *AssignStmt:
		for _, stm := range e.Rhs {
			addDepFromExpr(dg, fromNode, stm)
		}
	case *SwitchStmt:
		addDepFromExpr(dg, fromNode, e.X)
		for _, clause := range e.Clauses {
			addDepFromExpr(dg, fromNode, clause.bodyStmt.Cond)
			for _, s := range clause.bodyStmt.Body {
				addDepFromExprStmt(dg, fromNode, s)
			}
		}
	case *ForStmt:
		addDepFromExpr(dg, fromNode, e.Cond)
		for _, s := range e.bodyStmt.Body {
			addDepFromExprStmt(dg, fromNode, s)
		}
	case *BlockStmt:
		for _, s := range e.Block.bodyStmt.Body {
			addDepFromExprStmt(dg, fromNode, s)
		}
	}
}

type graph struct {
	edges    map[string][]string
	vertices []string
}

func (g *graph) addEdge(u, v string) {
	g.edges[u] = append(g.edges[u], v)
}

func (g *graph) addVertex(v string) {
	g.vertices = append(g.vertices, v)
}

func (g *graph) topologicalSortUtil(v string, visited map[string]bool, stack *[]string) {
	visited[v] = true

	for _, u := range g.edges[v] {
		if !visited[u] {
			g.topologicalSortUtil(u, visited, stack)
		}
	}

	*stack = append([]string{v}, *stack...)
}

func (g *graph) topologicalSort() []string {
	stack := make([]string, 0)
	visited := make(map[string]bool)

	for _, v := range g.vertices {
		if !visited[v] {
			g.topologicalSortUtil(v, visited, &stack)
		}
	}

	return stack
}

func isTuple(vd *ValueDecl) bool {
	return len(vd.NameExprs) > len(vd.Values) && len(vd.Values) > 0
}

func isCompoundNode(node string) bool {
	return strings.Contains(node, ", ")
}

func processCompound(node string, vd *ValueDecl, decl Decl) Decl {
	names := strings.Split(node, ", ")

	if len(names) != len(vd.NameExprs) {
		panic("should not happen")
	}

	equal := true

	for i, name := range names {
		if vd.NameExprs[i].String() != name {
			equal = false
			break
		}
	}

	if !equal {
		panic(fmt.Sprintf("names: %+v != nameExprs: %+v\n", names, vd.NameExprs))
	}

	return decl
}
