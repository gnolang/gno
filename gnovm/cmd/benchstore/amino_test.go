//go:build genproto2

package benchstore

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// BenchmarkAminoMarshalReflect benchmarks the reflection-based amino encoding path.
func BenchmarkAminoMarshalReflect(b *testing.B) {
	tvs := BuildTestValues()
	sort.Slice(tvs, func(i, j int) bool { return len(tvs[i].Bytes) < len(tvs[j].Bytes) })

	for _, tv := range tvs {
		bz, err := amino.MarshalReflect(tv.Value)
		if err != nil {
			b.Fatalf("MarshalReflect(%s): %v", tv.Name, err)
		}
		b.Run(fmt.Sprintf("%04dB/%s", len(bz), tv.Name), func(b *testing.B) {
			v := tv.Value
			b.SetBytes(int64(len(bz)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				amino.MarshalReflect(v)
			}
		})
	}
}

// BenchmarkAminoMarshalBinary2 benchmarks the genproto2 direct-encoding path.
func BenchmarkAminoMarshalBinary2(b *testing.B) {
	tvs := BuildTestValues()
	sort.Slice(tvs, func(i, j int) bool { return len(tvs[i].Bytes) < len(tvs[j].Bytes) })

	for _, tv := range tvs {
		pbm2, ok := tv.Value.(amino.PBMessager2)
		if !ok {
			continue
		}
		bz, err := amino.MarshalBinary2(pbm2)
		if err != nil {
			b.Fatalf("MarshalBinary2(%s): %v", tv.Name, err)
		}
		b.Run(fmt.Sprintf("%04dB/%s", len(bz), tv.Name), func(b *testing.B) {
			b.SetBytes(int64(len(bz)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				amino.MarshalBinary2(pbm2)
			}
		})
	}
}

// newZeroPtr returns a pointer to a fresh zero value of the same concrete type as v.
func newZeroPtr(v interface{}) interface{} {
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	return reflect.New(rt).Interface()
}

// BenchmarkAminoUnmarshalReflect benchmarks the reflection-based amino decoding path.
func BenchmarkAminoUnmarshalReflect(b *testing.B) {
	tvs := BuildTestValues()
	sort.Slice(tvs, func(i, j int) bool { return len(tvs[i].Bytes) < len(tvs[j].Bytes) })

	for _, tv := range tvs {
		bz, err := amino.MarshalReflect(tv.Value)
		if err != nil {
			b.Fatalf("MarshalReflect(%s): %v", tv.Name, err)
		}
		b.Run(fmt.Sprintf("%04dB/%s", len(bz), tv.Name), func(b *testing.B) {
			b.SetBytes(int64(len(bz)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ptr := newZeroPtr(tv.Value)
				amino.UnmarshalReflect(bz, ptr)
			}
		})
	}
}

// BenchmarkAminoUnmarshalBinary2 benchmarks the genproto2 direct-decoding path.
func BenchmarkAminoUnmarshalBinary2(b *testing.B) {
	tvs := BuildTestValues()
	sort.Slice(tvs, func(i, j int) bool { return len(tvs[i].Bytes) < len(tvs[j].Bytes) })

	for _, tv := range tvs {
		pbm2, ok := tv.Value.(amino.PBMessager2)
		if !ok {
			continue
		}
		bz, err := amino.MarshalBinary2(pbm2)
		if err != nil {
			b.Fatalf("MarshalBinary2(%s): %v", tv.Name, err)
		}
		rt := reflect.TypeOf(tv.Value)
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		b.Run(fmt.Sprintf("%04dB/%s", len(bz), tv.Name), func(b *testing.B) {
			b.SetBytes(int64(len(bz)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ptr := reflect.New(rt).Interface().(amino.PBMessager2)
				ptr.UnmarshalBinary2(amino.Gcdc(), bz)
			}
		})
	}
}
