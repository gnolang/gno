package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func TestHTTPGetBodyReachesStdout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"path":"gno.land/r/x/y","status":"rejected"}`)
	}))
	t.Cleanup(srv.Close)

	adapter := NewTestscriptT(testLogger(t), false)
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	script := fmt.Sprintf("http_get %s/status\nstdout '\"status\":\"rejected\"'\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed)
}

func TestHTTPGetFailsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "boom")
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	// An un-negated call must still fail the script on a non-200.
	adapter := NewTestscriptT(testLogger(t), false)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, fmt.Sprintf("http_get %s/status\n", srv.URL))},
		Cmds:  cmds,
	})
	require.True(t, adapter.Failed)

	// The response body must still reach stdout on a non-200, so a scenario
	// can see what the endpoint actually said rather than just "it failed".
	// The call is negated so the script expects the failure and keeps
	// running to the stdout check instead of aborting on it; testscript
	// captures a negated custom command's stdout the same as any other, so
	// this script is expected to complete without failing.
	negAdapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf("! http_get %s/status\nstdout 'boom'\n", srv.URL)
	testscript.RunT(negAdapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.False(t, negAdapter.Failed)
}

func TestHTTPGetPatternMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"approved"}`)
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	adapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf(`http_get %s/status '"status":"approved"'`+"\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.False(t, adapter.Failed)
}

func TestHTTPGetPatternFailsOnNonMatchingBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"pending"}`)
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	// The response is a plain 200: the gate itself must be what fails the
	// command here, not the status code.
	adapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf(`http_get %s/status '"status":"approved"'`+"\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.True(t, adapter.Failed)
}

func TestHTTPGetPatternNegationFailsWhenBodyMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"approved"}`)
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	// Round 1 found a Critical defect where "! gnokey" could never fail
	// because a recover() swallowed the panic reporting unexpected success.
	// http_get has no such recover, but that is exactly the kind of
	// assumption that hid the last defect, so verify it directly: a negated
	// call must still fail when the body actually matches the pattern.
	adapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf(`! http_get %s/status '"status":"approved"'`+"\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.True(t, adapter.Failed, "! http_get must fail when the body matches the pattern")
}

func TestHTTPGetInvalidPatternIsUsageErrorEvenWhenNegated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "anything")
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	// A malformed regex must be a usage error, not "did not match": otherwise
	// "! http_get $URL '['" would pass silently for entirely the wrong
	// reason. Usage errors are unconditional, so this must fail even negated.
	adapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf(`! http_get %s/status '['`+"\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.True(t, adapter.Failed, "an invalid regex must fail even when http_get is negated")
}

func TestHTTPGetTooManyArgsIsUsageError(t *testing.T) {
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}

	adapter := NewTestscriptT(testLogger(t), false)
	script := "http_get http://example.invalid pattern extra\n"
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})
	require.True(t, adapter.Failed)
}

func TestHTTPGetEventuallyPollsUntilBodyMatches(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			fmt.Fprint(w, `{"status":"pending"}`)
			return
		}
		fmt.Fprint(w, `{"status":"approved"}`)
	}))
	t.Cleanup(srv.Close)

	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"http_get": HTTPGetCmd(),
	}
	cmds["eventually"] = EventuallyCmd(cmds)

	// This is the capability TASK-4 needs: eventually must keep polling past
	// early non-matching bodies and only return once the body matches, not
	// on the first (still-200) response.
	adapter := NewTestscriptT(testLogger(t), false)
	script := fmt.Sprintf(`eventually 2s 10ms http_get %s/status '"status":"approved"'`+"\n", srv.URL)
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, script)},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed)
	require.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(3),
		"eventually must poll past the first non-matching response before succeeding")
}
