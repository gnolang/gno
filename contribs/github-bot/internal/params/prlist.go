package params

import (
	"encoding"
	"fmt"
	"strconv"
	"strings"
)

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

	return []byte(strings.Join(prNumsStr, ",")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *PRList) UnmarshalText(text []byte) error {
	for _, prNumStr := range strings.Split(string(text), ",") {
		prNum, err := strconv.Atoi(strings.TrimSpace(prNumStr))
		if err != nil {
			return err
		}

		if prNum <= 0 {
			return fmt.Errorf("invalid pull request number (<= 0): original(%s) parsed(%d)", prNumStr, prNum)
		}

		*p = append(*p, prNum)
	}

	return nil
}
