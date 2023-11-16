package main

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestPrecompile(t *testing.T) {
	testscript.Run(t, setupTestScript(t, "testdata/gno_precompile"))
}
