package gnoland

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The Linux binaries release / chain-tag publishes must start on glibc 2.35:
// Ubuntu 22.04's, and below Debian 12's 2.36. A binary built with cgo needs
// at least the glibc of the runner that built it, and ubuntu-latest moves on
// by itself: v1.7.0's gnoland, gnokey and gnodev, built on Ubuntu 24.04,
// needed 2.38 and did not start on either.
func TestReleaseLinuxBinariesStartOnGlibc235(t *testing.T) {
	t.Parallel()

	t.Run("linux/amd64 is built on ubuntu-22.04", func(t *testing.T) {
		t.Parallel()

		// A native build, so cgo is on and the runner's glibc is the floor.
		runner := grepOne(t, releaseChainTagWorkflow, `(?m)^ +- os: (\S+)\n +goos: linux\n +goarch: amd64$`)
		assert.Equal(t, "ubuntu-22.04", runner)
	})

	t.Run("a step refuses a Linux binary that needs a newer glibc", func(t *testing.T) {
		t.Parallel()

		const name = "      - name: Assert the Linux binaries start on glibc 2.35\n"
		step := grepOne(t, releaseChainTagWorkflow, `(?ms)^`+name+`(.*?)\n\n`)
		assert.Contains(t, step, "if: matrix.goos == 'linux'")
		assert.Contains(t, step, "MAX_GLIBC: GLIBC_2.35")
		assert.Contains(t, step, "readelf")
		// Before the upload, so such a binary never reaches the release.
		wf := repoFile(t, releaseChainTagWorkflow)
		assert.Less(t, strings.Index(wf, name), strings.Index(wf, "actions/upload-artifact"))
	})
}
