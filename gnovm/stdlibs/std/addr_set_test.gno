package std

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAddrSet(t *testing.T) {
	addresslist := NewAddressList()
	if addresslist.Size() != 0 {
		t.Errorf("Expected size 0, got %d", addresslist.Size())
	}
	test1 := testutils.TestAddress("test1")
	if err := addresslist.AddAddress(test1); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !addresslist.HasAddress(test1) {
		t.Errorf("Address doesnt exist: %v", addresslist.HasAddress(test1))
	}
	if addresslist.Size() != 1 {
		t.Errorf("Expected size 1, got %d", addresslist.Size())
	}
	if err := addresslist.AddAddress(test1); err.Error() != "address already exists" {
		t.Errorf("Expected \"address already exists\", got %v", err)
	}
}
