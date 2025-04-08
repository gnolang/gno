package utils

import (
	"std"
	"strconv"
)

func GetCaller() std.Address {
	return std.PreviousRealm().Address()
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