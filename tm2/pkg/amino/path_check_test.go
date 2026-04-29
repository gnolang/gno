package amino_test

import (
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// TestAllTypesUseGenproto2 verifies that Marshal/Unmarshal for every
// registered genproto2 type actually takes the genproto2 path, not
// silently falling through to reflect.
func TestAllTypesUseGenproto2(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	allTypes := append([]any{}, tests.StructTypes...)
	allTypes = append(allTypes, tests.AminoTagTypes...)

	for _, ptr := range allTypes {
		rt := reflect.TypeOf(ptr).Elem()
		name := rt.Name()

		// Check if this type has genproto2 methods.
		rv := reflect.New(rt)
		ptrVal := rv.Interface()
		if _, ok := ptrVal.(amino.PBMessager2); !ok {
			continue // skip non-genproto2 types
		}
		if !amino.HasNativeGenproto2(rv.Type()) {
			continue
		}

		t.Run(name, func(t *testing.T) {
			// Marshal via *T
			encBefore := atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)
			bz, err := cdc.Marshal(ptrVal)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			encAfter := atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)
			if encAfter != encBefore+1 {
				t.Errorf("Marshal(*%s) did NOT use genproto2 (enc counter: %d → %d)", name, encBefore, encAfter)
			}

			// Marshal via bare T
			encBefore = atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)
			bz2, err := cdc.Marshal(rv.Elem().Interface())
			if err != nil {
				t.Fatalf("Marshal(bare %s) failed: %v", name, err)
			}
			encAfter = atomic.LoadInt64(&cdc.GetStats().Genproto2Encodes)
			if encAfter != encBefore+1 {
				t.Errorf("Marshal(%s) bare value did NOT use genproto2 (enc counter: %d → %d)", name, encBefore, encAfter)
			}
			_ = bz2

			// Unmarshal via *T
			decBefore := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)
			dst := reflect.New(rt).Interface()
			err = cdc.Unmarshal(bz, dst)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			decAfter := atomic.LoadInt64(&cdc.GetStats().Genproto2Decodes)
			if decAfter != decBefore+1 {
				t.Errorf("Unmarshal(*%s) did NOT use genproto2 (dec counter: %d → %d)", name, decBefore, decAfter)
			}
		})
	}
}
