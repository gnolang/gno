package mock

import (
	"testing"
	"testing/fuzz"
	"testing/fuzzinglib"
	"unicode/utf8"
)

func main() {
	f := fuzz.New_F()
	var u1 uint = 30
	// var u2 uint16 = 20
	var u2 uint = 30
	var u3 uint = 30
	f.Add("apple", u1)
	f.Add("hello", u2)
	f.Add("kkk", u3)

	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		rev := fuzzinglib.Reverse1(v)
		doubleRev := fuzzinglib.Reverse1(rev)
		u, ok2 := orig[1].(uint)
		if !ok2 {
			panic("dont match")
		}
		if v != doubleRev && u > 300 && u < 1000 {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) && u == 300 {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}
