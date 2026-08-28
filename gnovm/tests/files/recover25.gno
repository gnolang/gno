// Go issue 4353. A runtime index-out-of-range panic from indexing an array
// (or pointer-to-array) must be recoverable via a deferred recover(),
// the same as slice/string OOB and division-by-zero are (see recover14).

package main

var aib [100000]int
var paib *[100000]int = &aib

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	i := 100023
	_ = paib[i]
}

// Output:
// recover: runtime error: index out of range [100023] with length 100000
