package gnolang

import (
	"strings"
	"testing"
)

func TestHashlet_UnmarshalAmino(t *testing.T) {
	var pid PkgID
	var vh ValueHash
	decoders := map[string]func(string) error{
		"PkgID":     pid.UnmarshalAmino,
		"ValueHash": vh.UnmarshalAmino,
	}
	for name, decode := range decoders {
		for _, h := range []string{
			"abcd",
			strings.Repeat("ab", 21), // one byte past the hash
		} {
			err := decode(h)
			if err == nil || !strings.Contains(err.Error(), "invalid hash length") {
				t.Errorf("%s %q: got error %v, want invalid hash length", name, h, err)
			}
		}
		if err := decode(strings.Repeat("ab", HashSize)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
