package doc

import (
	"fmt"
	"go/ast"
	"strings"
)

func generateFuncSignature(fn *ast.FuncDecl) string {
	if fn == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("func ")

	if fn.Recv != nil {
		var receiverNames []string
		for _, field := range fn.Recv.List {
			var fieldName string
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}
			receiverNames = append(receiverNames, fmt.Sprintf("%s %s", fieldName, typeString(field.Type)))
		}
		if len(receiverNames) > 0 {
			b.WriteString(fmt.Sprintf("(%s) ", strings.Join(receiverNames, ", ")))
		}
	}

	fmt.Fprintf(&b, "%s(", fn.Name.Name)

	var params []string
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			paramType := typeString(param.Type)
			if len(param.Names) == 0 {
				params = append(params, paramType)
			} else {
				paramNames := make([]string, len(param.Names))
				for i, name := range param.Names {
					paramNames[i] = name.Name
				}
				params = append(params, fmt.Sprintf("%s %s", strings.Join(paramNames, ", "), paramType))
			}
		}
	}

	fmt.Fprintf(&b, "%s)", strings.Join(params, ", "))

	results := []string{}
	if fn.Type.Results != nil {
		hasNamedParams := false
		for _, result := range fn.Type.Results.List {
			if len(result.Names) == 0 {
				results = append(results, typeString(result.Type))
			} else {
				hasNamedParams = true
				var resultNames []string
				for _, id := range result.Names {
					resultNames = append(resultNames, id.Name)
				}
				results = append(results, fmt.Sprintf("%s %s", strings.Join(resultNames, ", "), typeString(result.Type)))
			}
		}

		if len(results) > 0 {
			b.WriteString(" ")
			returnType := strings.Join(results, ", ")

			if hasNamedParams || len(results) > 1 {
				returnType = fmt.Sprintf("(%s)", returnType)
			}

			b.WriteString(returnType)
		}
	}

	return b.String()
}

// This code is inspired by the code at https://cs.opensource.google/go/go/+/refs/tags/go1.20.1:src/go/doc/reader.go;drc=40ed3591829f67e7a116180aec543dd15bfcf5f9;bpv=1;bpt=1;l=124
func typeString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return typeString(t.X)
	case *ast.IndexListExpr:
		return typeString(t.X)
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return fmt.Sprintf("%s.%s", typeString(t.X), t.Sel.Name)
		}
	case *ast.ParenExpr:
		return typeString(t.X)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", typeString(t.X))
	case *ast.BasicLit:
		return t.Value
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", typeString(t.Elt))
	case *ast.FuncType:
		var params []string
		if t.Params != nil {
			for _, field := range t.Params.List {
				paramType := typeString(field.Type)
				if len(field.Names) > 0 {
					for _, name := range field.Names {
						params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
					}
				} else {
					params = append(params, paramType)
				}
			}
		}
		var results []string
		if t.Results != nil {
			for _, field := range t.Results.List {
				resultType := typeString(field.Type)
				if len(field.Names) > 0 {
					for _, name := range field.Names {
						results = append(results, fmt.Sprintf("%s %s", name.Name, resultType))
					}
				} else {
					results = append(results, resultType)
				}
			}
		}

		return strings.TrimSpace(fmt.Sprintf("func(%s) %s", strings.Join(params, ", "), strings.Join(results, ", ")))
	case *ast.StructType:
		var fields []string
		for _, field := range t.Fields.List {
			fieldType := typeString(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					fields = append(fields, fmt.Sprintf("%s %s", name.Name, fieldType))
				}
			} else {
				fields = append(fields, fieldType)
			}
		}
		return fmt.Sprintf("struct{%s}", strings.Join(fields, "; "))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeString(t.Key), typeString(t.Value))
	case *ast.ChanType:
		chanDir := "chan"
		if t.Dir == ast.SEND {
			chanDir = "chan<-"
		} else if t.Dir == ast.RECV {
			chanDir = "<-chan"
		}
		return fmt.Sprintf("%s %s", chanDir, typeString(t.Value))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", typeString(t.Len), typeString(t.Elt))
	case *ast.SliceExpr:
		return fmt.Sprintf("[]%s", typeString(t.X))
	}
	return ""
}

func isFuncExported(fn *ast.FuncDecl) bool {
	if !fn.Name.IsExported() {
		return false
	}

	if fn.Recv == nil {
		return true
	}

	for _, recv := range fn.Recv.List {
		if ast.IsExported(removePointer(typeString(recv.Type))) {
			return true
		}
	}

	return false
}

func removePointer(name string) string {
	return strings.TrimPrefix(name, "*")
}
