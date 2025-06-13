package whatever_test

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/r/whatever"
)

func TestWhateverTest(t *testing.T) {
	whatever := whatever.Whatever{}
	uassert.Equal(t, "whatever", whatever.Whatever())
	println("whatever")
}
