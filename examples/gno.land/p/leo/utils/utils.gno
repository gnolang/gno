package utils

import (
	"strings"
)

func srtSpliter(full string, sep string) (string, string) {
	i := strings.Index(full, sep)
	if i == -1 {
		return full, ""
	}
	return full[:i], full[i+1:]
}

func ParseParams(fullpath string) map[string]string {
	params := map[string]string{}

	_, paramsStr := srtSpliter(fullpath, "?")
	paramList := strings.Split(paramsStr, "&")
	for _, v := range paramList {
		x, y := srtSpliter(v, "=")
		params[x] = y
	}
	return params
}

func ShowRaw(data string) string {
	return "```\n" + data + "```\n\n"
}

func SliceContainsInt(s []int, f func(int, int) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}

func SliceContainsString(s []string, f func(int, string) bool) bool {
	for i, e := range s {
		if f(i, e) {
			return true
		}
	}
	return false
}
