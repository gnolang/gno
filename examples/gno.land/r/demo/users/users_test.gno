package users

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestPreRegisteredTest1(t *testing.T) {
	names := ListUsersByPrefix("test1", 1)
	uassert.Equal(t, len(names), 1)
	uassert.Equal(t, names[0], "test1")
}
