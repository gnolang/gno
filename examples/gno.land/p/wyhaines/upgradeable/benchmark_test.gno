package upgradeable

import (
	"testing"
	
	"gno.land/p/demo/testutils"
)

// BenchmarkDirectFunctionCall benchmarks a direct function call
func BenchmarkDirectFunctionCall(b *testing.B) {
	fn := func(s string) string {
		return "prefix: " + s
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn("test")
	}
}

// BenchmarkRegistryGet benchmarks retrieving a function from the registry
func BenchmarkRegistryGet(b *testing.B) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	
	registry := New()
	fn := func(s string) string {
		return "prefix: " + s
	}
	
	_ = registry.RegisterFunction("test", fn)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetFunction("test")
	}
}

// BenchmarkFunctionHolderGet benchmarks getting a function from the holder
func BenchmarkFunctionHolderGet(b *testing.B) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	
	registry := New()
	fn := func(s string) string {
		return "prefix: " + s
	}
	
	holder := NewFunctionHolder(registry, "test", fn)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = holder.Get()
	}
}

// BenchmarkStringFuncHolderGet benchmarks getting from a specialized holder
func BenchmarkStringFuncHolderGet(b *testing.B) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	
	registry := New()
	fn := func(s string) string {
		return "prefix: " + s
	}
	
	holder := NewStringFuncHolder(registry, "test", fn)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = holder.Get()
	}
}

// BenchmarkStringFuncHolderGetAndCall benchmarks getting and calling a function
func BenchmarkStringFuncHolderGetAndCall(b *testing.B) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	
	registry := New()
	fn := func(s string) string {
		return "prefix: " + s
	}
	
	holder := NewStringFuncHolder(registry, "test", fn)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := holder.Get()
		_ = f("test")
	}
}

// BenchmarkFunctionUpgrade benchmarks upgrading a function
func BenchmarkFunctionUpgrade(b *testing.B) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)
	
	registry := New()
	fn1 := func(s string) string {
		return "prefix1: " + s
	}
	fn2 := func(s string) string {
		return "prefix2: " + s
	}
	
	holder := NewStringFuncHolder(registry, "test", fn1)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between two functions to simulate upgrades
		if i%2 == 0 {
			_ = holder.Update(fn2)
		} else {
			_ = holder.Update(fn1)
		}
	}
}
