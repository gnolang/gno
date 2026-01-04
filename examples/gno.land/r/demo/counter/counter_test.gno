package counter

import "testing"

func TestIncrement(t *testing.T) {
	counter = 0
	val := Increment(cross)
	if val != 1 {
		t.Fatalf("result from Increment(): %d != 1", val)
	}
	if counter != val {
		t.Fatalf("counter (%d) != val (%d)", counter, val)
	}
}

func TestRender(t *testing.T) {
	counter = 1337
	res := Render("")
	if res != "1337" {
		t.Fatalf("render result %q != %q", res, "1337")
	}
}
