package doc

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
)

type Package struct {
	ImportPath string
	Name       string
	Doc        string
	Filenames  []string
	Funcs      []*Func
	Methods    []*Func
	Vars       []*Value
	Consts     []*Value
	Types      []*Type
}

func (p *Package) filterTypeFuncs(typeName string) (funcs []*Func, methods []*Func) {
	var remainingFuncs []*Func

	for _, fn := range p.Funcs {
		if fn.Recv == nil {
			matched := false
			for _, r := range fn.Returns {
				if r.Type == typeName || r.Type == "*"+typeName {
					funcs = append(funcs, fn)
					matched = true
					break
				}
			}

			if !matched {
				remainingFuncs = append(remainingFuncs, fn)
			}
			continue
		}
		for _, n := range fn.Recv {
			if n == typeName || n == "*"+typeName {
				methods = append(methods, fn)
				break
			} else {
				remainingFuncs = append(remainingFuncs, fn)
			}
		}
	}

	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].Name < funcs[j].Name
	})

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	p.Funcs = remainingFuncs

	return
}

func (p *Package) filterTypeValues(typeName string) (vars []*Value, consts []*Value) {
	var (
		remainingVars   []*Value
		remainingConsts []*Value
	)

	for _, v := range p.Vars {
		var matched bool
		for _, item := range v.Items {
			if item.Type == typeName || item.Type == "*"+typeName {
				vars = append(vars, v)
				matched = true
				break
			}
		}
		if !matched {
			remainingVars = append(remainingVars, v)
		}
	}

	for _, c := range p.Consts {
		var matched bool
		for _, item := range c.Items {
			if item.Type == typeName {
				consts = append(consts, c)
				matched = true
				break
			}
		}
		if !matched {
			remainingConsts = append(remainingConsts, c)
		}
	}

	p.Vars = remainingVars
	p.Consts = remainingConsts

	return
}

type Value struct {
	Doc       string
	Names     []string
	Items     []*ValueItem
	Signature string
}

type ValueItem struct {
	Doc   string
	Type  string
	Name  string
	Value string
}

type FuncParam struct {
	Type  string
	Names []string
}

type FuncReturn struct {
	Type  string
	Names []string
}

type Func struct {
	Doc       string
	Name      string
	Params    []*FuncParam
	Returns   []*FuncReturn
	Recv      []string
	Signature string
}

type Type struct {
	Doc        string
	Name       string
	Definition string
	Consts     []*Value
	Vars       []*Value
	Funcs      []*Func
	Methods    []*Func
}

func extractFunc(x *ast.FuncDecl) *Func {
	fn := Func{
		Doc:       x.Doc.Text(),
		Name:      x.Name.String(),
		Signature: generateFuncSignature(x),
	}
	if x.Recv != nil {
		for _, rcv := range x.Recv.List {
			fn.Recv = append(fn.Recv, typeString(rcv.Type))
		}
	}

	for _, field := range x.Type.Params.List {
		paramNames := []string{}
		for _, name := range field.Names {
			paramNames = append(paramNames, name.Name)
		}
		paramType := typeString(field.Type)
		param := &FuncParam{
			Type:  paramType,
			Names: paramNames,
		}
		fn.Params = append(fn.Params, param)
	}

	if x.Type.Results != nil {
		for _, field := range x.Type.Results.List {
			returnNames := []string{}
			for _, name := range field.Names {
				if name != nil {
					continue
				}
				returnNames = append(returnNames, name.Name)
			}
			returnType := typeString(field.Type)
			ret := &FuncReturn{
				Type:  returnType,
				Names: returnNames,
			}
			fn.Returns = append(fn.Returns, ret)
		}
	}

	return &fn
}

func extractValue(fset *token.FileSet, x *ast.GenDecl) (*Value, error) {
	value := Value{
		Doc: x.Doc.Text(),
	}
	x.Doc = nil
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, x); err != nil {
		return nil, err
	}
	value.Signature = buf.String()
	for _, spec := range x.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				if !name.IsExported() {
					continue
				}
				valueItem := ValueItem{
					Type: typeString(valueSpec.Type),
					Name: name.String(),
				}
				if len(valueSpec.Values) > i {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						valueItem.Value = lit.Value
					} else if ident, ok := valueSpec.Values[i].(*ast.Ident); ok {
						valueItem.Value = ident.Name
					}
				}
				if valueSpec.Doc != nil {
					valueItem.Doc = valueSpec.Doc.Text()
				}
				value.Items = append(value.Items, &valueItem)
			}
		}
	}
	return &value, nil
}

func extractType(fset *token.FileSet, x *ast.TypeSpec) (*Type, error) {
	newType := Type{}
	newType.Name = x.Name.String()
	newType.Doc = x.Doc.Text()

	x.Doc = nil
	var buf bytes.Buffer
	buf.WriteString("type ")
	if err := format.Node(&buf, fset, x); err != nil {
		return nil, err
	}

	newType.Definition = buf.String()
	return &newType, nil
}
