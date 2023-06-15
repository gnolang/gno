package showcase

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	cases := []struct {
		path           string
		shouldContains string
	}{
		{"", "Package showcase provides a framework"},
		{"a-404-path", "404"},
	}

	for _, tc := range cases {
		name := strings.ReplaceAll(tc.path, "/", "~")
		t.Run(name, func(t *testing.T) {
			result := Render(tc.path)
			println("result", result)
			println("shouldContain", tc.shouldContains)
			if !strings.Contains(result, tc.shouldContains) {
				t.Errorf("expected %q\ngot:      %q", result)
			}
		})
	}
}
