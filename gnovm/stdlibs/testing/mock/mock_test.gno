package mock

import (
	"testing"
	"unicode/utf8"
)

func TestMock(t *testing.T) {
	println("ss")
}

func FuzzMock(f *testing.F) {
	f.Add("apple hello", int(400002131323))
	f.Add("rainy day", int(98401132231331))
	f.Add("winter comes", int(12349123123))
	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		i, ok2 := orig[1].(int)
		if !ok2 {
			panic("dont match")
		}
		rev := testing.Reverse1(v)
		doubleRev := testing.Reverse1(rev)
		if v != doubleRev && i > 300 && i < 500 {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) && i > 300 && i < 1000 {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}
