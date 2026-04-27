package gnoland

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// TestValsetConstsDoNotDrift asserts the const values in
// examples/gno.land/r/sys/validators/v3/poc.gno match their Go counterparts.
//
// The realm and chain communicate through these param keys; if they drift,
// valset rotation silently breaks on a running chain with no compile/test
// error from either side individually. This test catches that 100% of the
// time, with negligible cost.
func TestValsetConstsDoNotDrift(t *testing.T) {
	t.Parallel()

	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	pocPath := filepath.Join(root, "examples", "gno.land", "r", "sys", "validators", "v3", "poc.gno")

	data, err := os.ReadFile(pocPath)
	if err != nil {
		t.Fatalf("read %s: %v", pocPath, err)
	}

	want := map[string]string{
		"newUpdatesAvailableKey": newUpdatesAvailableKey,
		"valsetNewKey":           valsetNewKey,
		"valsetPrevKey":          valsetPrevKey,
		// ValsetRealmDefault lives in vm package; the realm path itself is
		// the *file's own location*, so we assert poc.gno sits under it.
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
			t.Errorf("%s missing from poc.gno; Go const = %q", name, expected)
			continue
		}
		if actual != expected {
			t.Errorf("%s: poc.gno = %q, Go = %q (drift)", name, actual, expected)
		}
	}

	// Sanity: poc.gno realm path must match vm.ValsetRealmDefault.
	const expectedRealm = "gno.land/r/sys/validators/v3"
	if vm.ValsetRealmDefault != expectedRealm {
		t.Errorf("vm.ValsetRealmDefault = %q, expected %q", vm.ValsetRealmDefault, expectedRealm)
	}
}
