package gnolang

import (
	"testing"
	"tinygo.org/x/go-llvm"
)

func TestCompileJITFactoriel(t *testing.T) {
	var expect uint64 = 10 * 9 * 8 * 7 * 6 * 5 * 4 * 3 * 2 * 1

	engine, fac := compileFactoriel()
	exec_args := []llvm.GenericValue{llvm.NewGenericValueFromInt(llvm.Int32Type(), 10, false)}
	exec_res := engine.RunFunction(fac, exec_args)

	got := exec_res.Int(false)

	if got != expect {
		t.Errorf("expected %v got %v\n", got, expect)
	}
}

func BenchmarkJITFactoriel(b *testing.B) {
	engine, fac := compileFactoriel()

	for i := 0; i < b.N; i++ {
		exec_args := []llvm.GenericValue{llvm.NewGenericValueFromInt(llvm.Int32Type(), uint64(i), false)}
		exec_res := engine.RunFunction(fac, exec_args)

		_ = exec_res.Int(false)
	}
}

func factorial(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1)
}

func BenchmarkGOFactorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = factorial(uint64(i))
	}
}
