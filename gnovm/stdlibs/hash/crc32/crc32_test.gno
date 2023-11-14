package crc32_test

import (
	"hash/crc32"
	"testing"
)

func TestCastagnoli(t *testing.T) {
	ieee := NewIEEE()
	MakeTable(Castagnoli)
	ieee.Write([]byte("hello"))
}
