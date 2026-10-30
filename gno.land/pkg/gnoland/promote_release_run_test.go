package gnoland

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The gh stub answers only the calls promote-release.sh makes, word for word,
// from files in $GH_STUB, and logs each call; anything else fails.
const ghStub = `#!/usr/bin/env bash
set -u
s="${GH_STUB}"
printf 'gh %s\n' "$*" >>"${STUB_LOG}"
case "$*" in
"api user --jq .login") echo tester ;;
"auth token") echo garbage-token ;;
"api --paginate repos/gnolang/gno/releases --jq .[] | select(.draft == false and .prerelease == false) | .tag_name") cat "${s}/releases" ;;
"api repos/gnolang/gno/releases/latest --jq .tag_name") cat "${s}/latest" ;;
"release view "*" --repo gnolang/gno --json isDraft --jq .isDraft")
	[[ $# -eq 9 ]] || exit 64
	if grep -qx -- "$3" "${s}/published"; then
		echo false
	elif grep -qx -- "$3" "${s}/drafts"; then
		echo true
	else
		echo "release not found" >&2
		exit 1
	fi
	;;
"release edit "*" --repo gnolang/gno --prerelease=false --latest")
	[[ $# -eq 7 ]] || exit 64
	printf '%s\n' "$3" >"${s}/latest"
	;;
*)
	echo "gh stub: unmodelled call: $*" >&2
	exit 64
	;;
esac
`

// The curl shim reaches the test's server and nothing else: https://ghcr.io/ is
// rewritten to it, its own URLs pass, and any other URL is refused.
const curlShim = `#!/usr/bin/env bash
set -u
args=()
for a in "$@"; do
	case "${a}" in
	https://ghcr.io/*) a="${STUB_SERVER}/${a#https://ghcr.io/}" ;;
	"${STUB_SERVER}" | "${STUB_SERVER}"/*) ;;
	http://* | https://*)
		echo "curl shim: refused ${a}" >&2
		exit 97
		;;
	esac
	args+=("${a}")
done
exec "${REAL_CURL}" "${args[@]}"
`

const indexType = "application/vnd.oci.image.index.v1+json"

// promoteFixture runs promote-release.sh whole: a fake ghcr.io and mainnet RPC
// behind an httptest server, the gh stub and the curl shim first on PATH, a
// scratch HOME, a garbage GH_TOKEN and proxies on a dead port. Every request
// the server sees and every gh call land, in order, in one event log.
type promoteFixture struct {
	t      *testing.T
	dir    string
	curl   string
	server *httptest.Server

	logMu sync.Mutex
	log   string

	mu        sync.Mutex
	manifests map[string][]byte // by "<image>:<ref>"
	types     map[string]string // by "<image>:<ref>"
	digests   map[string]string // Docker-Content-Digest overrides, by "<image>:<ref>"
	halt      string            // mainnet's halt_min_version
	rpcAnswer string            // a raw answer to the RPC query, replacing the one built from halt
}

// promoteState is everything a case sets before the run. The fixture takes it
// in before its server starts: the server's handlers run on their own
// goroutines, and the requests reach them from a bash process the race
// detector cannot see, so nothing may be set once it serves.
type promoteState struct {
	releases  []string          // the full releases, as GitHub lists them
	published []string          // every published release, pre-releases included
	drafts    []string          // releases still drafts
	latest    string            // GitHub's Latest
	images    string            // the version every image's latest points at
	halt      string            // mainnet's halt_min_version
	rpcAnswer string            // a raw answer to the RPC query, replacing the one built from halt
	digests   map[string]string // Docker-Content-Digest overrides, by "<image>:<ref>"
	types     map[string]string // Content-Type overrides, by "<image>:<ref>"
}

func newPromoteFixture(t *testing.T, st promoteState) *promoteFixture {
	t.Helper()

	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash is not installed")
	}
	curl, err := exec.LookPath("curl")
	if err != nil {
		t.Skip("curl is not installed")
	}

	f := &promoteFixture{
		t:         t,
		dir:       t.TempDir(),
		curl:      curl,
		manifests: map[string][]byte{},
		types:     map[string]string{},
		digests:   map[string]string{},
		halt:      st.halt,
		rpcAnswer: st.rpcAnswer,
	}
	f.log = filepath.Join(f.dir, "events")
	for _, sub := range []string{"bin", "home", "gh"} {
		require.NoError(t, os.Mkdir(filepath.Join(f.dir, sub), 0o700))
	}
	for rel, content := range map[string]string{
		"bin/gh":       ghStub,
		"bin/curl":     curlShim,
		"gh/releases":  lines(st.releases),
		"gh/published": lines(st.published),
		"gh/drafts":    lines(st.drafts),
		"gh/latest":    st.latest + "\n",
		"events":       "",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(f.dir, rel), []byte(content), 0o700))
	}

	for _, img := range promoteLatestImages(t) {
		for _, ref := range []string{"v1.5.1", "v1.7.0", "v1.7.1", "v2.0.0", "v1.7.0-rc.1"} {
			key := img + ":" + ref
			f.manifests[key] = []byte(`{"schemaVersion":2,"mediaType":"` + indexType + `","manifests":[],"annotations":{"ref":"` + key + `"}}`)
			f.types[key] = indexType
		}
		f.manifests[img+":latest"] = f.manifests[img+":"+st.images]
		f.types[img+":latest"] = indexType
	}
	maps.Copy(f.digests, st.digests)
	maps.Copy(f.types, st.types)

	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.server.Close)
	return f
}

func lines(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return strings.Join(ss, "\n") + "\n"
}

// promoteLatestImages reads the images the script moves latest on.
func promoteLatestImages(t *testing.T) []string {
	t.Helper()

	return strings.Fields(grepOne(t, promoteReleaseScript, `(?m)^readonly LATEST_IMAGES=\((.+)\)$`))
}

func (f *promoteFixture) event(format string, args ...any) {
	f.logMu.Lock()
	defer f.logMu.Unlock()
	fh, err := os.OpenFile(f.log, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer fh.Close()
	_, _ = fmt.Fprintf(fh, format+"\n", args...)
}

var manifestPath = regexp.MustCompile(`^/v2/gnolang/gno/([a-z]+)/manifests/([^/]+)$`)

func (f *promoteFixture) serve(w http.ResponseWriter, r *http.Request) {
	m := manifestPath.FindStringSubmatch(r.URL.Path)
	switch {
	case r.URL.Path == "/token":
		f.event("token %s", r.URL.Query().Get("scope"))
		_, _ = io.WriteString(w, `{"token":"stub-token"}`)
	case m != nil && r.Header.Get("Authorization") != "Bearer stub-token":
		f.event("unauthorized %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusUnauthorized)
	case m != nil:
		f.manifest(w, r, m[1]+":"+m[2])
	case r.URL.Path == "/" && r.Method == http.MethodPost:
		f.event("rpc")
		f.mu.Lock()
		answer, halt := f.rpcAnswer, f.halt
		f.mu.Unlock()
		if answer == "" {
			data, _ := json.Marshal(halt)
			answer = `{"jsonrpc":"2.0","id":1,"result":{"response":{"ResponseBase":{"Error":null,"Data":"` + base64.StdEncoding.EncodeToString(data) + `"}}}}`
		}
		_, _ = io.WriteString(w, answer)
	default:
		f.event("unexpected %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (f *promoteFixture) manifest(w http.ResponseWriter, r *http.Request, key string) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		f.event("%s %s", r.Method, key)
		f.mu.Lock()
		body, ok := f.manifests[key]
		digest, ctype := f.digests[key], f.types[key]
		f.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if digest == "" {
			digest = sha256Digest(body)
		}
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Docker-Content-Digest", digest)
		if r.Method == http.MethodGet {
			_, _ = w.Write(body)
		}
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		f.event("PUT %s %s %s", key, r.Header.Get("Content-Type"), sha256Digest(body))
		if !strings.HasSuffix(key, ":latest") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		f.manifests[key], f.types[key] = body, r.Header.Get("Content-Type")
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// stored returns a manifest's bytes and media type, as the server holds them
// after a run.
func (f *promoteFixture) stored(key string) (string, string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return string(f.manifests[key]), f.types[key]
}

func sha256Digest(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// run runs a copy of promote-release.sh with args, and returns its output and
// exit status.
func (f *promoteFixture) run(script string, args ...string) (string, int) {
	f.t.Helper()

	cmd := exec.Command("bash", append([]string{script}, args...)...)
	cmd.Env = []string{
		"PATH=" + filepath.Join(f.dir, "bin") + string(os.PathListSeparator) + os.Getenv("PATH"),
		"HOME=" + filepath.Join(f.dir, "home"),
		"GH_TOKEN=garbage",
		"HTTPS_PROXY=http://127.0.0.1:9", "HTTP_PROXY=http://127.0.0.1:9", "ALL_PROXY=http://127.0.0.1:9",
		"NO_PROXY=127.0.0.1",
		"GH_STUB=" + filepath.Join(f.dir, "gh"),
		"STUB_LOG=" + f.log,
		"STUB_SERVER=" + f.server.URL,
		"REAL_CURL=" + f.curl,
		"MAINNET_RPC=" + f.server.URL,
	}
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return string(out), 0
	case errors.As(err, &exitErr):
		return string(out), exitErr.ExitCode()
	default:
		f.t.Fatalf("could not run bash: %v", err)
		return "", -1
	}
}

// events returns the event log: gh calls and server requests, in order.
func (f *promoteFixture) events() []string {
	f.t.Helper()

	b, err := os.ReadFile(f.log)
	require.NoError(f.t, err)
	if s := strings.TrimSpace(string(b)); s != "" {
		return strings.Split(s, "\n")
	}
	return nil
}

// matching returns the events starting with prefix.
func matching(events []string, prefix string) []string {
	var out []string
	for _, e := range events {
		if strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return out
}

// promote-release.sh run whole, against stubs that cannot reach the network:
// what it refuses, what it asks, and what it writes, in which order.
func TestPromoteReleaseRuns(t *testing.T) {
	t.Parallel()

	script := repoPath(t, promoteReleaseScript)
	images := promoteLatestImages(t)
	held := promoteState{
		releases:  []string{"chain/onyx", "v1.5.1", "v1.4.0"},
		published: []string{"v1.5.1", "v1.4.0", "v1.7.0", "v1.7.0-rc.1"},
		latest:    "v1.5.1",
		images:    "v1.5.1",
		halt:      "v1.7.0",
	}

	t.Run("a dry run reads and writes nothing", func(t *testing.T) {
		t.Parallel()

		f := newPromoteFixture(t, held)
		out, code := f.run(script, "v1.7.0")
		require.Equal(t, 0, code, out)
		assert.Contains(t, out, "dry run")
		ev := f.events()
		assert.Empty(t, matching(ev, "PUT"), "a dry run writes no manifest")
		assert.Empty(t, matching(ev, "gh release edit"), "a dry run edits no release")
		assert.Len(t, matching(ev, "GET"), len(images), "every image is read")
		for _, e := range matching(ev, "token") {
			assert.Regexp(t, `:pull$`, e, "a dry run asks no write token")
		}
		assert.Len(t, matching(ev, "rpc"), 1, "a MINOR asks mainnet once whether it is scheduled")
	})

	// A MINOR is a coordinated upgrade: it becomes the release operators land
	// on only once mainnet's halt proposal names it.
	for _, c := range []struct{ name, halt, answer, says string }{
		{"no halt scheduled", "", "", "mainnet has not scheduled v1.7.0"},
		{"a halt for another release", "v1.6.0", "", "mainnet has not scheduled v1.7.0"},
		{"an answer with no data", "", `{"jsonrpc":"2.0","id":1,"result":{}}`, "gave no halt_min_version"},
		{"an answer with an error", "", `{"jsonrpc":"2.0","id":1,"result":{"response":{"ResponseBase":{"Error":{"msg":"boom"},"Data":null}}}}`, "gave no halt_min_version"},
		{"an answer that is not JSON", "", `<html>bad gateway</html>`, "gave no halt_min_version"},
	} {
		t.Run("a MINOR is refused on "+c.name, func(t *testing.T) {
			t.Parallel()

			st := held
			st.halt = c.halt
			st.rpcAnswer = c.answer
			f := newPromoteFixture(t, st)
			out, code := f.run(script, "v1.7.0", "--push")
			assert.NotEqual(t, 0, code, out)
			assert.Contains(t, out, c.says)
			ev := f.events()
			assert.Empty(t, matching(ev, "token"), "no image is read")
			assert.Empty(t, matching(ev, "PUT"))
			assert.Empty(t, matching(ev, "gh release edit"))
		})
	}

	t.Run("a PATCH asks mainnet nothing", func(t *testing.T) {
		t.Parallel()

		st := held
		st.releases = []string{"v1.7.0", "v1.5.1"}
		st.published = []string{"v1.7.0", "v1.7.1", "v1.5.1"}
		st.halt = ""
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v1.7.1")
		require.Equal(t, 0, code, out)
		assert.Empty(t, matching(f.events(), "rpc"))
	})

	t.Run("a MAJOR is refused before mainnet is asked", func(t *testing.T) {
		t.Parallel()

		st := held
		st.releases = []string{"v1.7.0", "v1.5.1"}
		st.published = []string{"v1.7.0", "v2.0.0", "v1.5.1"}
		st.halt = "v2.0.0"
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v2.0.0")
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "a MAJOR is a new network")
		ev := f.events()
		assert.Empty(t, matching(ev, "rpc"))
		assert.Empty(t, matching(ev, "token"))
	})

	t.Run("a release candidate is refused before anything is asked", func(t *testing.T) {
		t.Parallel()

		f := newPromoteFixture(t, held)
		out, code := f.run(script, "v1.7.0-rc.1", "--push")
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "only a final vMAJOR.MINOR.PATCH release is promoted")
		assert.Empty(t, f.events(), "nothing is asked of GitHub or the registry")
	})

	// release / chain-tag publishes the release once every binary is
	// attached: a draft, or no release at all, means its run has not ended.
	for _, c := range []struct {
		name   string
		drafts []string
		says   string
	}{
		{"no release at all", nil, "no release v1.7.0 on gnolang/gno"},
		{"a draft", []string{"v1.7.0"}, "v1.7.0 is still a draft"},
	} {
		t.Run("a final is refused on "+c.name, func(t *testing.T) {
			t.Parallel()

			st := held
			st.published = []string{"v1.5.1", "v1.4.0"}
			st.drafts = c.drafts
			f := newPromoteFixture(t, st)
			out, code := f.run(script, "v1.7.0", "--push")
			assert.NotEqual(t, 0, code, out)
			assert.Contains(t, out, c.says)
			ev := f.events()
			assert.Empty(t, matching(ev, "token"), "no image is read")
			assert.Empty(t, matching(ev, "PUT"))
			assert.Empty(t, matching(ev, "gh release edit"))
		})
	}

	t.Run("an older final is refused before any image is read", func(t *testing.T) {
		t.Parallel()

		st := held
		st.releases = []string{"v1.7.0", "v1.5.1"}
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v1.5.1", "--push")
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "v1.7.0 is the latest release, and v1.5.1 is older")
		ev := f.events()
		assert.Empty(t, matching(ev, "token"))
		assert.Empty(t, matching(ev, "PUT"))
		assert.Empty(t, matching(ev, "gh release edit"))
	})

	t.Run("bytes that do not match their digest stop the run before any write", func(t *testing.T) {
		t.Parallel()

		st := held
		st.digests = map[string]string{images[len(images)-1] + ":v1.7.0": "sha256:" + strings.Repeat("0", 64)}
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v1.7.0", "--push")
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "bytes hash to")
		ev := f.events()
		assert.Empty(t, matching(ev, "PUT"), "no image moves, not even those read before")
		assert.Empty(t, matching(ev, "gh release edit"))
	})

	t.Run("a media type release / docker does not push stops the run before any write", func(t *testing.T) {
		t.Parallel()

		st := held
		st.types = map[string]string{images[len(images)-1] + ":v1.7.0": "application/json"}
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v1.7.0", "--push")
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "not a manifest type release / docker pushes")
		ev := f.events()
		assert.Empty(t, matching(ev, "PUT"))
		assert.Empty(t, matching(ev, "gh release edit"))
	})

	t.Run("a push moves every latest byte for byte, then the release", func(t *testing.T) {
		t.Parallel()

		f := newPromoteFixture(t, held)
		out, code := f.run(script, "v1.7.0", "--push")
		require.Equal(t, 0, code, out)
		for _, img := range images {
			want, _ := f.stored(img + ":v1.7.0")
			got, ctype := f.stored(img + ":latest")
			assert.Equal(t, want, got, "%s:latest is %s:v1.7.0's bytes", img, img)
			assert.Equal(t, indexType, ctype)
		}
		ev := f.events()
		for _, e := range matching(ev, "token") {
			assert.Regexp(t, `:pull,push$`, e, "a push asks a write token while reading")
		}
		puts, edits := matching(ev, "PUT"), matching(ev, "gh release edit")
		require.Len(t, puts, len(images))
		require.Equal(t, []string{"gh release edit v1.7.0 --repo gnolang/gno --prerelease=false --latest"}, edits)
		assert.Less(t, slices.Index(ev, puts[len(puts)-1]), slices.Index(ev, edits[0]), "the release is edited after every image")
		assert.Contains(t, out, "v1.7.0 is a full release, and Latest")
	})

	// bash 3.2 ends a script aborted by set -u with the EXIT trap's status, 0
	// unless the trap says otherwise; the caller must still see a failure.
	t.Run("an abort exits non-zero, and --help zero", func(t *testing.T) {
		t.Parallel()

		f := newPromoteFixture(t, held)
		src := repoFile(t, promoteReleaseScript)
		broken := strings.Replace(src, "  check_login\n", "  check_login\n  : \"${UNSET_BY_THE_TEST}\"\n", 1)
		require.NotEqual(t, src, broken, "could not inject the abort")
		copied := filepath.Join(t.TempDir(), "promote-release.sh")
		require.NoError(t, os.WriteFile(copied, []byte(broken), 0o700))

		out, code := f.run(copied, "v1.7.0")
		assert.NotEqual(t, 0, code, "an unset variable aborts the run:\n%s", out)
		assert.Contains(t, out, "UNSET_BY_THE_TEST")

		out, code = f.run(script, "--help")
		assert.Equal(t, 0, code, out)
		assert.Contains(t, out, "Usage:")
	})

	// The recovery from a release skipped after its promotion: mark it a
	// pre-release again, then promote the release the chain runs.
	t.Run("promoting the running release puts latest back", func(t *testing.T) {
		t.Parallel()

		st := held
		st.releases = []string{"v1.5.1"}
		st.latest = "v1.7.0"
		st.images = "v1.7.0"
		f := newPromoteFixture(t, st)
		out, code := f.run(script, "v1.5.1", "--push")
		require.Equal(t, 0, code, out)
		assert.Contains(t, out, "v1.5.1 is the latest release already")
		for _, img := range images {
			want, _ := f.stored(img + ":v1.5.1")
			got, _ := f.stored(img + ":latest")
			assert.Equal(t, want, got, "%s:latest is back on v1.5.1", img)
		}
		ev := f.events()
		assert.Equal(t, []string{"gh release edit v1.5.1 --repo gnolang/gno --prerelease=false --latest"}, matching(ev, "gh release edit"))
		assert.Empty(t, matching(ev, "rpc"), "the release the chain runs needs no scheduled halt")
	})
}
