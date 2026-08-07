package gnolang

import "testing"

func TestException_WithPrevious_E2HasNext(t *testing.T) {
	e1 := &Exception{Value: TypedValue{}}
	e2 := &Exception{Value: TypedValue{}}
	e3 := &Exception{Value: TypedValue{}}

	// Link e2 to e3 first
	e2.Next = e3

	// Now trying to use e2 as previous should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when e2 already has a next link")
		}
	}()

	e1.WithPrevious(e2)
}
