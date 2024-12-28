package mock

import (
	"testing"
	"testing/fuzz"
	"testing/fuzzinglib"
	"unicode/utf8"
)

func main() {
	f := fuzz.New_F()

	f.Add("apple")
	f.Add("banana")

	f.Fuzz(func(t *testing.T, orig ...interface{}) {
		v, ok := orig[0].(string)
		if !ok {
			panic("dont match")
		}
		rev := fuzzinglib.Reverse1(v)
		doubleRev := fuzzinglib.Reverse1(rev)
		if v != doubleRev {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(v) && !utf8.ValidString(rev) {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}
