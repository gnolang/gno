package gnolang

import (
	goerrors "errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
)

var (
	reFloat    = regexp.MustCompile(`^[0-9\.]+([eE][\-\+]?[0-9]+)?$`)
	reHexFloat = regexp.MustCompile(`^0[xX][0-9a-fA-F\.]+([pP][\-\+]?[0-9a-fA-F]+)?$`)
)

func (m *Machine) doOpEval() {
	x := m.PeekExpr(1)
	if debug {
		debug.Printf("EVAL: (%T) %v\n", x, x)
	}
	// This case moved out of switch for performance.
	// TODO: understand this better.
	if nx, ok := x.(*NameExpr); ok {
		m.PopExpr()
		if nx.Path.Depth == 0 {
			// Name is in uverse (global).
			gv := Uverse().GetBlock(nil).GetPointerTo(nil, nx.Path)
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
			x.Value = strings.ReplaceAll(x.Value, blankIdentifier, "")
			// temporary optimization
			bi := big.NewInt(0)
			// TODO optimize.
			// TODO deal with base.
			var ok bool
			if len(x.Value) >= 2 && x.Value[0] == '0' {
				switch x.Value[1] {
				case 'b', 'B':
					_, ok = bi.SetString(x.Value[2:], 2)
				case 'o', 'O':
					_, ok = bi.SetString(x.Value[2:], 8)
				case 'x', 'X':
					_, ok = bi.SetString(x.Value[2:], 16)
				case '0', '1', '2', '3', '4', '5', '6', '7':
					_, ok = bi.SetString(x.Value, 8)
				default:
					ok = false
				}
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
			x.Value = strings.ReplaceAll(x.Value, blankIdentifier, "")

			if reFloat.MatchString(x.Value) {
				value := x.Value
				bd, c, err := apd.NewFromString(value)
				if err != nil {
					panic(fmt.Sprintf(
						"invalid decimal constant: %s",
						x.Value))
				}
				if c.Inexact() {
					panic(fmt.Sprintf(
						"could not represent decimal exactly: %s",
						x.Value))
				}
				m.PushValue(TypedValue{
					T: UntypedBigdecType,
					V: BigdecValue{V: bd},
				})
				return
			} else if reHexFloat.MatchString(x.Value) {
				originalInput := x.Value
				value := x.Value[2:]
				var hexString string
				var exp int64
				eIndex := strings.IndexAny(value, "Pp")
				if eIndex == -1 {
					panic("should not happen")
				}

				// ----------------------------------------
				// NewFromHexString()
				// TODO: move this to another function.

				// Step 1 get exp component.
				expInt, err := strconv.ParseInt(value[eIndex+1:], 10, 32)
				if err != nil {
					if e, ok := err.(*strconv.NumError); ok && goerrors.Is(e.Err, strconv.ErrRange) {
						panic(fmt.Sprintf(
							"can't convert %s to decimal: fractional part too long",
							value))
					}
					panic(fmt.Sprintf(
						"can't convert %s to decimal: exponent is not numeric",
						value))
				}
				value = value[:eIndex]
				exp = expInt
				// Step 2 adjust exp from dot.
				pIndex := -1
				vLen := len(value)
				for i := range vLen {
					if value[i] == '.' {
						if pIndex > -1 {
							panic(fmt.Sprintf(
								"can't convert %s to decimal: too many .s",
								value))
						}
						pIndex = i
					}
				}
				if pIndex == -1 {
					// There is no decimal point, we can just parse the original string as
					// a hex
					hexString = value
				} else {
					if pIndex+1 < vLen {
						hexString = value[:pIndex] + value[pIndex+1:]
					} else {
						hexString = value[:pIndex]
					}
					expInt := -len(value[pIndex+1:])
					exp += int64(expInt)
				}
				bexp := apd.New(0, 0)
				_, err = apd.BaseContext.WithPrecision(1024).Pow(
					bexp,
					apd.New(2, 0),
					apd.New(exp, 0))
				if err != nil {
					panic(fmt.Sprintf("error computing exponent: %v", err))
				}
				// Step 3 make Decimal from mantissa and exp.
				dValue := new(apd.BigInt)
				_, ok := dValue.SetString(hexString, 16)
				if !ok {
					panic(fmt.Sprintf("can't convert %s to decimal", value))
				}
				if exp < math.MinInt32 || exp > math.MaxInt32 {
					// NOTE(vadim): I doubt a string could realistically be this long
					panic(fmt.Sprintf("can't convert %s to decimal: fractional part too long", originalInput))
				}
				res := apd.New(0, 0)
				_, err = apd.BaseContext.WithPrecision(1024).Mul(
					res,
					apd.NewWithBigInt(dValue, 0),
					bexp)
				if err != nil {
					panic(fmt.Sprintf("canot calculate hexadecimal: %v", err))
				}

				// NewFromHexString() END
				// ----------------------------------------

				m.PushValue(TypedValue{
					T: UntypedBigdecType,
					V: BigdecValue{V: res},
				})
				return
			} else {
				panic(fmt.Sprintf("unexpected decimal/float format %s", x.Value))
			}
		case IMAG:
			// NOTE: this is a syntax and grammar problem, not an
			// AST one.  Imaginaries should get evaluated as a
			// type like any other.  See
			// github.com/Quasilyte/go-complex-nums-emulation
			// and github.com/golang/go/issues/19921
			panic("imaginaries are not supported")
		case CHAR:
			// Matching character literal parsing in go/constant.MakeFromLiteral.
			val := x.Value
			rne, _, _, err := strconv.UnquoteChar(val[1:len(val)-1], '\'')
			if err != nil {
				panic("error in parsing character literal: " + err.Error())
			}
			tv := TypedValue{T: UntypedRuneType}
			tv.SetInt32(rne)
			m.PushValue(tv)
		case STRING:
			m.PushValue(TypedValue{
				T: UntypedStringType,
				V: m.Alloc.NewString(x.GetString()),
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
		// Eval args.
		args := x.Args
		for i := len(args) - 1; 0 <= i; i-- {
			m.PushExpr(args[i])
			m.PushOp(OpEval)
		}
		// evaluate func
		m.PushExpr(x.Func)
		m.PushOp(OpEval)
	case *IndexExpr:
		if x.HasOK {
			m.PushOp(OpIndex2)
		} else {
			m.PushOp(OpIndex1)
		}
		// evaluate index
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
		// evaluate max
		if x.Max != nil {
			m.PushExpr(x.Max)
			m.PushOp(OpEval)
		}
		// evaluate high
		if x.High != nil {
			m.PushExpr(x.High)
			m.PushOp(OpEval)
		}
		// evaluate low
		if x.Low != nil {
			m.PushExpr(x.Low)
			m.PushOp(OpEval)
		}
		// evaluate x
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
	case *ConstExpr:
		m.PopExpr()
		// push preprocessed value
		tv := x.TypedValue
		// see .pkgSelector; const(ref(pkgPath)).  do not fill in;
		// nodes may be more persistent than values in a tx.
		// (currently all nodes are cached, but we don't want to cache
		// all packages too).
		m.PushValue(tv)
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
