package integration

import (
	"testing"
)

func TestTestdata(t *testing.T) {
	t.Parallel()

	RunGnolandTestscripts(t, "testdata")
}
