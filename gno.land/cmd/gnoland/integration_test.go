package main

import (
	"testing"

	integration "github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestTestdata(t *testing.T) {
	testscript.Run(t, integration.SetupGnolandTestScript(t, "testdata"))
}
