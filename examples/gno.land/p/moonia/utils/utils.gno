package utils

import (
	"std"
	"strconv"
	"time"
)

func GetBlockTime() int64 {
	return time.Now().Unix()
}

func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("02 Jan 2006, 15:04")
}

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