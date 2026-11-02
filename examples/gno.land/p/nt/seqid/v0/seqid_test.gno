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

func TestID_Binary(cur realm, t *testing.T) {
	var i ID
	prev := i.Binary()

	for j := 0; j < 1000; j++ {
		id := i.Next().Binary()
		if id <= prev {
			t.Fatalf("id %x > prev %x", id, prev)
		}
		prev = id
	}
}

func TestID_String(cur realm, t *testing.T) {
	var i ID
	prev := i.String()

	for j := 0; j < 1000; j++ {
		id := i.Next().String()
		if id <= prev {
			t.Fatalf("id %s > prev %s", id, prev)
		}
		prev = id
	}

	// Test for when cford32 switches over to the long encoding.
	i = 1<<34 - 512
	for j := 0; j < 1024; j++ {
		id := i.Next().String()
		// println(id)
		if id <= prev {
			t.Fatalf("id %s > prev %s", id, prev)
		}
		prev = id
	}
}
