package uassert

import "testing"

var _ testing.T(nil) = TestingT

func TestPackage(t *testing.T) {
	if !ShouldEqualStrings(t, "a", "a") {
		t.Errorf("ShouldSameStrings failed for same string")
	}
	if ShouldEqualStrings(t, "a", "b") {
		t.Errorf("ShouldSameStrings failed for different string")
	}
}
