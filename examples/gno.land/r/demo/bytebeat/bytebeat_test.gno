package bytebeat

import (
	"testing"
)

func TestByteBeat(t *testing.T) {
	s := Render("")
	if len(s) != 128194 {
		t.Errorf("correct len: %d", len(s))
	}
}
