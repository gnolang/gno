package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
)

const httpGetTimeout = 10 * time.Second

// HTTPGetCmd returns a testscript command that fetches a URL and writes the
// response body where the stdout matcher can see it.
//
// This is how a scenario reads the oracle's own verdict. The status board is
// the only machine-readable place the oracle distinguishes rejected, pending,
// gave_up and blocked; everywhere else those live in operator output.
//
// An optional second argument gates on the body: the status board answers
// HTTP 200 for every verdict it knows, including "unknown", so without a
// pattern "eventually ... http_get" would return on the first response no matter
// what it says. The gate has to live in here rather than as a following
// "stdout <pattern>": that check runs only once eventually has already returned, so
// it would report whatever verdict the poll stopped on instead of waiting for
// the one the scenario is after.
func HTTPGetCmd() func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if len(args) != 1 && len(args) != 2 {
			ts.Fatalf("usage: http_get <url> [regex]")
		}

		var pattern *regexp.Regexp
		if len(args) == 2 {
			// Compiled up front and treated like the arg-count check above:
			// a malformed pattern is a usage error, never a match failure.
			// Otherwise "! http_get $URL '['" would pass silently, for
			// entirely the wrong reason.
			p, err := regexp.Compile(args[1])
			if err != nil {
				ts.Fatalf("invalid http_get pattern %q: %v", args[1], err)
			}
			pattern = p
		}

		ctx, cancel := context.WithTimeout(context.Background(), httpGetTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, args[0], nil)
		if err != nil {
			TSValidateError(ts, "http_get", neg, err)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			TSValidateError(ts, "http_get", neg, err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			TSValidateError(ts, "http_get", neg, err)
			return
		}
		fmt.Fprint(ts.Stdout(), string(body))

		if resp.StatusCode != http.StatusOK {
			TSValidateError(ts, "http_get", neg,
				fmt.Errorf("GET %s: %s", args[0], resp.Status))
			return
		}

		if pattern != nil && !pattern.Match(body) {
			TSValidateError(ts, "http_get", neg,
				fmt.Errorf("GET %s: body did not match %q", args[0], pattern))
			return
		}
		TSValidateError(ts, "http_get", neg, nil)
	}
}
