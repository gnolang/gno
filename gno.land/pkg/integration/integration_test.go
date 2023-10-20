package integration

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestTestdata(t *testing.T) {
	testscript.Run(t, SetupGnolandTestScript(t, "testdata"))
}
