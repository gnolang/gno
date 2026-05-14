package daokit

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
	carol = testutils.TestAddress("carol")
	dave  = testutils.TestAddress("dave")
)

func TestNewDAO(t *testing.T) {
	testing.SetOriginCaller(alice)
	dao := NewCore()

	resourcesLen := dao.Resources.Tree.Size()
	if resourcesLen != 0 { // There is 0 default resources
		t.Errorf("Expected 0 resources, got %d", resourcesLen)
	}
}
