package mock

import (
	"testing"
	"unicode/utf8"
)

func TestMock(t *testing.T) {
	println("ss")
}

func FuzzMock(f *testing.F) {
	f.Add("apple hello", int(40000213132312))
	f.Add("rainy day", int(9840113223133111))
	f.Add("winter comes", int(1234912312331))
	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		rev := testing.Reverse1(v)
		doubleRev := testing.Reverse1(rev)
		u, ok2 := orig[1].(int)
		if !ok2 {
			panic("dont match")
		}
		if v != doubleRev && u > 300 && u < 500 {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) && u > 300 && u < 1000 {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}
