package avl_pager_params

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestRenderEscapesUnknownKey(t *testing.T) {
	got := Render("[click](https://evil.example)")
	uassert.False(t, strings.Contains(got, "[click](https://evil.example)"), "an unknown key must not render as a link")
}
