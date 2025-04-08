package utils

import (
	"std"
	"strconv"
	"strings"
)

func GetCaller() std.Address {
	return std.PreviousRealm().Address()
}

func Parse(fullpath string) (string, map[string]string) {
	parts := strings.Split(fullpath, ":")
	if len(parts) == 2 {
		return parts[1], map[string]string{}
	}
	return "", map[string]string{}
}

func ParseIndex(indexStr string, max int) int {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		panic("Invalid index.")
	}
	if index < 0 || index >= max {
		panic("Proposal does not exist.")
	}
	return index
}