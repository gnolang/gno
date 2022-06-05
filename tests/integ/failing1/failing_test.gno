package failing

import "testing"

func TestAlwaysFailing(t *testing.T) {
	t.Fatal("failed")
}
