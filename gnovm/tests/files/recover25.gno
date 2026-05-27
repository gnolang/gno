// Issue 4353. A runtime index-out-of-range panic from indexing an array
// (or pointer-to-array) must be recoverable via a deferred recover(),
// the same as slice/string OOB and division-by-zero are (see recover14).

package main

var aib [100000]int
var paib *[100000]int = &aib
var i64 int64 = 100023

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	_ = paib[i64]
}

// Output:
// recover: runtime error: index out of range [100023] with length 100000
