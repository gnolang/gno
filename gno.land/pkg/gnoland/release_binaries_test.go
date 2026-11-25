package gnoland

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const glibcStepName = "      - name: Assert the Linux binaries start on glibc 2.35"

// glibcStep returns the lines of release / chain-tag's glibc step.
func glibcStep(t *testing.T) []string {
	t.Helper()

	return workflowStep(t, glibcStepName)
}

// workflowStep returns the lines of release / chain-tag's step with the given
// name line, from that line up to the next step or job.
func workflowStep(t *testing.T, name string) []string {
	t.Helper()

	lines := strings.Split(repoFile(t, releaseChainTagWorkflow), "\n")
	for i, line := range lines {
		if line != name {
			continue
		}
		end := i + 1
		for end < len(lines) && !strings.HasPrefix(lines[end], "      - ") &&
			(lines[end] == "" || strings.HasPrefix(lines[end], "       ")) {
			end++
		}
		return lines[i:end]
	}
	t.Fatalf("%s: no step %q", releaseChainTagWorkflow, strings.TrimSpace(name))
	return nil
}

// installStep returns the lines of release / chain-tag's step that installs
// the container's build tools.
func installStep(t *testing.T) []string {
	t.Helper()

	return workflowStep(t, "      - name: Install the build tools in the container")
}

// glibcScript returns the glibc step's run script, without its YAML
// indentation.
func glibcScript(t *testing.T) string {
	t.Helper()

	step := glibcStep(t)
	for i, line := range step {
		if line != "        run: |" {
			continue
		}
		var script []string
		for _, body := range step[i+1:] {
			script = append(script, strings.TrimPrefix(body, "          "))
		}
		return strings.Join(script, "\n")
	}
	t.Fatalf("%s: the glibc step has no run script", releaseChainTagWorkflow)
	return ""
}

// readelfStub prints a fixture's content as its readelf output, and fails on
// a fixture that starts with NOT-ELF, as readelf does on a file that is not
// an ELF. The test runs on macOS too, where no readelf is installed.
const readelfStub = `#!/bin/sh
for f; do :; done
if grep -q '^NOT-ELF' "$f"; then
	echo "readelf: Error: Not an ELF file - it has the wrong magic bytes at the start" >&2
	exit 1
fi
cat "$f"
`

// readelfDynamic is readelf's output for a dynamic binary that needs the
// given glibc symbol versions.
func readelfDynamic(versions ...string) string {
	var out strings.Builder
	out.WriteString("Program Headers:\n  INTERP 0x318 0x400318 0x400318 0x1c 0x1c R 0x1\n" +
		"      [Requesting program interpreter: /lib64/ld-linux-x86-64.so.2]\n" +
		"Version needs section '.gnu.version_r' contains 1 entry:\n" +
		" Addr: 0x0  Offset: 0x0  Link: 6 (.dynstr)\n  000000: Version: 1  File: libc.so.6  Cnt: 3\n")
	for _, v := range versions {
		out.WriteString("  0x0010:   Name: " + v + "  Flags: none  Version: 2\n")
	}
	return out.String()
}

const readelfStatic = "Program Headers:\n  LOAD 0x0 0x400000 0x400000 0x1000 0x1000 R E 0x1000\n\nNo version information found in this file.\n"

// runGlibcStep runs the glibc step's script as GitHub Actions runs a step
// with shell: bash, in a checkout whose dist/ holds the given files, with
// MAX_GLIBC as the step sets it, and with a stub readelf only if withReadelf.
// It returns the output and whether the step passed.
func runGlibcStep(t *testing.T, files map[string]string, withReadelf bool) (string, bool) {
	t.Helper()

	checkout := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(checkout, "dist"), 0o755))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(checkout, "dist", name), []byte(content), 0o755))
	}
	// Only the tools the script needs, so a readelf installed on the host
	// cannot stand in for a missing one.
	tools := t.TempDir()
	for _, tool := range []string{"grep", "sort", "tail", "cat", "sh"} {
		p, err := exec.LookPath(tool)
		require.NoError(t, err)
		require.NoError(t, os.Symlink(p, filepath.Join(tools, tool)))
	}
	if withReadelf {
		require.NoError(t, os.WriteFile(filepath.Join(tools, "readelf"), []byte(readelfStub), 0o755))
	}
	script := filepath.Join(t.TempDir(), "step.sh")
	require.NoError(t, os.WriteFile(script, []byte(glibcScript(t)), 0o644))

	bash, err := exec.LookPath("bash")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, bash, "--noprofile", "--norc", "-eo", "pipefail", script)
	cmd.Dir = checkout
	cmd.Env = []string{"PATH=" + tools, "MAX_GLIBC=GLIBC_2.35"}
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("could not run bash: %v", err)
	}
	return string(out), err == nil
}

// The Linux binaries release / chain-tag publishes must start on glibc 2.35,
// Ubuntu 22.04's. A binary built with cgo needs at most the glibc of the
// system that built it, and may need exactly that: v1.7.0's gnoland, gnokey
// and gnodev, built on Ubuntu 24.04 (glibc 2.39), needed 2.38 and did not
// start on Ubuntu 22.04 or Debian 12.
func TestReleaseLinuxBinariesStartOnGlibc235(t *testing.T) {
	t.Parallel()

	t.Run("linux/amd64 is built with cgo in an ubuntu:22.04 container", func(t *testing.T) {
		t.Parallel()

		image := grepOne(t, releaseChainTagWorkflow, `(?m)^ +container: (\S+)\n +goos: linux\n +goarch: amd64$`)
		assert.Regexp(t, `^ubuntu:22\.04@sha256:[0-9a-f]{64}$`, image, "linux/amd64 builds in ubuntu:22.04, pinned by digest")
		wf := repoFile(t, releaseChainTagWorkflow)
		assert.Contains(t, wf, "\n    container: ${{ matrix.container }}\n", "the binaries job runs in the row's container")
		// Go turns cgo off by itself when it finds no C compiler, which would
		// drop lmdb, mdbx and Ledger support without failing anything.
		install := installStep(t)
		assert.Contains(t, install, "        if: matrix.container", "the container row installs its compiler")
		assert.Contains(t, install, `          echo "CGO_ENABLED=1" >> "${GITHUB_ENV}"`, "the container's build requires cgo")
		// The job runs as root in a workspace the runner's user owns: without
		// this, git refuses it, and go build fails on its VCS stamping.
		assert.Contains(t, install, `          git config --system --add safe.directory "${GITHUB_WORKSPACE}"`, "git must accept the workspace")
	})

	t.Run("the glibc step runs on every Linux row, before the upload, and fails the run", func(t *testing.T) {
		t.Parallel()

		step := glibcStep(t)
		assert.Contains(t, step, "        if: matrix.goos == 'linux'")
		assert.Contains(t, step, "        shell: bash", "pipefail: a readelf failure must fail the step")
		assert.Contains(t, step, "          MAX_GLIBC: GLIBC_2.35")
		for _, line := range step {
			assert.NotContains(t, line, "continue-on-error", "the glibc step must fail the run")
		}
		wf := repoFile(t, releaseChainTagWorkflow)
		assert.Less(t, strings.Index(wf, glibcStepName), strings.Index(wf, "actions/upload-artifact"),
			"the glibc step must run before the binaries are uploaded")
	})

	for _, c := range []struct {
		name        string
		files       map[string]string
		withReadelf bool
		passes      bool
		says        string
	}{
		{"binaries needing 2.35 or less, a static one and the checksums", map[string]string{
			"gnoland_linux_amd64":       readelfDynamic("GLIBC_2.2.5", "GLIBC_2.34"),
			"gno_linux_amd64":           readelfDynamic("GLIBC_2.2.5", "GLIBC_2.35", "GLIBC_2.17"),
			"gnoland_linux_arm64":       readelfStatic,
			"checksums_linux_amd64.txt": "NOT-ELF checksums",
		}, true, true, "dist/gno_linux_amd64: needs GLIBC_2.35"},
		{"a binary needing 2.36", map[string]string{
			"gnoland_linux_amd64": readelfDynamic("GLIBC_2.2.5", "GLIBC_2.36", "GLIBC_2.34"),
		}, true, false, "needs GLIBC_2.36, newer than GLIBC_2.35"},
		{"v1.7.0's gnoland, needing 2.38", map[string]string{
			"gnoland_linux_amd64": readelfDynamic("GLIBC_2.2.5", "GLIBC_2.38", "GLIBC_2.34"),
			"gno_linux_amd64":     readelfDynamic("GLIBC_2.34"),
		}, true, false, "needs GLIBC_2.38, newer than GLIBC_2.35"},
		{"a file readelf cannot read", map[string]string{
			"gnoland_linux_amd64": "NOT-ELF",
		}, true, false, "Not an ELF file"},
		{"no readelf", map[string]string{
			"gnoland_linux_amd64": readelfDynamic("GLIBC_2.34"),
		}, false, false, "readelf"},
		{"no Linux binary at all", map[string]string{}, true, false, "no Linux binary in dist/"},
		{"a dynamic binary with no glibc version", map[string]string{
			"gnoland_linux_amd64": readelfDynamic(),
		}, true, false, "dynamic, but needs no glibc symbol version"},
	} {
		t.Run("the glibc step on "+c.name, func(t *testing.T) {
			t.Parallel()

			out, ok := runGlibcStep(t, c.files, c.withReadelf)
			assert.Equal(t, c.passes, ok, out)
			assert.Contains(t, out, c.says)
		})
	}
}
