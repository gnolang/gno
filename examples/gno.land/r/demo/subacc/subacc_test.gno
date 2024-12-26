package subacc_test

import (
	"std"
	"testing"

	"gno.land/p/demo/urequire"
	"gno.land/r/demo/subacc"
)

func TestEntityID(t *testing.T) {
	test1 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // test1
	expected := "/g1u5hdzuqfln65xyx5dvz9ldl9e45pmew5exg302/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/savings"
	urequire.Equal(t, expected, subacc.EntityID(test1, "savings"))
}
