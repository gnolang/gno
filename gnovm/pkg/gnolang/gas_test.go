package gnolang

import (
	"fmt"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommentGasRuntimeOverhead reproduces the runtime gas overhead reported in
// gnolang/gno#4919 and demonstrates that it is caused by comments shifting
// line numbers in the source code.
//
// Mechanism:
//
//	Source comments → parser (go2gno.go SpanFromGo) records higher line numbers
//	  → Span{Pos.Line, End.Line} stored in AST node Attributes
//	  → on realm save, FuncValue.Source becomes RefNode{Location: Span{...}}
//	  → amino serializes Pos.Line as varint (binary_encode.go:141 EncodeVarint)
//	  → varint uses zig-zag encoding: value N → 2N, then 7 bits per byte
//
// Varint byte thresholds for line numbers:
//
//	Line 1–63     → zig-zag 2–126     → 1 byte
//	Line 64–8191  → zig-zag 128–16382 → 2 bytes
//	Line 8192+    → zig-zag 16384+    → 3 bytes
//
// Each Span has 4 integers (Line, Column, End.Line, End.Column). When comments
// push a function from line 3 to line 94, Pos.Line crosses the 1→2 byte
// threshold, adding +1 byte per integer that crosses. With multiple functions,
// these bytes accumulate — in this test, +4 bytes total for the heavy-comment
// package, which at GasGetObject=16/byte translates to +64 extra gas.
//
// The test uses source code based on the original issue's benchmark:
// https://github.com/gnoswap-labs/gnoswap/blob/main/tests/integration/testdata/base/comment_gas_measurement.txtar
//
// Note: the gas delta here (+64) is smaller than the integration test's
// (+1458/+2264) because this unit test only measures the direct package objects
// (FuncValues + PackageValue), not the full dependency graph (caller realm,
// cross-realm imports, type loading) that the end-to-end txtar test exercises.
func TestCommentGasRuntimeOverhead(t *testing.T) {
	// Source code based on the original issue's benchmark realms.
	// Package names are same-length (9 chars each) to eliminate PkgName/PkgPath
	// serialization bias, so all object size diffs are purely comment-caused.
	sources := map[string]string{
		"commentno": `package commentno

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

func Subtract(a, b int) int {
	return a - b
}

func Divide(a, b int) int {
	if b == 0 {
		panic("division by zero")
	}
	return a / b
}
`,
		"commentlo": `package commentlo

// Add adds two integers and returns the result.
// This function performs basic addition operation.
func Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers and returns the result.
// This function performs basic multiplication operation.
func Multiply(a, b int) int {
	return a * b
}

// Subtract subtracts second integer from first and returns the result.
// This function performs basic subtraction operation.
func Subtract(a, b int) int {
	return a - b
}

// Divide divides first integer by second and returns the result.
// This function performs basic division operation.
// It panics if divisor is zero.
func Divide(a, b int) int {
	if b == 0 {
		panic("division by zero")
	}
	return a / b
}
`,
		"commenthi": `// Package commenthi provides mathematical operations with extensive documentation.
// This package is designed to test whether comments affect gas consumption.
// All functions in this package perform basic arithmetic operations.
//
// The purpose of this test is to determine if the Gno VM charges gas for
// parsing and storing comments, or if comments are stripped during compilation.
//
// If comments do affect gas, developers should be mindful of comment length
// in production contracts to optimize gas usage.
package commenthi

// Add adds two integers and returns the result.
//
// This function takes two integer parameters and returns their sum.
// It is one of the most basic arithmetic operations available.
//
// Parameters:
//   - a: The first integer operand
//   - b: The second integer operand
//
// Returns:
//   - int: The sum of a and b
//
// Example:
//   result := Add(5, 3)  // result = 8
//
// Note: This function does not check for integer overflow.
// For large numbers, consider using a big integer library.
func Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers and returns the result.
//
// This function takes two integer parameters and returns their product.
// Multiplication is a fundamental arithmetic operation.
//
// Parameters:
//   - a: The first integer operand (multiplicand)
//   - b: The second integer operand (multiplier)
//
// Returns:
//   - int: The product of a and b
//
// Example:
//   result := Multiply(5, 3)  // result = 15
//
// Note: This function does not check for integer overflow.
// For large numbers, consider using a big integer library.
func Multiply(a, b int) int {
	return a * b
}

// Subtract subtracts the second integer from the first and returns the result.
//
// This function takes two integer parameters and returns their difference.
// The order of operands matters: a - b is not the same as b - a.
//
// Parameters:
//   - a: The integer to subtract from (minuend)
//   - b: The integer to subtract (subtrahend)
//
// Returns:
//   - int: The difference (a - b)
//
// Example:
//   result := Subtract(5, 3)  // result = 2
//
// Note: This function can return negative numbers if b > a.
func Subtract(a, b int) int {
	return a - b
}

// Divide divides the first integer by the second and returns the result.
//
// This function takes two integer parameters and returns their quotient.
// Integer division truncates toward zero.
//
// Parameters:
//   - a: The integer to be divided (dividend)
//   - b: The integer to divide by (divisor)
//
// Returns:
//   - int: The quotient of a / b
//
// Panics:
//   - If b is zero, the function panics with "division by zero"
//
// Example:
//   result := Divide(10, 3)  // result = 3 (truncated)
//
// Note: This function performs integer division. For precise division,
// consider using floating-point numbers or a decimal library.
func Divide(a, b int) int {
	// Check for division by zero to prevent runtime panic
	// This is a critical safety check that must be performed
	// before any division operation
	if b == 0 {
		panic("division by zero")
	}
	// Perform the division and return the result
	// Note: Integer division truncates toward zero
	return a / b
}
`,
	}

	type funcInfo struct {
		name        string
		spanLine    int // start line from RefNode.Location.Span
		spanEnd     int // end line
		size        int // amino serialized bytes of FuncValue
		refNodeSize int // amino serialized bytes of RefNode alone
	}

	type pkgResult struct {
		pkgValueSize  int
		funcs         []funcInfo
		totalObjBytes int
		gasGetObject  int64 // estimated gas = totalObjBytes * GasGetObject(16)
	}

	deployAndInspect := func(t *testing.T, name, body, pkgPath string) pkgResult {
		t.Helper()

		db := memdb.NewMemDB()
		baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
		iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), stypes.StoreOptions{})
		st := NewStore(nil, baseStore, iavlStore)

		m := NewMachineWithOptions(MachineOptions{
			PkgPath: pkgPath,
			Store:   st,
		})
		defer m.Release()

		mpkg := &std.MemPackage{
			Name:  name,
			Path:  pkgPath,
			Files: []*std.MemFile{{Name: "main.gno", Body: body}},
		}
		mpkg.Type = MPUserAll

		_, pv := m.RunMemPackage(mpkg, true)
		require.NotNil(t, pv)

		var res pkgResult

		// Serialize PackageValue.
		o2 := copyValueWithRefs(pv)
		bz := amino.MustMarshalAny(o2)
		res.pkgValueSize = len(bz)
		res.totalObjBytes = len(bz)

		// Inspect each FuncValue.
		blk := pv.GetBlock(st)
		for _, tv := range blk.Values {
			fv, ok := tv.V.(*FuncValue)
			if !ok {
				continue
			}

			// Get the Span from the source BlockNode.
			loc := fv.Source.GetLocation()

			// Serialize the RefNode alone — this is the Source field
			// that carries the Location with Span (line numbers).
			refNode := toRefNode(fv.Source)
			refNodeBz := amino.MustMarshalAny(refNode)

			// Serialize the FuncValue as it would be stored.
			o2 := copyValueWithRefs(fv)
			bz := amino.MustMarshalAny(o2)

			res.funcs = append(res.funcs, funcInfo{
				name:        string(fv.Name),
				spanLine:    loc.Span.Pos.Line,
				spanEnd:     loc.Span.End.Line,
				size:        len(bz),
				refNodeSize: len(refNodeBz),
			})
			res.totalObjBytes += len(bz)
		}

		// Sort funcs by name for stable output.
		sort.Slice(res.funcs, func(i, j int) bool {
			return res.funcs[i].name < res.funcs[j].name
		})

		res.gasGetObject = int64(res.totalObjBytes) * DefaultGasConfig().GasGetObject
		return res
	}

	// All package names are 9 chars and paths are 25 chars, so PkgName and
	// PkgPath contribute zero bias. Any object size diff is purely comment-caused.
	results := map[string]pkgResult{
		"commentno": deployAndInspect(t, "commentno", sources["commentno"], "gno.land/r/test/commentno"),
		"commentlo": deployAndInspect(t, "commentlo", sources["commentlo"], "gno.land/r/test/commentlo"),
		"commenthi": deployAndInspect(t, "commenthi", sources["commenthi"], "gno.land/r/test/commenthi"),
	}

	// --- Print detailed comparison ---
	names := []string{"commentno", "commentlo", "commenthi"}

	// Collect all function names (sorted by deployAndInspect).
	var funcNames []string
	for _, fi := range results["commentno"].funcs {
		funcNames = append(funcNames, fi.name)
	}

	// For each function, show how comments shift its Span line numbers and
	// how that inflates the serialized RefNode and FuncValue sizes.
	//
	// Each FuncValue is stored with a RefNode (Source field) that records the
	// function's Location.Span — start/end line numbers. Amino encodes these
	// as varints, so higher line numbers (pushed down by preceding comments)
	// require more bytes.
	//
	// Since all paths are the same length, FuncValue diff == RefNode diff directly.
	//
	// varintSize returns the number of bytes needed to encode a signed int
	// as a protobuf/amino varint (zig-zag + 7 bits per byte).
	varintSize := func(n int) int {
		// zig-zag: positive n → 2n
		zz := uint64(n) << 1
		size := 1
		for zz >= 128 {
			zz >>= 7
			size++
		}
		return size
	}

	baselineNo := results["commentno"]
	t.Logf("")
	t.Logf("=== Per-function: how comments shift Span and inflate serialized sizes ===")
	t.Logf("  Each FuncValue stores a RefNode with Location.Span (start/end line numbers).")
	t.Logf("  Amino encodes line numbers as varints: higher lines → more bytes.")
	t.Logf("  Varint thresholds: line 1–63 → 1 byte, line 64–8191 → 2 bytes, line 8192+ → 3 bytes.")
	t.Logf("  FuncValue diff = RefNode diff (from Span) + PkgPath diff (test artifact).")
	t.Logf("")
	for _, fn := range funcNames {
		t.Logf("  Function: %s", fn)
		// Find nocomment baseline for this function.
		var baseFI funcInfo
		for _, fi := range baselineNo.funcs {
			if fi.name == fn {
				baseFI = fi
			}
		}
		for _, name := range names {
			for _, fi := range results[name].funcs {
				if fi.name == fn {
					refDiff := fi.refNodeSize - baseFI.refNodeSize
					fvDiff := fi.size - baseFI.size
					// Show varint byte count for the start line to connect
					// the observed data to the varint threshold table above.
					vb := varintSize(fi.spanLine)
					t.Logf("    %-14s  Span=[line %3d → %3d]  line varint: %d byte  RefNode=%d(%+d)  FuncValue=%d(%+d)",
						name, fi.spanLine, fi.spanEnd, vb,
						fi.refNodeSize, refDiff,
						fi.size, fvDiff)
				}
			}
		}
	}

	// --- Assertions ---

	// 1. Heavy comments must produce larger serialized objects.
	// commentlo may equal commentno if no line numbers cross a varint threshold.
	assert.GreaterOrEqual(t, results["commentlo"].totalObjBytes, results["commentno"].totalObjBytes,
		"commentlo should have >= serialized objects than commentno")
	assert.Greater(t, results["commenthi"].totalObjBytes, results["commentno"].totalObjBytes,
		"commenthi should have larger serialized objects than commentno")

	// 2. Per-function: since all names and paths are the same length,
	// FuncValue diff must exactly equal RefNode diff for each function.
	// This directly proves that 100% of the size overhead comes from
	// RefNode.Location.Span (line numbers shifted by comments).
	//
	// 3. Full attribution:
	//   total diff = sum(FuncValue diffs) + PkgValue diff
	// Both are purely comment-caused (no test artifacts).
	gasPerByte := DefaultGasConfig().GasGetObject

	for _, target := range []string{"commentlo", "commenthi"} {
		t.Logf("")
		t.Logf("=== Attribution: commentno → %s ===", target)
		t.Logf("")

		totalRefNodeDiff := 0
		totalFuncValueDiff := 0

		for _, fn := range funcNames {
			var noFunc, targetFunc funcInfo
			for _, fi := range results["commentno"].funcs {
				if fi.name == fn {
					noFunc = fi
				}
			}
			for _, fi := range results[target].funcs {
				if fi.name == fn {
					targetFunc = fi
				}
			}

			funcValueDiff := targetFunc.size - noFunc.size
			refNodeDiff := targetFunc.refNodeSize - noFunc.refNodeSize

			totalRefNodeDiff += refNodeDiff
			totalFuncValueDiff += funcValueDiff

			t.Logf("  %-10s  FuncValue diff: %+d  ==  RefNode diff: %+d",
				fn, funcValueDiff, refNodeDiff)

			assert.Greater(t, targetFunc.spanLine, noFunc.spanLine,
				fmt.Sprintf("[%s] func %s should have higher line number", target, fn))
			// RefNode diff may be 0 if line numbers stay within the same varint
			// byte range (e.g. both < 64). The key assertion is that FuncValue
			// diff exactly equals RefNode diff — no other field contributes.
			assert.Equal(t, funcValueDiff, refNodeDiff,
				fmt.Sprintf("[%s] func %s: FuncValue diff must exactly equal RefNode diff", target, fn))
		}

		// Full byte attribution.
		pkgValueDiff := results[target].pkgValueSize - baselineNo.pkgValueSize
		totalDiff := results[target].totalObjBytes - baselineNo.totalObjBytes
		reconstructed := totalFuncValueDiff + pkgValueDiff

		t.Logf("")
		t.Logf("  sum(FuncValue diffs)     = %+d  (all from RefNode Span line numbers)", totalFuncValueDiff)
		t.Logf("  actual total diff        = %+d", totalDiff)
		t.Logf("  gas impact               = %d bytes × GasGetObject(%d) = %+d extra gas per tx",
			totalDiff, gasPerByte, int64(totalDiff)*gasPerByte)

		assert.Equal(t, totalDiff, reconstructed,
			fmt.Sprintf("[%s] sum of FuncValue diffs + PkgValue diff must equal total object size diff", target))
	}

	// Final summary: the runtime gas overhead per tx, fully attributed.
	t.Logf("")
	t.Logf("=== Summary: runtime gas overhead from comments (issue #4919) ===")
	t.Logf("  (gas = totalObjBytes × GasGetObject(%d/byte), charged on every store read)", gasPerByte)
	t.Logf("")
	for _, name := range names {
		r := results[name]
		diff := r.totalObjBytes - baselineNo.totalObjBytes
		gasDiff := r.gasGetObject - baselineNo.gasGetObject
		t.Logf("  %-14s  total=%5d bytes  diff=%-+5d  gasDiff=%-+6d",
			name, r.totalObjBytes, diff, gasDiff)
	}
	t.Logf("")
	t.Logf("  Root cause: comments shift line numbers → larger varint in RefNode.Location.Span")
	t.Logf("    → larger serialized FuncValue → more gas on store read (GasGetObject)")
}
