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

// bar emits multiple events
func bar() {
	std.Emit(
		"bar1",
		"attr1", "val1",
		"attr2", "val2",
	)

	std.Emit(
		"bar2",
		"attr1", "val1",
		"attr2", "val2",
	)

	std.Emit(
		"bar3",
		"attr1", "val1",
		"attr2", "val2",
	)
}

func TestExpectEmit(t *T) {
	foo()

	attrs := []string{"attr1", "val1", "attr2", "val2"}
	ok := ExpectEmit("foo", attrs, -1, false)
	if !ok {
		t.Fatalf("expectEmit failed")
	}
}

func TestExpectEmit_MultipleEvents(t *T) {
	bar()

	// 1st
	attrs := []string{"attr1", "val1", "attr2", "val2"}
	ok := ExpectEmit("bar1", attrs, 0, false)
	if !ok {
		t.Fatalf("expectEmit failed")
	}

	// last
	ok = ExpectEmit("bar3", attrs, -1, false)
	if !ok {
		t.Fatalf("expectEmit failed")
	}

	// 2nd
	ok = ExpectEmit("bar2", attrs, 1, false)
	if !ok {
		t.Fatalf("expectEmit failed")
	}

	// wrong index
	ok = ExpectEmit("bar1", attrs, -2, false)
	if ok {
		t.Fatalf("expectEmit should have failed")
	}

	// partial match
	partialAttrs := []string{"attr1", "val1"}
	ok = ExpectEmit("bar1", partialAttrs, 0, true)
	if !ok {
		t.Fatalf("expectEmit failed")
	}

	// wrong partial match
	wrongPartialAttrs := []string{"attr1", "wrongVal"}
	ok = ExpectEmit("bar1", wrongPartialAttrs, 0, true)
	if ok {
		t.Fatalf("expectEmit should have failed")
	}
}
