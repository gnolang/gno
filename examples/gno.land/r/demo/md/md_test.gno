package md

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	output := Render("")
	if !strings.Contains(output, "Exploring markdown functions in the gno.land/p/md library") {
		t.Errorf("invalid output")
	}
}
