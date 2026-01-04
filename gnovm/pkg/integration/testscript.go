package integration

import (
	"os"
	"strconv"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// NewTestingParams setup and initialize base params for testing.
func NewTestingParams(t *testing.T, testdir string) testscript.Params {
	t.Helper()

	var params testscript.Params
	params.Dir = testdir

	params.UpdateScripts, _ = strconv.ParseBool(os.Getenv("UPDATE_SCRIPTS"))
	params.TestWork, _ = strconv.ParseBool(os.Getenv("TESTWORK"))
	if deadline, ok := t.Deadline(); ok && params.Deadline.IsZero() {
		params.Deadline = deadline
	}

	// Store the original setup scripts for potential wrapping
	params.Setup = func(env *testscript.Env) error {
		// Set the UPDATE_SCRIPTS environment variable
		env.Setenv("UPDATE_SCRIPTS", strconv.FormatBool(params.UpdateScripts))

		// Set the  environment variable
		env.Setenv("TESTWORK", strconv.FormatBool(params.TestWork))

		return nil
	}

	return params
}
