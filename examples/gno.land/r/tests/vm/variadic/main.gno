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
