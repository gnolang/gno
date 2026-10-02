package integration

import (
	"fmt"
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

// RegisterExecCommand exposes a binary as a testscript command.
func RegisterExecCommand(p *testscript.Params, name, bin string) {
	if p.Cmds == nil {
		p.Cmds = make(map[string]func(ts *testscript.TestScript, neg bool, args []string))
	}

	if _, exists := p.Cmds[name]; exists {
		panic(fmt.Errorf("unable register %q: command already exist", name))
	}

	p.Cmds[name] = func(ts *testscript.TestScript, neg bool, args []string) {
		err := ts.Exec(bin, args...)
		if err != nil {
			ts.Logf("%s command error: %+v", name, err)
		}

		commandSucceeded := err == nil
		successExpected := !neg
		if commandSucceeded != successExpected {
			ts.Fatalf("unexpected %s command outcome (err=%t expected=%t)", name, commandSucceeded, successExpected)
		}
	}
}

// RunTestscript runs a testscript suite using the shared integration dependency.
func RunTestscript(t *testing.T, p testscript.Params) {
	t.Helper()
	testscript.Run(t, p)
}
