package main

import (
	"strconv"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	zeros         = "00000000000000" // 14 zeros
)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.CreateBoard(cross, formatName(0), true, true)
}

func formatName(i int) string {
	s := strconv.Itoa(i)
	return "b" + zeros[len(s):] + s
}

// Gas:
// 38342056
