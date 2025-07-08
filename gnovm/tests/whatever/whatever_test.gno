package whatever

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestWhatever(t *testing.T) {
	whatever := Whatever{}
	uassert.Equal(t, "whatever", whatever.Whatever())
	println("whatever")
}
