package context

import "testing"

func TestContextExample(t *testing.T) {
	type favContextKey string

	k := favContextKey("language")
	ctx := WithValue(Empty(), k, "Gno")

	if v := ctx.Value(k); v != nil {
		if v.(string) != "Gno" {
			t.Errorf("language value should be Gno, but is %s", v)
		}
	} else {
		t.Errorf("language key value was not found")
	}

	if v := ctx.Value(favContextKey("color")); v != nil {
		t.Errorf("color key was found")
	}
}

// otherContext is a Context that's not one of the types defined in context.go.
// This lets us test code paths that differ based on the underlying type of the
// Context.
type otherContext struct {
	Context
}

type (
	key1 int
	key2 int
)

// func (k key2) String() string { return fmt.Sprintf("%[1]T(%[1]d)", k) }

var (
	k1 = key1(1)
	k2 = key2(1) // same int as k1, different type
	k3 = key2(3) // same type as k2, different int
)

func TestValues(t *testing.T) {
	check := func(c Context, nm, v1, v2, v3 string) {
		if v, ok := c.Value(k1).(string); ok == (len(v1) == 0) || v != v1 {
			t.Errorf(`%s.Value(k1).(string) = %q, %t want %q, %t`, nm, v, ok, v1, len(v1) != 0)
		}
		if v, ok := c.Value(k2).(string); ok == (len(v2) == 0) || v != v2 {
			t.Errorf(`%s.Value(k2).(string) = %q, %t want %q, %t`, nm, v, ok, v2, len(v2) != 0)
		}
		if v, ok := c.Value(k3).(string); ok == (len(v3) == 0) || v != v3 {
			t.Errorf(`%s.Value(k3).(string) = %q, %t want %q, %t`, nm, v, ok, v3, len(v3) != 0)
		}
	}

	c0 := Empty()
	check(c0, "c0", "", "", "")

	t.Skip() // XXX: depends on https://github.com/gnolang/gno/issues/2386

	c1 := WithValue(Empty(), k1, "c1k1")
	check(c1, "c1", "c1k1", "", "")

	/*if got, want := c1.String(), `context.Empty.WithValue(context_test.key1, c1k1)`; got != want {
		t.Errorf("c.String() = %q want %q", got, want)
	}*/

	c2 := WithValue(c1, k2, "c2k2")
	check(c2, "c2", "c1k1", "c2k2", "")

	/*if got, want := fmt.Sprint(c2), `context.Empty.WithValue(context_test.key1, c1k1).WithValue(context_test.key2(1), c2k2)`; got != want {
		t.Errorf("c.String() = %q want %q", got, want)
	}*/

	c3 := WithValue(c2, k3, "c3k3")
	check(c3, "c2", "c1k1", "c2k2", "c3k3")

	c4 := WithValue(c3, k1, nil)
	check(c4, "c4", "", "c2k2", "c3k3")

	o0 := otherContext{Empty()}
	check(o0, "o0", "", "", "")

	o1 := otherContext{WithValue(Empty(), k1, "c1k1")}
	check(o1, "o1", "c1k1", "", "")

	o2 := WithValue(o1, k2, "o2k2")
	check(o2, "o2", "c1k1", "o2k2", "")

	o3 := otherContext{c4}
	check(o3, "o3", "", "c2k2", "c3k3")

	o4 := WithValue(o3, k3, nil)
	check(o4, "o4", "", "c2k2", "")
}
