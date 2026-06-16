// recover25b — OOB on a byte-array (the av.Data != nil branch of
// ArrayValue.GetPointerAtIndexInt2) must be recoverable, just like recover25.

package main

var a [4]byte

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	i := 7
	_ = a[i]
}

// Output:
// recover: runtime error: index out of range [7] with length 4
