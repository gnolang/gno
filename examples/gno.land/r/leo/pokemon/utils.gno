package pokemon

import (
	"std"
)

func stringSliceContains(s []string, f func(int, string) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func addressSliceContains(s []std.Address, f func(int, std.Address) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func addressSliceIndex(s []std.Address, f func(int, std.Address) bool) int {
	for i, e := range s {
		if f(i, e) {
			return i
		}
	}
	return -1
}
