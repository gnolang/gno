package variadic

import "strings"

func Echo(cur realm, vals ...string) string {
	return strings.Join(vals, " ")
}

func Add(cur realm, nums ...int) int {
	res := 0

	for _, num := range nums {
		res += num
	}

	return res
}

func And(cur realm, booleans ...bool) bool {

	for _, boolean := range booleans {
		if !boolean {
			return false
		}
	}

	return true
}
