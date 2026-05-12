package main

// Regression test for iterative exception recovery in Machine.Run().
// Before the fix, each panicking defer added a Go stack frame via recursive
// m.Run(st), and ~500K defers would exceed the 1GB goroutine stack limit,
// crashing the process with runtime.throw("stack overflow").
// The iterative recovery loop handles this in O(1) Go stack frames.

type S struct {
	X int
}

func main() {
	defer func() {
		// Recover the final panic after all 500K panicking defers complete.
		println(recover())
	}()
	var p *S
	for i := 0; i < 500000; i++ {
		defer func() {
			// Nil pointer dereference triggers a Go-level panic(&Exception{}),
			// not the cooperative pushPanic path.
			_ = p.X
		}()
	}
}

// Output:
// runtime error: nil pointer dereference
