package msha1

import (
	"std"
	"testing"

	"crypto/sha1"
	"encoding/hex"
)

func TestFunc(t *testing.T) {
	got := sha1.Sum([]byte("sha1 this string"))
	expected := "cf23df2207d99a74fbe169e3eba035e633b65d94"

	if (hex.EncodeToString(got)) != expected {
		t.Errorf("got %v(%T), expected %v(%T)", got, got, expected, expected)
	}
}