package main

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
)

func TestTestdata(t *testing.T) {
	integration.RunGnolandTestscripts(t, "testdata")
}
