package mock

import (
	"testing"
	"time"
	"unicode/utf8"
)

func mock() {
	f := testing.New_F(1_000_000)
	var u1 int = 9670000
	var u2 int = -3013000
	var u3 int = 1312000
	f.Add("apple hello", u1)
	f.Add("rainy day", u2)
	f.Add("winter comes", u3)

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
