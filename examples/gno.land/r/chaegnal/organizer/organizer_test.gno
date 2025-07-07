package organizer

import (
	"testing"

	"gno.land/p/demo/uassert"
)

// Helpers

// Test getGroupFromId
func Test_getGroupFromId_invalidID(t *testing.T) {
	uassert.PanicsWithMessage(t, "invalid id", func() { GetGroupFromId("abc") })
}

func Test_getGroupFromId_notFound(t *testing.T) {
	uassert.PanicsWithMessage(t, "could not find group", func() { GetGroupFromId("100") })
}
