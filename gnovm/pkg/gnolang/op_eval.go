package gnolang

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/overflow"
)

var (
	reFloat    = regexp.MustCompile(`^[0-9\.]+([eE][\-\+]?[0-9]+)?$`)
	reHexFloat = regexp.MustCompile(`^0[xX][0-9a-fA-F\.]+([pP][\-\+]?[0-9a-fA-F]+)?$`)
)

// chargeBigLitParse charges quadratic CPU gas for the O(n^2) parse of an
// untyped bigint/bigdec literal (big.Int.SetString / big.ParseFloat) before it
// runs, so a huge literal OOGs first: gas = (chars/10)^2 * slope / 10, via
// overflow.Mulp as in incrCPUBigDecQuad.
//
// rawLen must be len(x.Value) with x.Value unmutated, i.e. the literal as
// written in source with separators included: the blank-identifier strip
// scans (and may copy) the whole literal outside m.Alloc, so charging on the
// stripped length would leave that pass unbilled and, worse, would bill a
// node's first evaluation differently from later ones. doOpEval therefore
// must not write the stripped form back into the AST node. Pinned by
// TestBigLitParseChargeIsEvalInvariant; see the ADR for why that is latent
// rather than reachable today.
func (m *Machine) chargeBigLitParse(rawLen int, slope int64) {
	d10 := int64(rawLen) / 10 // chars/10; prefix, '.', exponent and separators all counted
	m.incrCPU(overflow.Mulp(overflow.Mulp(d10, d10), slope) / 10)
}

// stripBlanks removes the blank identifiers Go allows as digit separators in
// numeric literals. It deliberately returns a new string rather than updating
// the AST node in place; see chargeBigLitParse.
func stripBlanks(value string) string {
	return strings.ReplaceAll(value, blankIdentifier, "")
}

func (m *Machine) doOpEval() {
	x := m.PeekExpr(1)
	m.Lastline = x.GetLine()
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
		}
		// Get value from scope.
		m.incrCPU(OpCPUSlopeEvalNameExpr * int64(nx.Path.Depth))
		lb := m.LastBlock()
		// Inline parent traversal for common shallow depths
		// to avoid GetPointerTo function call + blank check + loop overhead.
		b := lb
		for i := uint8(1); i < nx.Path.Depth; i++ {
			b = b.GetParent(m.Store)
		}
		ptr := b.GetPointerToInt(m.Store, int(nx.Path.Index))
		m.PushValue(ptr.Deref())
		return
	}
	switch x := x.(type) {
	// case NameExpr: handled above
	case *BasicLitExpr:
		m.PopExpr()
		switch x.Kind {
		case INT:
			m.chargeBigLitParse(len(x.Value), OpCPUSlopeBigIntSetString)
			value := stripBlanks(x.Value)
			// temporary optimization
			bi := big.NewInt(0)
			// TODO optimize.
			// TODO deal with base.
			var ok bool
			if len(value) >= 2 && value[0] == '0' {
				switch value[1] {
				case 'b', 'B':
					_, ok = bi.SetString(value[2:], 2)
				case 'o', 'O':
					_, ok = bi.SetString(value[2:], 8)
				case 'x', 'X':
					_, ok = bi.SetString(value[2:], 16)
				case '0', '1', '2', '3', '4', '5', '6', '7':
					_, ok = bi.SetString(value, 8)
				default:
					ok = false
				}
				if !ok {
					panic(fmt.Sprintf(
						"invalid integer constant: %s",
						value))
				}
			} else {
				_, ok := bi.SetString(value, 10)
				if !ok {
					panic(fmt.Sprintf(
						"invalid integer constant: %s",
						value))
				}
			}
			m.PushValue(TypedValue{
				T: UntypedBigintType,
				V: BigintValue{V: bi},
			})
		case FLOAT:
			// Charge FLOAT too: a trailing ".0"/"e1" routes an integer through
			// parseBigdecLiteral (big.ParseFloat + big.Rat.SetString), so an
			// INT-only charge is trivially bypassed.
			m.chargeBigLitParse(len(x.Value), OpCPUSlopeBigDecParse)
			value := stripBlanks(x.Value)

			if reFloat.MatchString(value) {
				m.PushValue(TypedValue{
					T: UntypedBigdecType,
					V: parseBigdecLiteral(value, "decimal"),
				})
				return
			} else if reHexFloat.MatchString(value) {
				m.PushValue(TypedValue{
					T: UntypedBigdecType,
					V: parseBigdecLiteral(value, "hex float"),
				})
				return
			} else {
				panic(fmt.Sprintf("unexpected decimal/float format %s", value))
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
	default:
		panic(fmt.Sprintf("unexpected expression %#v", x))
	}
}
