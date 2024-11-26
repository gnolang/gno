package users

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

// todo Test Resolve functions

func TestUpdateBio(t *testing.T) {
	t.Run("caller not registered", func(t *testing.T) {
		t.Parallel()

		std.TestSetOrigCaller(testutils.TestAddress("bob"))

		uassert.Error(t, UpdateBio("newbio"), ErrNotRegistered.Error())
	})
}
