package skiptime

import (
	"std"
	"testing"
	"time"
)

func TestSkipHeights(t *testing.T) {
	oldHeight := std.ChainHeight()
	shouldEQ(t, oldHeight, 123)

	oldNow := time.Now().Unix()
	shouldEQ(t, oldNow, 1234567890)

	// skip 3 blocks == 15 seconds
	testing.SkipHeights(3)

	shouldEQ(t, std.ChainHeight()-oldHeight, 3)
	shouldEQ(t, time.Now().Unix()-oldNow, 15)
}

func shouldEQ(t *testing.T, got, expected int64) {
	if got != expected {
		t.Fatalf("expected %d, got %d.", expected, got)
	}
}
