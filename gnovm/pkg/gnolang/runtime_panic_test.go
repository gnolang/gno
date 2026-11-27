package gnolang

import "testing"

func recoverRuntime(t *testing.T, fn func()) string {
	t.Helper()
	var got string
	func() {
		defer func() {
			r := recover()
			ex, ok := r.(*Exception)
			if !ok {
				t.Fatalf("panic type %T (%v), want *Exception", r, r)
			}
			sv, ok := ex.Value.V.(StringValue)
			if !ok {
				t.Fatalf("panic value %T, want StringValue", ex.Value.V)
			}
			got = string(sv)
		}()
		fn()
	}()
	if got == "" {
		t.Fatal("did not panic")
	}
	return got
}

func TestRuntimePanicWording(t *testing.T) {
	intType := IntType
	bigdecType := UntypedBigdecType

	if msg := zeroDivideException(intType).Value.V.(StringValue); string(msg) != runtimeIntegerDivZero {
		t.Fatalf("int div: %s", msg)
	}
	if msg := zeroDivideException(bigdecType).Value.V.(StringValue); string(msg) != bigdecDivZero {
		t.Fatalf("bigdec div: %s", msg)
	}

	cases := []struct {
		name string
		fn   func()
		want string
	}{
		{"index neg", func() { checkIndex(-1, 3) }, "runtime error: index out of range [-1]"},
		{"index big", func() { checkIndex(3, 2) }, "runtime error: index out of range [3] with length 2"},
		{"nil index 0", func() { checkIndex(0, 0) }, "runtime error: index out of range [0] with length 0"},
		{"nil index 5", func() { checkIndex(5, 0) }, "runtime error: index out of range [5] with length 0"},
		{"slice lo>hi", func() { checkSlice2(2, 1, 3, boundCapacity) }, "runtime error: slice bounds out of range [2:1]"},
		{"slice neg lo", func() { checkSlice2(-1, 2, 3, boundCapacity) }, "runtime error: slice bounds out of range [-1:]"},
		{"slice neg hi", func() { checkSlice2(0, -1, 3, boundCapacity) }, "runtime error: slice bounds out of range [:-1]"},
		{"slice both neg", func() { checkSlice2(-2, -1, 3, boundCapacity) }, "runtime error: slice bounds out of range [:-1]"},
		{"slice hi>cap", func() { checkSlice2(1, 5, 3, boundCapacity) }, "runtime error: slice bounds out of range [:5] with capacity 3"},
		{"string hi", func() { checkSlice2(0, 5, 3, boundLength) }, "runtime error: slice bounds out of range [:5] with length 3"},
		{"array hi", func() { checkSlice2(0, 5, 3, boundLength) }, "runtime error: slice bounds out of range [:5] with length 3"},
		{"nil slice 0:1", func() { checkSlice2(0, 1, 0, boundCapacity) }, "runtime error: slice bounds out of range [:1] with capacity 0"},
		{"3 lo>hi", func() { checkSlice3(2, 1, 3, 3, boundCapacity) }, "runtime error: slice bounds out of range [2:1:]"},
		{"3 hi>max", func() { checkSlice3(0, 5, 3, 3, boundCapacity) }, "runtime error: slice bounds out of range [:5:3]"},
		{"3 max>cap", func() { checkSlice3(0, 1, 10, 3, boundCapacity) }, "runtime error: slice bounds out of range [::10] with capacity 3"},
		{"3 max>len array", func() { checkSlice3(0, 1, 10, 3, boundLength) }, "runtime error: slice bounds out of range [::10] with length 3"},
		{"3 neg lo", func() { checkSlice3(-1, 1, 2, 3, boundCapacity) }, "runtime error: slice bounds out of range [-1::]"},
		{"3 neg hi", func() { checkSlice3(0, -1, 2, 3, boundCapacity) }, "runtime error: slice bounds out of range [:-1:]"},
		{"3 neg max", func() { checkSlice3(0, 0, -1, 3, boundCapacity) }, "runtime error: slice bounds out of range [::-1]"},
		{"3 hi and max over cap", func() { checkSlice3(0, 4, 6, 5, boundCapacity) }, "runtime error: slice bounds out of range [::6] with capacity 5"},
		{"3 lo>hi and hi>max", func() { checkSlice3(2, 1, 0, 5, boundCapacity) }, "runtime error: slice bounds out of range [:1:0]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := recoverRuntime(t, tc.fn); got != tc.want {
				t.Fatalf("got %q\nwant %q", got, tc.want)
			}
		})
	}

	// In-range checks return.
	checkIndex(1, 3)
	checkSlice2(1, 2, 3, boundCapacity)
	checkSlice3(0, 1, 2, 3, boundCapacity)
}
