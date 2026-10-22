// MAXALLOC: 100000000
package main

import "runtime"

// doOpDefer must pop the deferred call's operands on every branch. The
// nil-func and non-func branches left them on the operand stack until the
// frame returned, and GC counts the operand stack, so MemStats grew by 40 B
// per leftover slot.

// Each defer left its two int operands behind.
func nilFunc() {
	defer func() { recover() }()
	var nf func(int, int)
	defer nf(1, 2)
	defer nf(1, 2)
	defer nf(1, 2)
	runtime.GC()
	println("nil func:", runtime.MemStats())
}

// The type checker rejects this; paths that skip it, such as filetests
// and vm/qeval queries, still get here. The conversion and its operand were
// left behind, still on the stack when the deferred closure ran.
func notFunc() {
	defer func() {
		println(recover())
		runtime.GC()
		println("not func:", runtime.MemStats())
	}()
	var x int
	defer int64(x)
}

func main() {
	nilFunc()
	notFunc()
}

// Output:
// nil func: Allocator{maxBytes:100000000, bytes:9672}
// invalid defer function call: typeval{int64}
// not func: Allocator{maxBytes:100000000, bytes:10472}

// TypeCheckError:
// main/alloc_defer_operand_stack.gno:32:8: defer requires function call, not conversion int64(x) (value of type int64)
