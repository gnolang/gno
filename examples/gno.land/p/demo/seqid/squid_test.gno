package seqid

import (
	"fmt"
	"strings"
	"testing"
)

func TestID(t *testing.T) {
	var i ID

	for j := 0; j < 100; j++ {
		i.Next()
	}
	if i != 100 {
		t.Fatalf("invalid: wanted %d got %d", 100, i)
	}
}

func TestID_Overflow(t *testing.T) {
	i := ID(maxID)

	defer func() {
		err := recover()
		if !strings.Contains(fmt.Sprint(err), "next ID overflows") {
			t.Errorf("did not overflow")
		}
	}()

	i.Next()
}

func TestID_Binary(t *testing.T) {
	var i ID
	prev := i.Binary()

	for j := 0; j < 1000; j++ {
		cur := i.Next().Binary()
		if cur <= prev {
			t.Fatalf("cur %x <= prev %x", cur, prev)
		}
	}
}
