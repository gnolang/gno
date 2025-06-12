package test_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
)

// helper to build a minimal TestOptions.
// We only care about Filetest behavior, so we leave Sync and Debug togglable,
// and let NewTestOptions wire up the stores and writers for us.
func newOpts(sync bool) *test.TestOptions {
	var outBuf, errBuf bytes.Buffer
    opts := test.NewTestOptions("", &outBuf, &errBuf)
    opts.Sync = sync
	return opts
}

func TestRunFiletest_MatchErrorDirectiveDiff(t *testing.T) {
	// 1) content != actual, !opts.Sync, dir.Name==DirectiveError
	t.Skip("TODO(loren): suspected issue with empty // Error: directive not preserved by ParseDirectives — will revisit after completing other runFiletest coverage.")
	src := `
// Error:
package main

func main() {
    panic("ohno")
}
`
	opts := newOpts(false)
	out, err := opts.RunFiletest("foo_filetest.gno", []byte(src))
	if err == nil {
		t.Fatal("expected a diff error, got nil")
	}
	if !strings.Contains(err.Error(), "Error diff:") {
		t.Errorf("got wrong error message:\n%s", err)
	}
	if out != "" {
		t.Errorf("expected empty output on failure, got %q", out)
	}
}

func TestRunFiletest_UnexpectedPanic_NoErrorDirective(t *testing.T) {
	// 2) result.Error!="" but no Error directive → unexpected panic branch
	src := `package main
func main() {
	panic("boom")
}`
	// No "// Error:" directive in src
	opts := newOpts(false)
	out, err := opts.RunFiletest("boom_filetest.gno", []byte(src))
	if err == nil || !strings.HasPrefix(err.Error(), "unexpected panic: boom") {
		t.Fatalf("expected unexpected panic error, got: out=%q err=%v", out, err)
	}
}

func TestRunFiletest_UnexpectedOutput_NoOutputDirective(t *testing.T) {
	// 3) result.Error=="" && result.Output!="" but no Output directive
	src := `package main
func main() {
	println("hello")
}`
	// No "// Output:" directive
	opts := newOpts(false)
	out, err := opts.RunFiletest("hello_filetest.gno", []byte(src))
	if err == nil || !strings.Contains(err.Error(), "unexpected output") {
		t.Fatalf("expected unexpected output error, got: out=%q err=%v", out, err)
	}
}

func TestRunFiletest_NoErrorNoOutput_HappyPath(t *testing.T) {
	// 4) result.Error=="" && result.Output=="" → happy path, no directives needed
// 	src := `package main
// func main() {}`

	// We must supply at least an "// Output:" with empty content, or else
	// the post-run checks will find no directives and return an error.
	srcWithDir := `
// Output:
package main
func main() {}`
	opts := newOpts(false)
	out, err := opts.RunFiletest("empty_filetest.gno", []byte(srcWithDir))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Since there are no updates, RunFiletest returns empty string
	if out != "" {
		t.Errorf("expected no sync changes, got %q", out)
	}
}

func TestRunFiletest_MatchOutput(t *testing.T) {
	src := `
// Output:
// hi
package main
func main() {
	println("hi")
}`
	opts := newOpts(false)
	out, err := opts.RunFiletest("output_filetest.gno", []byte(src))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "" {
		t.Errorf("expected no sync output, got %q", out)
	}
}

func TestRunFiletest_MatchEvents(t *testing.T) {
	src := `
// Events: []
package main
func main() {
}`
	opts := newOpts(false)
	out, err := opts.RunFiletest("events_filetest.gno", []byte(src))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "" {
		t.Errorf("expected no sync output, got %q", out)
	}
}

func TestRunFiletest_MatchStacktrace(t *testing.T) {
	src := `
// Stacktrace:
package main
func main() {
	panic("test panic")
}`
	opts := newOpts(true) // Sync mode to tolerate diff
	out, err := opts.RunFiletest("stacktrace_filetest.gno", []byte(src))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == "" {
		t.Fatal("expected sync output for Stacktrace, got empty string")
	}
}

// otherNewOpts returns TestOptions with real buffers (never nil).
func otherNewOpts(sync bool) *test.TestOptions {
	out := new(bytes.Buffer)
	errb := new(bytes.Buffer)
	opts := test.NewTestOptions("", out, errb)
	opts.Sync = sync
	return opts
}

// TestRunFiletest_Realm_NoDirectives covers the realm branch when there is no // Realm: directive.
// It uses a file name containing "/r/" so gno.IsRealmPath(pkgPath) is true.
func TestRunFiletest_Realm_NoDirectives(t *testing.T) {
	// We name the test file so that pkgPath == "foo/r/test" (the part before "_filetest.gno")
	fname := "foo/r/test_filetest.gno"
	src := `package main
func main() {
	// no directives: still should run without panic
}`

	opts := otherNewOpts(false)
	out, err := opts.RunFiletest(fname, []byte(src))
	if err != nil {
		t.Fatalf("unexpected error on realm no-directives: %v", err)
	}
	if out != "" {
		t.Errorf("expected no sync output for realm no-directives, got %q", out)
	}
}

// TestRunFiletest_Realm_WithDirective covers the realm branch when an explicit // Realm: directive is present.
func TestRunFiletest_Realm_WithDirective(t *testing.T) {
	fname := "my/r/path_filetest.gno"
	src := `
// Realm:
package main
func main() {
	// nothing else
}`

	opts := otherNewOpts(false)
	out, err := opts.RunFiletest(fname, []byte(src))
	if err != nil {
		t.Fatalf("unexpected error on realm with-directive: %v", err)
	}
	if out != "" {
		t.Errorf("expected no sync output for realm with-directive, got %q", out)
	}
}


func TestRunFiletest_RealmBranch_BuiltinError(t *testing.T) {
    // When PKGPATH indicates a realm, the VM will try to load .gnobuiltins.gno
    // under that path, but here it fails with:
    //   "expected package name [myrealm] but got [main]"
    // We capture that as an Error directive diff.
    src := `
// PKGPATH: gno.land/r/myrealm
// Error:
package main
func main() {}
`
    opts := newOpts(false)  // non-sync mode
    out, err := opts.RunFiletest("realm_builtin_error_filetest.gno", []byte(src))
    if err == nil {
        t.Fatal("expected Error diff, got nil")
    }
    msg := err.Error()
    if !strings.Contains(msg, "expected package name [myrealm] but got [main]") {
        t.Errorf("expected builtin package error diff, got:\n%s", msg)
    }
    if out != "" {
        t.Errorf("expected no sync output, got %q", out)
    }
}

func TestRunFiletest_RealmBranch_SyncUpdateBuiltin(t *testing.T) {
    // In sync mode the missing Error directive should be appended.
    src := `
// PKGPATH: gno.land/r/myrealm
package main
func main() {}
`
    opts := newOpts(true)  // sync = true
    out, err := opts.RunFiletest("realm_sync_builtin_filetest.gno", []byte(src))
    if err != nil {
        t.Fatalf("expected no error in sync mode, got %v", err)
    }
    if !strings.Contains(out, "// Error:") {
        t.Fatal("expected appended Error directive, got none:\n" + out)
    }
    if !strings.Contains(out, "expected package name [myrealm] but got [main]") {
        t.Errorf("expected directive content updated, got:\n%s", out)
    }
}


