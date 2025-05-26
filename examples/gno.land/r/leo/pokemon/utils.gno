package pokemon

import (
	"std"
)

func ShowRaw(data string) string {
	return "\n\n```\n" + data + "\n```\n\n"
}

func IntSliceContains(s []int, f func(int, int) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func StringSliceContains(s []string, f func(int, string) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func AddressSliceContains(s []std.Address, f func(int, std.Address) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func AddressSliceIndex(s []std.Address, f func(int, std.Address) bool) int {
	for i, e := range s {
		if f(i, e) {
			return i
		}
	}
	return -1
}
