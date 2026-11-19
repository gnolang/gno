package gnolang

import "testing"

func TestException_WithPrevious_E2HasNext(t *testing.T) {
	e1, e2, e3 := &Exception{}, &Exception{}, &Exception{}
	e2.Next = e3

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when e2 already has a next link")
		}
	}()

	e1.WithPrevious(e2)
}
