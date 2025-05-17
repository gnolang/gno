package utils

import (
	"std"
	"testing"
)

func TestShowRaw(t *testing.T) {
	data := "hello world"
	expected := "\n\n```\n" + data + "\n```\n\n"

	result := ShowRaw(data)
	if result != expected {
		t.Fatalf("Expected %q, got %q", expected, result)
	}
}

func TestIntSliceContains(t *testing.T) {
	s := []int{1, 2, 3, 4}

	found := IntSliceContains(s, func(i, e int) bool {
		return e == 3
	})
	if !found {
		t.Errorf("Expected to find 3 in slice")
	}

	notFound := IntSliceContains(s, func(i, e int) bool {
		return e == 10
	})
	if notFound {
		t.Errorf("Did not expect to find 10 in slice")
	}
}

func TestStringSliceContains(t *testing.T) {
	s := []string{"foo", "bar", "baz"}

	found := StringSliceContains(s, func(i int, e string) bool {
		return e == "bar"
	})
	if !found {
		t.Errorf("Expected to find 'bar' in slice")
	}

	notFound := StringSliceContains(s, func(i int, e string) bool {
		return e == "qux"
	})
	if notFound {
		t.Errorf("Did not expect to find 'qux' in slice")
	}
}

func TestAddressSliceContains(t *testing.T) {
	s := []std.Address{
		"g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv",
		"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa",
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq",
	}

	found := AddressSliceContains(s, func(i int, e std.Address) bool {
		return e == "g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv"
	})
	if !found {
		t.Errorf("Expected to find Address with 2 in slice")
	}

	notFound := AddressSliceContains(s, func(i int, e std.Address) bool {
		return e == "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa"
	})
	if notFound {
		t.Errorf("Did not expect to find Address with 99 in slice")
	}
}

func TestAddressSliceIndex(t *testing.T) {
	s := []std.Address{
		"g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv",
		"g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa",
		"g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq",
	}

	idx := AddressSliceIndex(s, func(i int, e std.Address) bool {
		return e == "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa"
	})
	if idx != 1 {
		t.Errorf("Expected index 1, got %d", idx)
	}

	notFoundIdx := AddressSliceIndex(s, func(i int, e std.Address) bool {
		return e == "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa"
	})
	if notFoundIdx != -1 {
		t.Errorf("Expected index -1, got %d", notFoundIdx)
	}
}
