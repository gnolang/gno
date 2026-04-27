package gnoland

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// TestValsetConstsDoNotDrift asserts the param-key string values in the
// gno-side helper (examples/gno.land/r/sys/params/valset.gno) match the
// Go-side constants in this package. The realm and chain communicate
// through these keys; if they drift, valset rotation silently breaks at
// runtime with no compile/test error from either side individually.
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

	want := map[string]string{
		"valsetDirtyKey":       valsetDirtyKey,
		"valsetNewPubKeysKey":  valsetNewPubKeysKey,
		"valsetNewPowersKey":   valsetNewPowersKey,
		"valsetPrevPubKeysKey": valsetPrevPubKeysKey,
		"valsetPrevPowersKey":  valsetPrevPowersKey,
	}

	// Match: <name> = "<value>"
	re := regexp.MustCompile(`(?m)^\s*([a-zA-Z][a-zA-Z0-9]*)\s*=\s*"([^"]+)"`)
	got := map[string]string{}
	for _, m := range re.FindAllStringSubmatch(string(data), -1) {
		got[m[1]] = m[2]
	}

	for name, expected := range want {
		actual, ok := got[name]
		if !ok {
			t.Errorf("%s missing from %s; Go const = %q", name, gnoPath, expected)
			continue
		}
		if actual != expected {
			t.Errorf("%s: gno = %q, Go = %q (drift)", name, actual, expected)
		}
	}

	// Sanity: realm path constant matches v3 realm location.
	const expectedRealm = "gno.land/r/sys/validators/v3"
	if vm.ValsetRealmDefault != expectedRealm {
		t.Errorf("vm.ValsetRealmDefault = %q, expected %q", vm.ValsetRealmDefault, expectedRealm)
	}
}
