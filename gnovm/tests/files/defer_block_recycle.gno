package main

// Regression test for block recycling vs pending defers.
//
// A deferred call resolves its scope from its FuncValue (parent + copied
// heap-item captures), never from the block it was registered in. So that
// origin block is dead once popped and is recycled through the machine's
// block pool — possibly zeroed and reused for an unrelated scope before the
// defer fires. These cases assert that recycling never corrupts a pending
// defer's captured state, named-return storage, or execution order.

// Named return lives in the function's call block — the very block recycled
// at return. The deferred closure must still mutate the value that is returned.
func namedRet() (r int) {
	defer func() { r++ }()
	return 5
}

// Per-iteration defers in inner block scopes that pop early, then heavy
// allocation churn to force pool reuse before the defers run at return.
func recycleChurn() {
	for i := 0; i < 3; i++ {
		{
			v := i * 10
			defer func() { println("churn defer:", v) }()
		}
	}
	s := 0
	for j := 0; j < 200; j++ {
		buf := make([]int, 8)
		buf[0] = j
		s += buf[0]
	}
	println("churn sum:", s)
}

// continue releases the loop-body blocks (releaseBlocksFrom) while a defer
// registered in that body is still pending.
func continueDefer() {
	for i := 0; i < 3; i++ {
		w := i + 1
		defer func() { println("continue defer:", w) }()
		if i == 1 {
			continue
		}
		_ = w
	}
}

// A deferred closure that itself registers a defer (the defer-exec block is
// itself acquired from the pool).
func nestedDefer() {
	defer func() {
		println("nested outer")
		defer func() { println("nested inner") }()
	}()
	println("nested body")
}

func main() {
	println("namedRet:", namedRet())
	recycleChurn()
	continueDefer()
	nestedDefer()
}

// Output:
// namedRet: 6
// churn sum: 19900
// churn defer: 20
// churn defer: 10
// churn defer: 0
// continue defer: 3
// continue defer: 2
// continue defer: 1
// nested body
// nested outer
// nested inner
