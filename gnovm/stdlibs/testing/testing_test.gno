package testing

import "std"

// dummy function to emit an event
func foo() {
	std.Emit(
		"foo",
		"attr1", "val1",
		"attr2", "val2",
	)
}

func TestExpectEmit(t *T) {
	foo()

	attrs := []string{"attr1", "val1", "attr2", "val2"}
	ok := ExpectEmit("foo", attrs)
	if !ok {
		t.Fatalf("expectEmit failed")
	}
}
