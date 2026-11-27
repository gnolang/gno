package gnolang

import "fmt"

// Runtime panic sentences that match gc (Go 1.25, src/runtime/error.go and
// panic.go). errorString.Error adds the "runtime error: " prefix; boundsError
// does too. The numbers are the ones gc prints for the same failed check:
// a negative index omits the length, and a slice bound that fails against
// len or cap is printed alone rather than as a pair.
const (
	runtimeNilPointerDeref = "runtime error: invalid memory address or nil pointer dereference"
	runtimeIntegerDivZero  = "runtime error: integer divide by zero"
	runtimeNegShift        = "runtime error: negative shift amount"

	// bigdecDivZero is not an integer divide. gc has no runtime decimal type;
	// this stays a distinct sentence so it is not reported as one.
	bigdecDivZero = "runtime error: division by zero"

	boundLength   = "length"
	boundCapacity = "capacity"
)

func panicRuntimeError(msg string) {
	panic(&Exception{Value: typedRuntimeError(msg)})
}

func panicRuntimeErrorf(format string, args ...any) {
	panicRuntimeError(fmt.Sprintf(format, args...))
}

// zeroDivideException is the runtime panic for a zero divisor.
// Integer widths, including untyped bigint, use gc's sentence.
// Untyped bigdec keeps its own sentence: it is not an integer divide.
func zeroDivideException(t Type) *Exception {
	msg := runtimeIntegerDivZero
	if baseOf(t) == UntypedBigdecType {
		msg = bigdecDivZero
	}
	return &Exception{Value: typedRuntimeError(msg)}
}

// checkIndex panics unless 0 <= i < length, using gc's boundsIndex text.
// A negative i is "index out of range [i]" with no length; any other miss
// is "index out of range [i] with length n". A nil slice is length 0.
func checkIndex(i, length int) {
	if i >= 0 && i < length {
		return
	}
	if i < 0 {
		panicRuntimeErrorf("runtime error: index out of range [%d]", i)
	}
	panicRuntimeErrorf("runtime error: index out of range [%d] with length %d", i, length)
}

// checkSlice2 panics unless 0 <= low <= high <= limit.
// limit is a length for strings and arrays and a capacity for slices,
// which is the word gc puts in the message. The high bound is checked
// first, matching gc: s[-2:-1] reports the high bound.
func checkSlice2(low, high, limit int, word string) {
	if high < 0 {
		panicRuntimeErrorf("runtime error: slice bounds out of range [:%d]", high)
	}
	if high > limit {
		panicRuntimeErrorf("runtime error: slice bounds out of range [:%d] with %s %d", high, word, limit)
	}
	if low < 0 {
		panicRuntimeErrorf("runtime error: slice bounds out of range [%d:]", low)
	}
	if low > high {
		panicRuntimeErrorf("runtime error: slice bounds out of range [%d:%d]", low, high)
	}
}

// checkSlice3 panics unless 0 <= low <= high <= maxBound <= limit
// for a three-index slice. word is "length" for arrays and "capacity"
// for slices, matching gc. maxBound is checked first, then high, then low.
func checkSlice3(low, high, maxBound, limit int, word string) {
	if maxBound < 0 {
		panicRuntimeErrorf("runtime error: slice bounds out of range [::%d]", maxBound)
	}
	if maxBound > limit {
		panicRuntimeErrorf("runtime error: slice bounds out of range [::%d] with %s %d", maxBound, word, limit)
	}
	if high < 0 {
		panicRuntimeErrorf("runtime error: slice bounds out of range [:%d:]", high)
	}
	if high > maxBound {
		panicRuntimeErrorf("runtime error: slice bounds out of range [:%d:%d]", high, maxBound)
	}
	if low < 0 {
		panicRuntimeErrorf("runtime error: slice bounds out of range [%d::]", low)
	}
	if low > high {
		panicRuntimeErrorf("runtime error: slice bounds out of range [%d:%d:]", low, high)
	}
}
