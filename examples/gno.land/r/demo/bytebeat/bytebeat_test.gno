package bytebeat

import (
	"testing"
)

func TestByteBeat(t *testing.T) {
	s := Render("(t>>6|t|t>>(t>>16))*10+((t>>11)&7)")
	if len(s) != 21571 {
		t.Errorf("correct len: %d", len(s))
	}
}
