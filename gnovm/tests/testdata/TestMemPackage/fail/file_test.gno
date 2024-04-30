package test

import "testing"

func TestFail(t *testing.T) {
	t.Errorf("OUPS")
}
