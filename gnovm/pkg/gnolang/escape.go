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
	heapVars := make(map[string]bool)
	vars := make(map[string]bool)
	ast.Inspect(f.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GoStmt:
			for _, arg := range x.Call.Args {
				if !heapVars[getVarName(arg)] {
					continue
				}
				if !vars[getVarName(arg)] {
					continue
				}

				heapVars[getVarName(arg)] = true
			}
		case *ast.CallExpr:
			for _, arg := range x.Args {
				if !heapVars[getVarName(arg)] &&
					!vars[getVarName(arg)] &&
					!isReference(arg) {
					continue
				}

				heapVars[getVarName(arg)] = true
			}
		case *ast.Ident:
			vars[x.Name] = true
		case *ast.FuncLit:
			// TODO: skip walking the body in the outer scope
			ast.Inspect(x.Body, func(n ast.Node) bool {
				if v, ok := n.(*ast.Ident); ok {
					if heapVars[v.Name] || vars[v.Name] {
						heapVars[v.Name] = true
					}
				}

				return true
			})

			for _, v := range x.Type.Params.List {
				if !isReference(v.Type) && !isSpecialType(v.Type) {
					continue
				}

				for _, m := range v.Names {
					heapVars[m.Name] = true
				}
			}

			if x.Type.Results != nil {
				for _, v := range x.Type.Results.List {
					if !isReference(v.Type) {
						continue
					}

					for _, v := range v.Names {
						heapVars[v.Name] = true
					}
				}
			}
		case *ast.AssignStmt:
			for i, expr := range x.Rhs {
				ln := getVarName(x.Lhs[i])
				rn := getVarName(expr)

				if isReference(expr) {
					if ln != "" && ln != "_" {
						heapVars[ln] = true
					}
					if rn != "" && rn != "_" {
						heapVars[rn] = true
					}

				} else if heapVars[rn] && ln != "" && ln != "_" {
					heapVars[ln] = true
				}
			}
		case *ast.ReturnStmt:
			for _, result := range x.Results {
				if !isReference(result) {
					continue
				}

				heapVars[getVarName(result)] = true
			}
		case *ast.ValueSpec:
			if !isSpecialType(x.Type) {
				return true
			}

			for _, n := range x.Names {
				heapVars[getVarName(n)] = true
			}
		}

		return true
	})

	for _, v := range f.Type.Params.List {
		if !isSpecialType(v.Type) {
			continue
		}

		for _, m := range v.Names {
			heapVars[m.Name] = true
		}
	}

	var out []string
	for k, _ := range heapVars {
		out = append(out, k)
	}

	return out
}

func isSpecialType(expr ast.Expr) bool {
	switch ex := expr.(type) {
	case *ast.ArrayType:
		return true
	case *ast.MapType:
		return true
	case *ast.InterfaceType:
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

	case *ast.CallExpr:
		// special case for new(blah)
		if id, ok := ex.Fun.(*ast.Ident); ok {
			if id.Name != "new" {
				break
			}

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

func getIdent(expr ast.Expr) *ast.Ident {
	switch x := expr.(type) {
	case *ast.Ident:
		return x
		//case *ast.StarExpr:
		//	return getIdent(x.X)
		//case *ast.UnaryExpr:
		//	return getIdent(x.X)
	}
	return nil
}

func checkEscaped(ident string, escapedList []string) bool {
	for _, s := range escapedList {
		if s == ident {
			return true
		}
	}
	return false
}
