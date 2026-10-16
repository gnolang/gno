package gnolang

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Rules for import paths, constants and branch statements (Go spec:
// "Import declarations", "Constant expressions", "Break statements",
// "Continue statements").

// assertValidImportPath checks the spelling of an import path: not
// empty, valid UTF-8, graphic characters outside Go's excluded set, and
// not absolute.
func assertValidImportPath(path string) {
	if path == "" {
		fail(CodeImportPath, "invalid import path (empty string)")
	}
	if !utf8.ValidString(path) {
		failf(CodeImportPath, nil, "import path contains invalid UTF-8 character: %q", path)
	}
	for _, r := range path {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) ||
			strings.ContainsRune("!\"#$%&'()*,:;<=>?[\\]^`{|}", r) {
			failf(CodeImportPath, nil, "invalid import path %q: invalid character %q", path, r)
		}
	}
	if strings.HasPrefix(path, "/") {
		failf(CodeImportPath, nil, "import path cannot be absolute path: %q", path)
	}
}

// constWidth is the width go/types gives untyped integer constants.
const constWidth = 512

// assertConstWidth rejects a folded untyped integer constant wider than
// constWidth bits, naming the operation that produced it.
func assertConstWidth(x Expr, cv *TypedValue) {
	if cv.T != UntypedBigintType || cv.GetBigInt().BitLen() <= constWidth {
		return
	}
	op := ""
	switch x := x.(type) {
	case *BinaryExpr:
		op = constOpNames[x.Op]
	case *UnaryExpr:
		switch x.Op {
		case XOR:
			op = "bitwise complement"
		case SUB:
			op = "negation"
		}
	}
	if op == "" {
		fail(CodeConstOverflow, "constant overflow")
	}
	failf(CodeConstOverflow, nil, "constant %s overflow", op)
}

var constOpNames = map[Word]string{
	ADD: "addition", SUB: "subtraction", MUL: "multiplication", QUO: "division",
	SHL: "shift", SHR: "shift", XOR: "bitwise XOR", BAND: "bitwise AND",
	BOR: "bitwise OR", BAND_NOT: "bitwise AND NOT",
}

// assertBranchTarget checks that the statement a labeled break or
// continue names can be its target: a loop, or for break also a switch
// or select.
func assertBranchTarget(n *BranchStmt, bn BlockNode) {
	switch bn.(type) {
	case *ForStmt, *RangeStmt:
		return
	case *SwitchClauseStmt, *SelectCaseStmt:
		if n.Op == BREAK {
			return
		}
	}
	if n.Op == BREAK {
		failf(CodeBranchLabel, nil, "invalid break label %s", n.Label)
	}
	failf(CodeBranchLabel, nil, "invalid continue label %s", n.Label)
}

func assertValidConstExpr(store Store, last BlockNode, n *ValueDecl, expr Expr) {
	if n.Type != nil {
		nt := evalStaticType(store, last, n.Type)
		if _, ok := baseOf(nt).(PrimitiveType); !ok {
			panic(fmt.Sprintf("invalid constant type %s", nt.String()))
		}
	}

	nt := evalStaticTypeOf(store, last, expr)

	if nt == nil {
		panic(fmt.Sprintf("%s (variable of type nil) is not constant", expr))
	}

	if _, ok := baseOf(nt).(PrimitiveType); !ok {
		panic(fmt.Sprintf("%s (variable of type %s) is not constant", expr, nt))
	}

	assertValidConstValue(store, last, expr)
}

func assertValidConstValue(store Store, last BlockNode, currExpr Expr) {
Main:
	switch currExpr := currExpr.(type) {
	case *ConstExpr:
	case *UnaryExpr:
		// *, & is filter out previously since they are not primitive
		assertValidConstValue(store, last, currExpr.X)
	case *TypeAssertExpr:
		panic(fmt.Sprintf("%s (comma, ok expression of type %s) is not constant", currExpr.String(), currExpr.Type))
	case *CallExpr:
		ift := evalStaticTypeOf(store, last, currExpr.Func)
		switch baseOf(ift).(type) {
		case *FuncType:
			tup := evalStaticTypeOfRaw(store, last, currExpr).(*tupleType)

			// check for built-in functions
			if cx, ok := currExpr.Func.(*ConstExpr); ok {
				if fv, ok := cx.V.(*FuncValue); ok {
					if fv.PkgPath == uversePkgPath {
						// TODO: should support min, max, real, imag
						switch fv.Name {
						case "len":
							at := evalStaticTypeOf(store, last, currExpr.Args[0])
							if _, ok := unwrapPointerType(baseOf(at)).(*ArrayType); ok {
								// ok
								break Main
							}
							assertValidConstValue(store, last, currExpr.Args[0])
							break Main
						case "cap":
							at := evalStaticTypeOf(store, last, currExpr.Args[0])
							if _, ok := unwrapPointerType(baseOf(at)).(*ArrayType); ok {
								// ok
								break Main
							}
							assertValidConstValue(store, last, currExpr.Args[0])
							break Main
						}
					}
				}
			}

			switch {
			case len(tup.Elts) == 0:
				panic(fmt.Sprintf("%s (no value) used as value", currExpr.String()))
			case len(tup.Elts) == 1:
				panic(fmt.Sprintf("%s (value of type %s) is not constant", currExpr.String(), tup.Elts[0]))
			default:
				panic(fmt.Sprintf("multiple-value %s (value of type %s) in single-value context", currExpr.String(), tup.Elts))
			}
		case *TypeType:
			for _, arg := range currExpr.Args {
				assertValidConstValue(store, last, arg)
			}
		default:
			panic(fmt.Sprintf(
				"unexpected func type %v (%v)",
				ift, reflect.TypeOf(ift)))
		}
	case *BinaryExpr:
		assertValidConstValue(store, last, currExpr.Left)
		assertValidConstValue(store, last, currExpr.Right)
	case *SelectorExpr:
		xt := evalStaticTypeOf(store, last, currExpr.X)
		switch xt := xt.(type) {
		case *PackageType:
			var pv *PackageValue
			if cx, ok := currExpr.X.(*ConstExpr); ok {
				// NOTE: *Machine.TestMemPackage() needs this
				// to pass in an imported package as *ConstEzpr.
				pv = cx.V.(*PackageValue)
			} else {
				// otherwise, packages can only be referred to by
				// *NameExprs, and cannot be copied.
				pvc := evalConst(store, last, currExpr.X)
				pv_, ok := pvc.V.(*PackageValue)
				if !ok {
					panic(fmt.Sprintf(
						"missing package in selector expr %s",
						currExpr.String()))
				}
				pv = pv_
			}
			if pv.GetBlock(store).Source.GetIsConst(store, currExpr.Sel) {
				break Main
			}

			tt := pv.GetBlock(store).Source.GetStaticTypeOf(store, currExpr.Sel)
			panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), tt))
		case *PointerType, *DeclaredType, *StructType, *InterfaceType, *TypeType:
			ty := evalStaticTypeOf(store, last, currExpr)
			if _, ok := ty.(*TypeType); ok {
				ty = evalStaticType(store, last, currExpr)
			}
			panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), ty))
		default:
			fail(CodeMissingField, fmt.Sprintf(
				"%s.%s undefined (type %s has no field or method %s)",
				exprString(currExpr.X), currExpr.Sel, xt.String(), currExpr.Sel), currExpr)
		}
	default:
		ift := evalStaticTypeOf(store, last, currExpr)
		if _, ok := ift.(*TypeType); ok {
			ift = evalStaticType(store, last, currExpr)
		}
		panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), ift))
	}
}
