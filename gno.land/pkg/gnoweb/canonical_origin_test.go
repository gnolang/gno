package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeCanonicalOrigin(t *testing.T) {
	t.Parallel()

	// Harmless spellings are fixed, not refused: a refusal stops gnoweb.
	for in, want := range map[string]string{
		"":                      "",
		"https://gno.land":      "https://gno.land",
		"http://localhost:8888": "http://localhost:8888",
		"https://gno.land/":     "https://gno.land",
		"https://gno.land//":    "https://gno.land",
		" https://gno.land \n":  "https://gno.land",
		"HTTPS://GNO.land":      "https://gno.land",
		"https://gno.land:443":  "https://gno.land",
		"http://gno.land:80":    "http://gno.land",
		"https://gno.land:8443": "https://gno.land:8443",
		"http://gno.land:443":   "http://gno.land:443",
		"http://[::1]:8888":     "http://[::1]:8888",
	} {
		got, err := normalizeCanonicalOrigin(in)
		if assert.NoError(t, err, "origin %q", in) {
			assert.Equal(t, want, got, "origin %q", in)
		}
	}

	for _, in := range []string{
		"gno.land",
		"ftp://gno.land",
		"https://",
		"https://gno.land/r/demo",
		"https://gno.land?x=1",
		"https://gno.land#top",
		"https://user@gno.land",
		"javascript:alert(1)",
		"https://gno.land\n/evil",
		"https://gno.land?",
		"https://gno.land#",
		"https:gno.land",
		"https://gno.land:",
		"https://gno.land:x:443",
		"https://gno.land:443:443",
	} {
		_, err := normalizeCanonicalOrigin(in)
		assert.Error(t, err, "origin %q", in)
	}
}
