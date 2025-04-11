package utils

import (
	"std"
	"strings"
)

func strSpliter(full string, sep string) (string, string) {
	i := strings.Index(full, sep)
	if i == -1 {
		return full, ""
	}
	return full[:i], full[i+1:]
}

func ParsePath(fullpath string) (string, map[string]string) {
	params := map[string]string{}

	path, paramsStr := strSpliter(fullpath, "?")
	paramList := strings.Split(paramsStr, "&")
	for _, v := range paramList {
		x, y := strSpliter(v, "=")
		params[x] = y
	}
	return path, params
}

func ShowRaw(data string) string {
	return "```\n" + data + "\n\n```\n\n"
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
