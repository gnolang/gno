package rand

import (
	"fmt"
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/rand"
)

func TestInstance(t *testing.T) {
	instance := rand.New()
	if instance == nil {
		t.Errorf("instance should not be nil")
	}
}

func TestIntn(t *testing.T) {
	baseRand := rand.New()
	baseResult := computeIntn(t, baseRand)

	sameHeightRand := rand.New()
	sameHeightResult := computeIntn(t, sameHeightRand)

	if baseResult != sameHeightResult {
		t.Errorf("should have the same result: new=%s, base=%s", sameHeightResult, baseResult)
	}

	std.TestSkipHeights(1)
	differentHeightRand := rand.New()
	differentHeightResult := computeIntn(t, differentHeightRand)

	if baseResult == differentHeightResult {
		t.Errorf("should have different result: new=%s, base=%s", differentHeightResult, baseResult)
	}
}

func computeIntn(t *testing.T, r *rand.Instance) string {
	t.Helper()

	arr := []string{}
	for i := 0; i < 10; i++ {
		arr = append(arr, fmt.Sprintf("%d", r.Intn(1000)))
	}

	out := strings.Join(arr, ",")
	return out
}
