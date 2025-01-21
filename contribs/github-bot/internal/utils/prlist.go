package utils

import (
	"encoding"
	"fmt"
	"strconv"
	"strings"
)

// Type used to (un)marshal input/output for check and matrix subcommands.
type PRList []int

// PRList is both a TextMarshaler and a TextUnmarshaler.
var (
	_ encoding.TextMarshaler   = PRList{}
	_ encoding.TextUnmarshaler = &PRList{}
)

// MarshalText implements encoding.TextMarshaler.
func (p PRList) MarshalText() (text []byte, err error) {
	prNumsStr := make([]string, len(p))

	for i, prNum := range p {
		prNumsStr[i] = strconv.Itoa(prNum)
	}

	return []byte(strings.Join(prNumsStr, ", ")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *PRList) UnmarshalText(text []byte) error {
	prNumsStr := strings.Split(string(text), ",")
	prNums := make([]int, len(prNumsStr))

	for i := range prNumsStr {
		prNum, err := strconv.Atoi(strings.TrimSpace(prNumsStr[i]))
		if err != nil {
			return err
		}

		if prNum <= 0 {
			return fmt.Errorf("invalid pull request number (<= 0): original(%s) parsed(%d)", prNumsStr[i], prNum)
		}

		prNums[i] = prNum
	}
	*p = prNums

	return nil
}
