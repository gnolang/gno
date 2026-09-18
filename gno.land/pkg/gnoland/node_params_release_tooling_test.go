package gnoland

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoFile reads a file by its path relative to the repository root.
func repoFile(t *testing.T, rel string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(rel)))
	require.NoError(t, err, "%s — has it moved? The release tooling is checked against this package.", rel)
	return string(b)
}

// grepOne returns the single capture of re in the named file.
func grepOne(t *testing.T, rel, re string) string {
	t.Helper()

	m := regexp.MustCompile(re).FindStringSubmatch(repoFile(t, rel))
	require.Len(t, m, 2, "%s: no line matching %s — was it reformatted onto several lines?", rel, re)
	return m[1]
}

// The release scripts under misc/release/ mirror parseReleaseVersion in a shell
// regex, and nothing else in CI reads them: ci-dir-misc.yml's matrix has no
// release row. Left alone, the two drift, and the direction that matters is the
// script being the looser of the pair — it would then accept a tag the node
// cannot parse, and --halt-height would write that tag into a governance
// proposal. A floor the node cannot parse degrades to byte equality and refuses
// the upgraded binary alongside the stale ones, so the chain does not restart.
//
// This test runs in a package the pull-request jobs already build.
func TestReleaseToolingMatchesTheParser(t *testing.T) {
	t.Parallel()

	const (
		cutRelease = "misc/release/cut-release.sh"
		bumpProto  = "misc/release/bump-protocol-version.sh"
		setHalt    = "misc/govdao-scripts/set-halt.sh"
		workflow   = ".github/workflows/release-chain-tag.yml"
	)

	shapeRE := grepOne(t, cutRelease, `(?m)^readonly VERSION_RE='(.+)'$`)

	t.Run("the shape check never accepts more than the node parses", func(t *testing.T) {
		t.Parallel()

		re, err := regexp.Compile(shapeRE)
		require.NoError(t, err, "VERSION_RE in %s does not compile as a Go regexp", cutRelease)

		for _, v := range []string{
			// Accepted by both.
			"v0.0.0", "v1.2.0", "v1.2.3", "v10.20.30", "v1.3.0-rc.1", "v1.3.0-rc.10",
			"v1.0.0-alpha", "v1.0.0-alpha.1", "v1.0.0-0.3.7", "v1.0.0-x-y-z.0",
			"v99999999999999999999.0.0",
			// Rejected by both. Leading zeros and signed components would alias
			// two tag names onto one version; a dangling separator would yield
			// an empty pre-release, the sentinel for "this is the release".
			"v1.02.0", "v1.2.03", "v01.2.3", "v1.-2.0", "v1.+2.0",
			"v1.2.3-", "v1.2.3-01", "v1.2.3-.", "v1.2.3-a..b",
			"v1.2", "v1", "v1.2.3.4", "v1.x.0", "v1..0", "1.2.0", "",
			"develop", "master.3335+bc43a5fb7", "mainnet",
		} {
			_, parses := parseReleaseVersion(v)
			assert.Equal(t, parses, re.MatchString(v),
				"%s and parseReleaseVersion disagree on %q (script accepts: %v, node parses: %v)",
				cutRelease, v, re.MatchString(v), parses)
		}

		// Betanet's shape is understood by the node but is not something the
		// script cuts: it tags the v line only.
		for _, v := range []string{"chain/gnoland1.0", "chain/gnoland1.1", "chain/mainnet"} {
			assert.False(t, re.MatchString(v), "%s should not accept %q", cutRelease, v)
		}

		// Accepted by the node, refused by the script on purpose: build
		// metadata takes no part in ordering, so parseReleaseVersion strips it,
		// while the release tooling should not cut a tag carrying "+". The
		// disagreement is fail-closed. The dangerous direction is the reverse,
		// a script accepting a tag the node cannot parse, and the corpus above
		// is what pins it.
		for _, v := range []string{"v1.2.0+deadbeef", "v1.3.0-rc.1+deadbeef"} {
			_, parses := parseReleaseVersion(v)
			assert.True(t, parses, "parseReleaseVersion should accept %q", v)
			assert.False(t, re.MatchString(v), "%s should refuse %q", cutRelease, v)
		}
	})

	t.Run("every script checks the same shape", func(t *testing.T) {
		t.Parallel()

		// Not a shared constant: they are standalone scripts, invoked from
		// different places — bump-protocol-version.sh runs against trees that
		// predate the others, and set-halt.sh runs from a deployment directory.
		bumpRE := grepOne(t, bumpProto, `(?m)^\t\[\[ \$\{target\} =~ (\^v.+\$) \]\] \\$`)
		assert.Equal(t, shapeRE, bumpRE,
			"%s and %s check different shapes; the protocol version is compared with semver.Compare too", cutRelease, bumpProto)

		// set-halt.sh is the one a human runs to set the floor, so a looser
		// shape there reaches the chain without passing the other two at all.
		haltRE := grepOne(t, setHalt, `(?m)^  VERSION_RE='(.+)'$`)
		assert.Equal(t, shapeRE, haltRE,
			"%s accepts a different shape than %s; it writes halt_min_version directly", setHalt, cutRelease)
	})

	// cut-release.sh's build check earns its place only by building the way the
	// workflow builds. If the workflow's -ldflags symbol is renamed and the
	// script's is not, the script goes on asserting a symbol nothing declares
	// and passes a release whose binaries report "develop".
	t.Run("the build check uses the workflow's version symbol", func(t *testing.T) {
		t.Parallel()

		pkg := grepOne(t, cutRelease, `(?m)^readonly VERSION_PKG="(.+)"$`)
		assert.True(t, strings.Contains(repoFile(t, workflow), "-X "+pkg+"="),
			"%s injects a different symbol than %s's VERSION_PKG (%s)", workflow, cutRelease, pkg)

		// And the symbol has to be the one the node actually reports.
		assert.True(t, strings.HasPrefix(pkg, "github.com/gnolang/gno/tm2/pkg/version."),
			"VERSION_PKG (%s) is not in tm2/pkg/version, which is what meetsMinVersion reads", pkg)
	})
}
