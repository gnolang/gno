package users

import (
	"github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"testing"
)

var (
	alice
)

func TestRegister(t *testing.T) {
	t.Run("paused", func(t *testing.T) {
		std.TestSetOrigCaller(admin)
		Pause()


		std.TestSetOrigCaller(alice)



	})


}
