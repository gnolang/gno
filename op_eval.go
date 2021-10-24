package gno

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func (m *Machine) doOpEval() {
	x := m.PeekExpr(1)
	if debug {
		debug.Printf("EVAL: %v\n", x)
		//fmt.Println(m.String())
	}
	// This case moved out of switch for performance.
	// TODO: understand this better.
	if nx, ok := x.(*NameExpr); ok {
		m.PopExpr()
		if nx.Path.Depth == 0 {
			// Name is in uverse (global).
			gv := Uverse().GetPointerTo(nil, nx.Path)
			m.PushValue(gv.Deref())
			return
		} else {
			// Get value from scope.
			lb := m.LastBlock()
			// Push value, done.
			ptr := lb.GetPointerTo(m.Store, nx.Path)
			m.PushValue(ptr.Deref())
			return
		}
	}
	switch x := x.(type) {
	// case NameExpr: handled above
	case *BasicLitExpr:
		m.PopExpr()
		switch x.Kind {
		case INT:
			// temporary optimization
			bi := big.NewInt(0)
			// TODO optimize.
			// TODO deal with base.
			if len(x.Value) > 2 && x.Value[0] == '0' &&
				strings.ContainsAny(x.Value[1:2], "bBoOxX") {
				_, ok := bi.SetString(x.Value[2:], 16)
				if !ok {
					panic(fmt.Sprintf(
						"invalid integer constant: %s",
						x.Value))
				}
			} else {
				_, ok := bi.SetString(x.Value, 10)
				if !ok {
					panic(fmt.Sprintf(
						"invalid integer constant: %s",
						x.Value))
				}
			}
			m.PushValue(TypedValue{
				T: UntypedBigintType,
				V: BigintValue{V: bi},
			})
		case FLOAT:
			// NOTE: I suspect we won't get hardware-level
			// consistency (determinism) in floating point numbers
			// yet, so hold off on this until we master this.
			panic("floats are not supported")
		case IMAG:
			// NOTE: this is a syntax and grammar problem, not an
			// AST one.  Imaginaries should get evaluated as a
			// type like any other.  See
			// github.com/Quasilyte/go-complex-nums-emulation
			// and github.com/golang/go/issues/19921
			panic("imaginaries are not supported")
		case CHAR:
			cstr, err := strconv.Unquote(x.Value)
			if err != nil {
				panic("error in parsing character literal: " + err.Error())
			}
			runes := []rune(cstr)
			if len(runes) != 1 {
				panic(fmt.Sprintf("error in parsing character literal: 1 rune expected, but got %v (%s)", len(runes), cstr))
			}
			tv := TypedValue{T: UntypedRuneType}
			tv.SetInt32(int32(rune(runes[0])))
			m.PushValue(tv)
		case STRING:
			m.PushValue(TypedValue{
				T: UntypedStringType,
				V: StringValue(x.GetString()),
			})
		default:
			panic(fmt.Sprintf("unexpected lit kind %v", x.Kind))
		}
	case *BinaryExpr:
		switch x.Op {
		case LAND, LOR:
			m.PushOp(OpBinary1)
			// evaluate left
			m.PushExpr(x.Left)
			m.PushOp(OpEval)
		default:
			op := word2BinaryOp(x.Op)
			m.PushOp(op)
			// alt: m.PushOp(OpBinary2)
			// evaluate right
			m.PushExpr(x.Right)
			m.PushOp(OpEval)
			// evaluate left
			m.PushExpr(x.Left)
			m.PushOp(OpEval)
		}
	case *CallExpr:
		m.PushOp(OpPrecall)
		// evaluate func
		m.PushExpr(x.Func)
		m.PushOp(OpEval)
	case *IndexExpr:
		if x.HasOK {
			m.PushOp(OpIndex2)
		} else {
			m.PushOp(OpIndex1)
		}
		// evalaute index
		m.PushExpr(x.Index)
		m.PushOp(OpEval)
		// evaluate x
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *SelectorExpr:
		m.PushOp(OpSelector)
		// evaluate x
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *SliceExpr:
		m.PushOp(OpSlice)
		// evalaute max
		if x.Max != nil {
			m.PushExpr(x.Max)
			m.PushOp(OpEval)
		}
		// evalaute high
		if x.High != nil {
			m.PushExpr(x.High)
			m.PushOp(OpEval)
		}
		// evalaute low
		if x.Low != nil {
			m.PushExpr(x.Low)
			m.PushOp(OpEval)
		}
		// evalaute x
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *StarExpr:
		m.PopExpr()
		m.PushOp(OpStar)
		// evaluate x.
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *RefExpr:
		m.PushOp(OpRef)
		// evaluate x
		m.PushForPointer(x.X)
	case *UnaryExpr:
		op := word2UnaryOp(x.Op)
		m.PushOp(op)
		// evaluate x
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *CompositeLitExpr:
		m.PushOp(OpCompositeLit)
		// evaluate type
		m.PushExpr(x.Type)
		m.PushOp(OpEval)
	case *FuncLitExpr:
		m.PushOp(OpFuncLit)
		// evaluate func type
		m.PushExpr(&x.Type)
		m.PushOp(OpEval)
	case *constExpr:
		m.PopExpr()
		// push preprocessed value
		m.PushValue(x.TypedValue)
	case *constTypeExpr:
		m.PopExpr()
		// push preprocessed type as value
		m.PushValue(asValue(x.Type))
	case *FieldTypeExpr:
		m.PushOp(OpFieldType)
		// evaluate field type
		m.PushExpr(x.Type)
		m.PushOp(OpEval)
		// evaluate tag?
		if x.Tag != nil {
			m.PushExpr(x.Tag)
			m.PushOp(OpEval)
		}
	case *ArrayTypeExpr:
		m.PushOp(OpArrayType)
		// evaluate length if set
		if x.Len != nil {
			m.PushExpr(x.Len)
			m.PushOp(OpEval) // OpEvalPrimitive?
		}
		// evaluate elem type
		m.PushExpr(x.Elt)
		m.PushOp(OpEval) // OpEvalType?
	case *SliceTypeExpr:
		m.PushOp(OpSliceType)
		// evaluate elem type
		m.PushExpr(x.Elt)
		m.PushOp(OpEval) // OpEvalType?
	case *InterfaceTypeExpr:
		m.PushOp(OpInterfaceType)
		// evaluate methods
		for i := len(x.Methods) - 1; 0 <= i; i-- {
			m.PushExpr(&x.Methods[i])
			m.PushOp(OpEval)
		}
	case *FuncTypeExpr:
		// NOTE params and results are evaluated in
		// the parent scope.
		m.PushOp(OpFuncType)
		// evaluate results (after params)
		for i := len(x.Results) - 1; 0 <= i; i-- {
			m.PushExpr(&x.Results[i])
			m.PushOp(OpEval)
		}
		// evaluate params
		for i := len(x.Params) - 1; 0 <= i; i-- {
			m.PushExpr(&x.Params[i])
			m.PushOp(OpEval)
		}
	case *MapTypeExpr:
		m.PopExpr()
		m.PushOp(OpMapType)
		// evaluate value type
		m.PushExpr(x.Value)
		m.PushOp(OpEval) // OpEvalType?
		// evaluate key type
		m.PushExpr(x.Key)
		m.PushOp(OpEval) // OpEvalType?
	case *StructTypeExpr:
		m.PushOp(OpStructType)
		// evaluate fields
		for i := len(x.Fields) - 1; 0 <= i; i-- {
			m.PushExpr(&x.Fields[i])
			m.PushOp(OpEval)
		}
	case *TypeAssertExpr:
		if x.HasOK {
			m.PushOp(OpTypeAssert2)
		} else {
			m.PushOp(OpTypeAssert1)
		}
		// evaluate type
		m.PushExpr(x.Type)
		m.PushOp(OpEval)
		// evaluate x
		m.PushExpr(x.X)
		m.PushOp(OpEval)
	case *ChanTypeExpr:
		m.PushOp(OpChanType)
		m.PushExpr(x.Value)
		m.PushOp(OpEval) // OpEvalType?
	default:
		panic(fmt.Sprintf("unexpected expression %#v", x))
	}
}
