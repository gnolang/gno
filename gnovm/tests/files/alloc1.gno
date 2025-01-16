// MAXALLOC: 25000
// max total allocation of 100 mb.
package main

import "runtime"

// fileBlock: 496 + n*40 = 536
var c = 1
var b = "hello" // rhs is const
// map: 176 + n*120
var m = map[string]int{"a": 1}

type Foo struct{ name string }

// after new block: 3483
func bar() {
	f := &Foo{"foo"} // + field 40 + struct 184 + pointer 24 + heapItem 72 + NewType: *main.Foo 232 = 4305
	f2 := f
	println(f)                                  // +536 + 72 + 248 = 4891
	bb := &Foo{"foo2"}                          // + 40 + 184 + pointer 24 + heapItem72 + new type 232 = 5443
	println("MemStats 2: ", runtime.MemStats()) // 5979 + 536 + new string 67 = 6046
}

// before this, declaration is 536 + 176 + 120 = 832,
// + new block: 536
func main() {
	a := 1
	println("MemStats 1: ", runtime.MemStats())         // + 536 + 536 + NewString: 67 + (list)72 + 288 = 2867
	bar()                                               // +616
	println("MemStats before GC: ", runtime.MemStats()) // 6046 + 536 + 72 + 288 + ... 7478(the moment memstas result generated)
	runtime.GC()
	println("MemStats after GC: ", runtime.MemStats())

	f := func() {
		println("func lit")
	}
	f()
	runtime.GC()
}

// Output:
// MemStats 1:  Allocator{maxBytes:10000000000, bytes:1944}
// &(struct{("foo" string)} main.Foo)
// MemStats 2:  Allocator{maxBytes:10000000000, bytes:6019}
// MemStats before GC:  Allocator{maxBytes:10000000000, bytes:7518}
// MemStats after GC:  Allocator{maxBytes:10000000000, bytes:2901}
// func lit
