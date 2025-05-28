package pokemon

import (
	"std"
	"testing"
)

func TeststringSliceContains(t *testing.T) {
	s := []string{"foo", "bar", "baz"}

	found := stringSliceContains(s, func(i int, e string) bool {
		return e == "bar"
	})
	if !found {
		t.Errorf("Expected to find 'bar' in slice")
	}

	notFound := stringSliceContains(s, func(i int, e string) bool {
		return e == "qux"
	})
	if notFound {
		t.Errorf("Did not expect to find 'qux' in slice")
	}
}

func TestaddressSliceContains(t *testing.T) {
	s := []std.Address{
		"g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv",
		"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa",
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq",
	}

	found := addressSliceContains(s, func(i int, e std.Address) bool {
		return e == "g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv"
	})
	if !found {
		t.Errorf("Expected to find Address with 2 in slice")
	}

	notFound := addressSliceContains(s, func(i int, e std.Address) bool {
		return e == "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa"
	})
	if notFound {
		t.Errorf("Did not expect to find Address with 99 in slice")
	}
}

func TestaddressSliceIndex(t *testing.T) {
	s := []std.Address{
		"g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv",
		"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa",
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq",
	}

	idx := addressSliceIndex(s, func(i int, e std.Address) bool {
		return e == "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa"
	})
	if idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}

	notFoundIdx := addressSliceIndex(s, func(i int, e std.Address) bool {
		return e == "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa"
	})
	if notFoundIdx != -1 {
		t.Errorf("Expected index -1, got %d", notFoundIdx)
	}
}
