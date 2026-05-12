package gnolang

import (
	"math/big"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoundedBuf_TruncatedSuffix(t *testing.T) {
	w := newBoundedBuf(5)
	n, err := w.WriteString("abcdefghij")
	require.NoError(t, err)
	require.Equal(t, 10, n) // returns full input length even when truncated
	assert.True(t, w.Truncated())
	assert.Equal(t, "abcde...", w.String())
}

func TestBoundedBuf_NoTruncation(t *testing.T) {
	w := newBoundedBuf(20)
	w.WriteString("hello")
	assert.False(t, w.Truncated())
	assert.Equal(t, "hello", w.String())
}

func TestBoundedSprintTV_String_Small(t *testing.T) {
	tv := TypedValue{T: StringType, V: StringValue("hello")}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "hello", out)
}

func TestBoundedSprintTV_String_Huge(t *testing.T) {
	huge := strings.Repeat("a", 10*1024*1024) // 10 MB
	tv := TypedValue{T: StringType, V: StringValue(huge)}
	out := BoundedSprintTV(tv, nil, 1024)
	assert.LessOrEqual(t, len(out), 1024+3, // +3 for "..." suffix
		"output must respect cap (got %d)", len(out))
	assert.True(t, strings.HasPrefix(out, "a"), "should start with content")
}

func TestBoundedSprintTV_String_HugeNoTransientBlowup(t *testing.T) {
	// Verify the source is pre-truncated before strconv.Quote runs.
	// We can't directly measure heap, but we can confirm the
	// output is bounded and the call returns quickly even for a
	// 100 MB input.
	huge := strings.Repeat("x", 100*1024*1024)
	tv := TypedValue{T: StringType, V: StringValue(huge)}
	out := BoundedSprintTV(tv, nil, 1024)
	assert.LessOrEqual(t, len(out), 1027)
}

func TestBoundedSprintTV_BigInt_Small(t *testing.T) {
	bi := big.NewInt(12345)
	tv := TypedValue{T: UntypedBigintType, V: BigintValue{V: bi}}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "12345", out)
}

func TestBoundedSprintTV_BigInt_Huge(t *testing.T) {
	// 1 << (1<<20) = a 1M-bit number ≈ 300K decimal digits.
	bi := new(big.Int).Lsh(big.NewInt(1), 1<<20)
	tv := TypedValue{T: UntypedBigintType, V: BigintValue{V: bi}}
	out := BoundedSprintTV(tv, nil, 1024)
	// Should hit the BitLen pre-check: BitLen() ≈ 1M > 1024*3.
	assert.Contains(t, out, "<bigint, bits=")
	assert.LessOrEqual(t, len(out), 1024+3)
}

func TestBoundedSprintTV_BigDec_Small(t *testing.T) {
	bd := apd.New(123, -2) // 1.23
	tv := TypedValue{T: UntypedBigdecType, V: BigdecValue{V: bd}}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Contains(t, out, "1.23")
}

func TestBoundedSprintTV_BigDec_Huge(t *testing.T) {
	// Coefficient = 1 << (1<<20) ≈ 300K decimal digits. Models the
	// runtime case where apd's unlimited-precision Add lets a user
	// grow the coefficient over many gas-metered steps before
	// triggering a panic that carries the value through bounded
	// rendering.
	bd := new(apd.Decimal)
	bd.Coeff.Lsh(apd.NewBigInt(1), 1<<20)
	bd.Exponent = 0
	tv := TypedValue{T: UntypedBigdecType, V: BigdecValue{V: bd}}
	out := BoundedSprintTV(tv, nil, 1024)
	// Should hit the BitLen pre-check: BitLen() ≈ 1M > 1024*3.
	// Crucially, this must NOT allocate the full decimal string.
	assert.Contains(t, out, "<bigdec, bits=")
	// Sanity: rendered output stays under cap (modulo a 3-byte
	// "..." truncation suffix).
	assert.LessOrEqual(t, len(out), 1024+3)
}

func TestBoundedSprintTV_Int(t *testing.T) {
	tv := TypedValue{T: IntType}
	tv.SetInt(42)
	assert.Equal(t, "42", BoundedSprintTV(tv, nil, 100))
}

func TestBoundedSprintTV_Bool(t *testing.T) {
	tv := TypedValue{T: BoolType}
	tv.SetBool(true)
	assert.Equal(t, "true", BoundedSprintTV(tv, nil, 100))
}

func TestBoundedSprintTV_Array_Empty(t *testing.T) {
	av := &ArrayValue{}
	tv := TypedValue{V: av}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "[]", out)
}

func TestBoundedSprintTV_Array_OverChildrenCap(t *testing.T) {
	// 100 elements of "x" — should cap at MaxCompositeChildren (32).
	list := make([]TypedValue, 100)
	for i := range list {
		list[i] = TypedValue{T: StringType, V: StringValue("x")}
	}
	av := &ArrayValue{List: list}
	tv := TypedValue{V: av}
	out := BoundedSprintTV(tv, nil, 4096)
	assert.Contains(t, out, "<...68 more>", "should mark elided elements")
}

func TestBoundedSprintTV_ByteArray_Big(t *testing.T) {
	data := make([]byte, 1024*1024) // 1 MB byte array
	for i := range data {
		data[i] = byte(i)
	}
	av := &ArrayValue{Data: data}
	tv := TypedValue{V: av}
	out := BoundedSprintTV(tv, nil, 1024)
	assert.Contains(t, out, "0x")
	assert.Contains(t, out, "...total 1048576 bytes")
	assert.LessOrEqual(t, len(out), 1027)
}

func TestBoundedSprintTV_NestedComposite_DepthCap(t *testing.T) {
	// Build a struct of struct of struct ... 20 levels deep.
	innermost := &StructValue{Fields: []TypedValue{
		{T: IntType}, // value 0
	}}
	cur := innermost
	for i := 0; i < 20; i++ {
		outer := &StructValue{Fields: []TypedValue{{V: cur}}}
		cur = outer
	}
	tv := TypedValue{V: cur}
	out := BoundedSprintTV(tv, nil, 4096)
	// Should hit depth cap somewhere along the way, emitting "{...}".
	assert.Contains(t, out, "{...}", "depth cap should emit marker")
}

func TestBoundedSprintTV_Map(t *testing.T) {
	mv := &MapValue{List: &MapList{}}
	for i := 0; i < 5; i++ {
		key := TypedValue{T: StringType, V: StringValue("k")}
		val := TypedValue{T: IntType}
		val.SetInt(int64(i))
		mli := &MapListItem{Key: key, Value: val}
		if mv.List.Head == nil {
			mv.List.Head = mli
			mv.List.Tail = mli
		} else {
			mv.List.Tail.Next = mli
			mli.Prev = mv.List.Tail
			mv.List.Tail = mli
		}
		mv.List.Size++
	}
	tv := TypedValue{V: mv}
	out := BoundedSprintTV(tv, nil, 4096)
	assert.Contains(t, out, "{")
	assert.Contains(t, out, "}")
	assert.Contains(t, out, "k:") // raw bytes — no quotes around string keys
}

func TestBoundedSprintTV_Func(t *testing.T) {
	fv := &FuncValue{Name: Name("MyFunc")}
	tv := TypedValue{V: fv}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "<func MyFunc>", out)
}

func TestBoundedSprintTV_BoundMethod(t *testing.T) {
	fv := &FuncValue{Name: Name("Inc")}
	bmv := &BoundMethodValue{Func: fv}
	tv := TypedValue{V: bmv}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "<bound-method Inc>", out)
}

func TestBoundedSprintTV_Nil(t *testing.T) {
	tv := TypedValue{}
	out := BoundedSprintTV(tv, nil, 100)
	assert.Equal(t, "undefined", out)
}

func TestBoundedSprintException_Nil(t *testing.T) {
	assert.Equal(t, "<nil>", BoundedSprintException(nil, nil, 100))
}

func TestBoundedSprintException_Small(t *testing.T) {
	e := &Exception{Value: TypedValue{T: StringType, V: StringValue("oops")}}
	out := BoundedSprintException(e, nil, 100)
	assert.Equal(t, "oops", out)
}

func TestBoundedSprintException_HeadOnly(t *testing.T) {
	// Verify Previous chain is NOT rendered.
	older := &Exception{Value: TypedValue{T: StringType, V: StringValue("older")}}
	newer := &Exception{
		Value:    TypedValue{T: StringType, V: StringValue("newer")},
		Previous: older,
	}
	out := BoundedSprintException(newer, nil, 100)
	assert.Equal(t, "newer", out)
	assert.NotContains(t, out, "older", "chain should not be walked")
}

func TestBoundedStacktrace_Empty(t *testing.T) {
	out := BoundedStacktrace(Stacktrace{}, 1024)
	assert.Equal(t, "", out)
}

func TestBoundedStacktrace_Frames(t *testing.T) {
	s := Stacktrace{
		LastLine: 5,
		Calls: []StacktraceCall{
			{
				FuncName: "Inc",
				FuncLoc:  Location{PkgPath: "gno.land/r/x", File: "x.gno"},
			},
			{
				FuncName: "main",
				FuncLoc:  Location{PkgPath: "main", File: "main.gno"},
				CallExpr: &CallExpr{},
			},
		},
	}
	out := BoundedStacktrace(s, 4096)
	assert.Contains(t, out, "Inc at gno.land/r/x/x.gno:5")
	assert.Contains(t, out, "main at main/main.gno:0")
}

func TestBoundedStacktrace_Anonymous(t *testing.T) {
	s := Stacktrace{
		LastLine: 7,
		Calls: []StacktraceCall{
			{
				FuncName: "", // anonymous
				FuncLoc:  Location{PkgPath: "p", File: "f.gno"},
			},
		},
	}
	out := BoundedStacktrace(s, 1024)
	assert.Contains(t, out, "(anonymous) at p/f.gno:7")
}

func TestBoundedStacktrace_Defer(t *testing.T) {
	s := Stacktrace{
		LastLine: 9,
		Calls: []StacktraceCall{
			{IsDefer: true, FuncName: "cleanup", FuncLoc: Location{PkgPath: "p", File: "f.gno"}},
		},
	}
	out := BoundedStacktrace(s, 1024)
	assert.Contains(t, out, "defer cleanup at p/f.gno:9")
}

func TestBoundedStacktrace_FrameCap(t *testing.T) {
	// Build 100 frames; expect cap at MaxStacktraceFrames + elision marker.
	calls := make([]StacktraceCall, 100)
	for i := range calls {
		calls[i] = StacktraceCall{
			FuncName: "F",
			FuncLoc:  Location{PkgPath: "p", File: "f.gno"},
			CallExpr: &CallExpr{},
		}
	}
	s := Stacktrace{Calls: calls, LastLine: 1}
	out := BoundedStacktrace(s, 64*1024)
	// Should have exactly MaxStacktraceFrames frames + elision marker.
	frameCount := strings.Count(out, "F at p/f.gno:")
	assert.Equal(t, MaxStacktraceFrames, frameCount,
		"should render exactly %d frames", MaxStacktraceFrames)
	assert.Contains(t, out, "frame(s) elided")
}

func TestStacktraceFuncName_FreeFunc(t *testing.T) {
	fr := &Frame{Func: &FuncValue{Name: Name("Free")}}
	assert.Equal(t, "Free", stacktraceFuncName(fr))
}

func TestStacktraceFuncName_Method_Value(t *testing.T) {
	dt := &DeclaredType{PkgPath: "gno.land/r/foo", Name: Name("Counter")}
	fr := &Frame{
		Func:     &FuncValue{Name: Name("Inc")},
		Receiver: TypedValue{T: dt},
	}
	assert.Equal(t, "gno.land/r/foo.Counter.Inc", stacktraceFuncName(fr))
}

func TestStacktraceFuncName_Method_Pointer(t *testing.T) {
	dt := &DeclaredType{PkgPath: "gno.land/r/foo", Name: Name("Counter")}
	pt := &PointerType{Elt: dt}
	fr := &Frame{
		Func:     &FuncValue{Name: Name("Inc")},
		Receiver: TypedValue{T: pt},
	}
	assert.Equal(t, "(*gno.land/r/foo.Counter).Inc", stacktraceFuncName(fr))
}

func TestStacktraceFuncName_NilFunc(t *testing.T) {
	fr := &Frame{}
	assert.Equal(t, "", stacktraceFuncName(fr))
}

func TestTruncateForLog(t *testing.T) {
	assert.Equal(t, "abc", truncateForLog("abc", 5))
	assert.Equal(t, "abcde", truncateForLog("abcde", 5))
	assert.Equal(t, "ab...", truncateForLog("abcdef", 5))
}
