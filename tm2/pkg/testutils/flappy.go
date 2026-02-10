package testutils

import (
	"os"
	"testing"
)

type StabilityMode string

const (
	Stable StabilityMode = "stable"
	Flappy StabilityMode = "flappy"
	Broken StabilityMode = "broken"
)

func FilterStability(t *testing.T, mode StabilityMode) {
	t.Helper()

	filter := os.Getenv("STABILITY_FILTER")
	if filter != string(mode) {
		t.Skipf("skip test with %q stability", mode)
	}
}
