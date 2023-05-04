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
	var heapVars, vars []string
	ast.Inspect(f.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GoStmt:
			for _, arg := range x.Call.Args {
				if !checkEscaped(getVarName(arg), heapVars) {
					continue
				}
				if !checkEscaped(getVarName(arg), vars) {
					continue
				}

				heapVars = append(heapVars, getVarName(arg))
			}
		case *ast.CallExpr:
			for _, arg := range x.Args {
				if !checkEscaped(getVarName(arg), heapVars) &&
					!checkEscaped(getVarName(arg), vars) &&
					!isReference(arg) {
					continue
				}

				heapVars = append(heapVars, getVarName(arg))
			}
		case *ast.Ident:
			vars = append(vars, x.String())
		case *ast.FuncLit:
			// TODO: skip walking the body in the outer scope
			ast.Inspect(x.Body, func(n ast.Node) bool {
				if v, ok := n.(*ast.Ident); ok {
					if checkEscaped(v.String(), heapVars) || checkEscaped(v.String(), vars) {
						heapVars = append(heapVars, v.String())
					}
				}

				return true
			})

			for _, v := range x.Type.Params.List {
				if isReference(v.Type) || isSpecialType(v.Type) {
					for _, m := range v.Names {
						heapVars = append(heapVars, m.Name)
					}
				}
			}

			if x.Type.Results != nil {
				for _, v := range x.Type.Results.List {
					if !isReference(v.Type) {
						continue
					}

					heapVars = append(heapVars, v.Names[0].Name)
				}
			}
		case *ast.AssignStmt:
			for i, expr := range x.Rhs {
				ln := getVarName(x.Lhs[i])
				rn := getVarName(expr)

				if isReference(expr) {
					if ln != "" && ln != "_" {
						heapVars = append(heapVars, ln)
					}
					heapVars = append(heapVars, rn)
				} else if checkEscaped(rn, heapVars) && ln != "" && ln != "_" {
					heapVars = append(heapVars, ln)
				}
			}
		case *ast.ReturnStmt:
			for _, result := range x.Results {
				if !isReference(result) {
					continue
				}

				heapVars = append(heapVars, getVarName(result))
			}
		}
		return true
	})
	return heapVars
}

func isSpecialType(expr ast.Expr) bool {
	switch ex := expr.(type) {
	case *ast.ArrayType:
		return true
	case *ast.MapType:
		return true
	case *ast.Ident:
		if ex.Name == "string" {
			return true
		}
	}

	return false
}

func isReference(expr ast.Expr) bool {
	switch ex := expr.(type) {
	case *ast.StarExpr:
		return true
	case *ast.UnaryExpr:
		if ex.Op == token.AND {
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
	case *ast.UnaryExpr:
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
		if o.path != path {
			continue
		}

		gc.roots = append(gc.roots[:i], gc.roots[i+1:]...)

		break
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

	if obj.ref == nil {
		return
	}
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
