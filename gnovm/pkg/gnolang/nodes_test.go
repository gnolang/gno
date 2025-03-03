package gnolang_test

import (
	"math"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestStaticBlock_Define2_MaxNames(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			panicString, ok := r.(string)
			if !ok {
				t.Errorf("expected panic string, got %v", r)
			}

			if panicString != "too many variables in block" {
				t.Errorf("expected panic string to be 'too many variables in block', got '%s'", panicString)
			}

			return
		}

		// If it didn't panic, fail.
		t.Errorf("expected panic when exceeding maximum number of names")
	}()

	staticBlock := new(gnolang.StaticBlock)
	staticBlock.NumNames = math.MaxUint16 - 1
	staticBlock.Names = make([]gnolang.Name, staticBlock.NumNames)

	// Adding one more is okay.
	staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType})
	if staticBlock.NumNames != math.MaxUint16 {
		t.Errorf("expected NumNames to be %d, got %d", math.MaxUint16, staticBlock.NumNames)
	}
	if len(staticBlock.Names) != math.MaxUint16 {
		t.Errorf("expected len(Names) to be %d, got %d", math.MaxUint16, len(staticBlock.Names))
	}

	// This one should panic because the maximum number of names has been reached.
	staticBlock.Define2(false, gnolang.Name("a"), gnolang.BoolType, gnolang.TypedValue{T: gnolang.BoolType})
}

func TestAttributesSetGetDel(t *testing.T) {
	attrs := new(gnolang.Attributes)
	key := gnolang.ATTR_IOTA
	if got, want := attrs.GetAttribute(key), (any)(nil); got != want {
		t.Errorf(".Get returned an unexpected value=%v, want=%v", got, want)
	}
	attrs.SetAttribute(key, 10)
	if got, want := attrs.GetAttribute(key), 10; got != want {
		t.Errorf(".Get returned an unexpected value=%v, want=%v", got, want)
	}
	attrs.SetAttribute(key, 20)
	if got, want := attrs.GetAttribute(key), 20; got != want {
		t.Errorf(".Get returned an unexpected value=%v, want=%v", got, want)
	}
	attrs.DelAttribute(key)
	if got, want := attrs.GetAttribute(key), (any)(nil); got != want {
		t.Errorf(".Get returned an unexpected value=%v, want=%v", got, want)
	}
}

var sink any = nil

func BenchmarkAttributesSetGetDel(b *testing.B) {
	n := 100
	keys := make([]gnolang.GnoAttribute, 0, n)
	for i := 0; i < n; i++ {
		keys = append(keys, gnolang.GnoAttribute(i))
	}
	attrCommon := gnolang.ATTR_TYPEOF_VALUE

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		attrs := new(gnolang.Attributes)
		for j := 0; j < 100; j++ {
			sink = attrs.GetAttribute(attrCommon)
		}
		for j := 0; j < 100; j++ {
			attrs.SetAttribute(attrCommon, j)
			sink = attrs.GetAttribute(attrCommon)
		}

		for j, key := range keys {
			attrs.SetAttribute(key, j)
		}

		for _, key := range keys {
			sink = attrs.GetAttribute(key)
		}

		sink = attrs
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
}
