package watchdog

import (
	"testing"
	"time"

	"gno.land/p/nt/uassert/v0"
)

func TestPackage(t *testing.T) {
	w := Watchdog{Duration: 5 * time.Minute}
	uassert.False(t, w.IsAlive())
	w.Alive()
	uassert.True(t, w.IsAlive())
	// XXX: add more tests when we'll be able to "skip time".
}
