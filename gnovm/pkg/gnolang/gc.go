package gnolang

import (
	"go/ast"
	"go/token"
)

// EscapeAnalysis tracks whether values
// need to be heap allocated
// here are the 3 rules we use
// 1. if a reference is assigned/passed as an arg
// 2. if a reference is returned
// 3. if a closure is using variables from the outer scope
// the escape analysis are done on function basis to avoid
// analysing complicated program flows
func EscapeAnalysis(f *ast.FuncDecl) []string {
	var heapVars []string
	ast.Inspect(f.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncLit:
			for _, v := range x.Type.Params.List {
				if isReference(v.Type) {
					heapVars = append(heapVars, v.Names[0].Name)
				}
			}
			if x.Type.Results != nil {
				for _, v := range x.Type.Results.List {
					if isReference(v.Type) {
						heapVars = append(heapVars, v.Names[0].Name)
					}
				}
			}
		case *ast.AssignStmt:
			//todo iterate over lhs and rhs and
			// add to lhs to heap vars if rhs is &T or T that is root
			for _, expr := range x.Rhs {
				if isReference(expr) {
					for _, v := range x.Lhs {
						heapVars = append(heapVars, getVarName(v))
					}
				}
			}
			for _, rhsExpr := range x.Rhs {
				if isReference(rhsExpr) {
					for _, lhsExpr := range x.Lhs {
						heapVars = append(heapVars, getVarName(lhsExpr))
						// If the LHS expression is a variable that holds a copy of the value
						// of another variable that references a heap-allocated value,
						// add the original variable to the heapVars slice as well.
						if ident, ok := lhsExpr.(*ast.Ident); ok && ident.Obj != nil && ident.Obj.Kind == ast.Var {
							if isReference(ident.Obj.Decl.(*ast.AssignStmt).Rhs[0]) {
								heapVars = append(heapVars, ident.Name)
							}
						}
					}
				}
			}
		case *ast.ReturnStmt:
			for _, result := range x.Results {
				if isReference(result) {
					heapVars = append(heapVars, getVarName(result))
				}
			}
		}
		return true
	})
	return heapVars
}

func isReference(expr ast.Expr) bool {
	switch ex := expr.(type) {
	case *ast.StarExpr:
		return true
	case *ast.UnaryExpr:
		if ex.Op == token.AND {
			//if ident, ok := ex.X.(*ast.Ident); ok {
			//	heapvars = append(heapvars, ident.String())
			//}
			return true
		}
	}
	return false
}

func getVarName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return getVarName(x.X)
	}
	return ""
}

func checkEscaped(ident string, escapedList []string) bool {
	for _, s := range escapedList {
		if s == ident {
			return true
		}
	}
	return false
}

type GC struct {
	objs  []*GCObj
	roots []*GCObj
}

type GCObj struct {
	value  interface{}
	marked bool
	ref    *GCObj
	path   string
}

func NewGC() *GC {
	return &GC{}
}

// AddObject use for escaped objects
func (gc *GC) AddObject(obj *GCObj) {
	gc.objs = append(gc.objs, obj)
}

func (gc *GC) RemoveRoot(path string) {
	for i, o := range gc.roots {
		if o.path == path {
			gc.objs = append(gc.objs[:i], gc.objs[i+1:]...)
			break
		}
	}
}

// AddRoot adds roots that won't be cleaned up by the GC
// use for stack variables/globals
func (gc *GC) AddRoot(root *GCObj) {
	gc.roots = append(gc.roots, root)
}

func (gc *GC) Collect() {
	// Mark phase
	for _, root := range gc.roots {
		gc.markObject(root)
	}

	// Sweep phase
	newObjs := make([]*GCObj, 0, len(gc.objs))
	for _, obj := range gc.objs {
		if !obj.marked {
			continue
		}
		obj.marked = false
		newObjs = append(newObjs, obj)
	}
	gc.objs = newObjs
}

func (gc *GC) markObject(obj *GCObj) {
	if obj.marked {
		return
	}
	obj.marked = true
	gc.markObject(obj.ref)
}

// use this only in tests
// because if you hold on to a reference of the GC object
// the Go GC cannot reclaim this memory
// only get GC object references through roots
func (gc *GC) getObjByPath(path string) *GCObj {
	for _, obj := range gc.objs {
		if obj.path == path {
			return obj
		}
	}
	return nil
}

func (gc *GC) getRootByPath(path string) *GCObj {
	for _, obj := range gc.roots {
		if obj.path == path {
			return obj
		}
	}
	return nil
}
