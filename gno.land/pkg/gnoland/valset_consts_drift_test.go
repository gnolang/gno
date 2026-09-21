package gnoland

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestValsetConstsDoNotDrift asserts the param-key string values in
// examples/gno.land/r/sys/params/valset.gno match the keys EndBlocker
// reads on the Go side. If they drift, valset rotation silently breaks
// at runtime with no compile/test error from either side individually.
func TestValsetConstsDoNotDrift(t *testing.T) {
	t.Parallel()

	// `go test` runs from the package directory: gno.land/pkg/gnoland.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..", "..")
	gnoPath := filepath.Join(root, "examples", "gno.land", "r", "sys", "params", "valset.gno")

	data, err := os.ReadFile(gnoPath)
	if err != nil {
		t.Fatalf("read %s: %v", gnoPath, err)
	}

	// Build the prefix the gno helper uses: node:valset:
	// (nodeModulePrefix is declared in halt.gno; valsetSubmodule in valset.gno.)
	prefix := "node:" + mustGnoConst(t, data, "valsetSubmodule") + ":"

	cases := []struct {
		gnoName string
		goPath  string // expected fully-qualified path
	}{
		{"valsetDirtyKey", valsetDirtyPath},
		{"valsetProposedKey", valsetProposedPath},
		{"valsetCurrentKey", valsetCurrentPath},
	}
	for _, tc := range cases {
		got := prefix + mustGnoConst(t, data, tc.gnoName)
		if got != tc.goPath {
			t.Errorf("%s: gno-built path = %q, Go path = %q (drift)", tc.gnoName, got, tc.goPath)
		}
	}
}

func mustGnoConst(t *testing.T, src []byte, name string) string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `\s*=\s*"([^"]+)"`)
	m := re.FindSubmatch(src)
	if len(m) < 2 {
		t.Fatalf("const %q not found in %s", name, strings.TrimSpace(string(src[:80])))
	}
	return string(m[1])
}
