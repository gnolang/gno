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

const promoteReleaseScript = "misc/release/promote-release.sh"

// codeLines returns the lines of a file with comments cut off. A comment
// starts at a "#" that is at the start of the line or after a space, is
// followed by a space or the end of the line, and has an even number of '"'
// and of "'" before it on its line. So a "#" inside a word, or inside a quoted
// string such as "Fixes #6429" or "held # until the halt", is kept. Quotes are
// counted, not parsed, so an escaped quote or an apostrophe inside double
// quotes throws the count off. Before a comment, that keeps the comment as
// code, which fails a check; before a "#" inside a string, as in "a \" # b",
// it cuts the line there and hides what follows from the checks.
func codeLines(t *testing.T, rel string) []string {
	t.Helper()

	comment := regexp.MustCompile(`(?:^|\s)#(?:\s|$)`)
	var lines []string
	for line := range strings.SplitSeq(repoFile(t, rel), "\n") {
		for _, loc := range comment.FindAllStringIndex(line, -1) {
			if before := line[:loc[0]]; strings.Count(before, `"`)%2 == 0 && strings.Count(before, "'")%2 == 0 {
				line = before
				break
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// A final release is published days before the chain it is for switches to
// it, so neither release workflow may make it the release operators land on:
// GitHub's "Latest", misc/install.sh's default and the CLI images' latest stay
// on the release the chain runs until promote-release.sh moves them, once the
// chain's switch is set.
func TestReleasesArePublishedHeld(t *testing.T) {
	t.Parallel()

	t.Run("every release is published as a pre-release, never Latest", func(t *testing.T) {
		t.Parallel()

		// Every word carrying a --prerelease or --latest flag, or the API's
		// make_latest and prerelease fields, in order: the draft is created as
		// a pre-release, whatever the tag, and published as one, without taking
		// Latest. Whole words, so a flag assigned to a variable shows as the
		// assignment it is and fails the comparison.
		flag := regexp.MustCompile(`\S*(?:--(?:prerelease|latest)|make_latest|\bprerelease=)\S*`)
		var flags []string
		for _, line := range codeLines(t, releaseChainTagWorkflow) {
			flags = append(flags, flag.FindAllString(line, -1)...)
		}
		assert.Equal(t, []string{"--prerelease", "--prerelease", "--latest=false"}, flags,
			"%s must create and publish every release as a pre-release, without Latest", releaseChainTagWorkflow)
	})

	t.Run("no image is published as latest", func(t *testing.T) {
		t.Parallel()

		// Wherever a tag could come from (metadata-action's rules, its
		// default, a bake override), the only latest in the workflow is the
		// flavor turning metadata-action's own off.
		word := regexp.MustCompile(`(?:^|[^-\w])latest\b`)
		var latest []string
		for _, line := range codeLines(t, releaseDockerWorkflow) {
			if word.MatchString(line) {
				latest = append(latest, line)
			}
		}
		assert.Equal(t, []string{"            latest=false"}, latest,
			"%s must push no latest tag, and turn off metadata-action's", releaseDockerWorkflow)
	})

	t.Run("promote-release.sh moves latest on images the release builds, never the node's", func(t *testing.T) {
		t.Parallel()

		built := regexp.MustCompile(`(?m)^          - (\S+)$`).FindAllStringSubmatch(
			grepOne(t, releaseDockerWorkflow, `(?ms)^        image:\n((?:          - \S+\n)+)`), -1)
		var images []string
		for _, m := range built {
			images = append(images, m[1])
		}
		assert.Contains(t, images, "gnoland", "%s: could not read the image matrix", releaseDockerWorkflow)

		latest := strings.Fields(grepOne(t, promoteReleaseScript, `(?m)^readonly LATEST_IMAGES=\((.+)\)$`))
		// RELEASING.md, "Container images": exactly the four CLI tools.
		assert.ElementsMatch(t, []string{"gno", "gnokey", "gnodev", "gnoweb"}, latest)
		for _, img := range latest {
			assert.Contains(t, images, img, "%s moves latest on %s, which %s does not build", promoteReleaseScript, img, releaseDockerWorkflow)
		}
		// A node image that moves by itself breaks a coordinated upgrade
		// (RELEASING.md, "Container images").
		assert.NotContains(t, latest, "gnoland", "%s must never move gnoland:latest", promoteReleaseScript)
	})
}

// promote-release.sh makes a final release the one misc/install.sh, GitHub's
// "Latest" and the CLI images' latest point at. What it may promote, and from
// which release, is decided in its functions below; the registry and release
// writes are not run here.
func TestPromoteRelease(t *testing.T) {
	t.Parallel()

	// The functions run without the script's top level, so its shape comes in
	// with the other globals.
	shape := scriptShape(t, promoteReleaseScript)

	t.Run("only a final release is promoted", func(t *testing.T) {
		t.Parallel()

		for _, v := range []string{"v1.7.0", "v1.10.0", "v2.0.0"} {
			t.Run(v, func(t *testing.T) {
				out, returned := runScriptFuncs(t, promoteReleaseScript, map[string]string{"VERSION": v, "VERSION_RE": shape}, "check_final")
				assert.True(t, returned, "check_final refused %q:\n%s", v, out)
			})
		}
		for _, v := range []string{"v1.7.0-rc.1", "v1.7.0+deadbeef", "v1.07.0", "v1.7", "1.7.0", "chain/mainnet", ""} {
			t.Run("refuses "+v, func(t *testing.T) {
				out, returned := runScriptFuncs(t, promoteReleaseScript, map[string]string{"VERSION": v, "VERSION_RE": shape}, "check_final")
				assert.False(t, returned, "check_final accepted %q:\n%s", v, out)
				assert.Contains(t, out, "only a final vMAJOR.MINOR.PATCH release is promoted", "check_final on %q", v)
			})
		}
	})

	// The floor is the highest full final release, the one misc/install.sh
	// installs: not GitHub's "Latest" marker, which a chain/<name> genesis
	// record created by hand can hold, and not the highest tag, since a patch
	// to the running line is promoted while the next MINOR is held.
	t.Run("a promotion only moves forward from the highest full release", func(t *testing.T) {
		t.Parallel()

		for _, c := range []struct {
			releases, version string
			moves             bool
			says              string
		}{
			{"chain/onyx\nv1.5.1\nv1.4.0\nchain/mainnet", "v1.7.0", true, "v1.5.1 → v1.7.0"},
			{"v1.9.0\nv1.2.0", "v1.10.0", true, "v1.9.0 → v1.10.0"},
			{"v1.7.0\nv1.5.1", "v1.7.0", true, "v1.7.0 is the latest release already"},
			// v1.8.0 held as a pre-release is not among the full releases.
			{"v1.7.0", "v1.7.1", true, "v1.7.0 → v1.7.1"},
			{"v1.8.0\nv1.7.0", "v1.7.1", false, "v1.8.0 is the latest release"},
			{"v1.10.0\nv1.9.0", "v1.9.0", false, "v1.10.0 is the latest release"},
			// A pre-release tag published as a full release by hand is no floor.
			{"v1.5.1\nv1.8.0-rc.1", "v1.7.0", true, "v1.5.1 → v1.7.0"},
			{"chain/onyx", "v1.7.0", false, "has no full release to move forward from"},
			// GitHub lists releases newest first, not in version order.
			{"v1.5.2\nv1.7.0", "v1.6.0", false, "v1.7.0 is the latest release"},
		} {
			t.Run(c.version+" over "+strings.ReplaceAll(c.releases, "\n", ","), func(t *testing.T) {
				out, returned := runScriptFuncs(t, promoteReleaseScript,
					map[string]string{"VERSION": c.version, "RELEASES": c.releases, "VERSION_RE": shape, "REPO": "gnolang/gno"},
					"pick_latest", "check_moves_forward")
				assert.Equal(t, c.moves, returned, "%s over %q:\n%s", c.version, c.releases, out)
				assert.Contains(t, out, c.says, "%s over %q", c.version, c.releases)
			})
		}
	})

	// The digest read from a header is what the manifest bytes are checked
	// against, and the media type goes back to the registry with them.
	t.Run("headers are read whole, whatever the protocol's case and line ends", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		for _, c := range []struct{ name, dump, header, want string }{
			{
				"http2", "HTTP/2 200\r\ncontent-type: application/vnd.oci.image.index.v1+json\r\ndocker-content-digest: sha256:aa\r\n\r\n",
				"docker-content-digest", "sha256:aa",
			},
			{
				"http1", "HTTP/1.1 200 OK\r\nDocker-Content-Digest: sha256:bb\r\nContent-Type: application/json\r\n\r\n",
				"docker-content-digest", "sha256:bb",
			},
			{
				"parameters", "HTTP/2 200\r\ncontent-type: application/json; charset=utf-8\r\n\r\n",
				"content-type", "application/json; charset=utf-8",
			},
			{
				"missing", "HTTP/2 200\r\ncontent-type: application/json\r\n\r\n",
				"docker-content-digest", "",
			},
		} {
			t.Run(c.name, func(t *testing.T) {
				dump := filepath.Join(dir, c.name)
				require.NoError(t, os.WriteFile(dump, []byte(c.dump), 0o600))
				// The path reaches bash through the environment, never spliced
				// into its source.
				out, returned := runScriptFuncs(t, promoteReleaseScript, map[string]string{"DUMP": dump}, "header "+c.header+` "$DUMP"`)
				require.True(t, returned, "header on %s:\n%s", c.name, out)
				assert.Equal(t, c.want+"\n", out, "header %s in the %s dump", c.header, c.name)
			})
		}
	})
}
