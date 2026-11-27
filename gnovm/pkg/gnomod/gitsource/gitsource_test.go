package gitsource

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRemote(t *testing.T) {
	cases := []struct {
		remote, want string
		wantErr      bool
	}{
		{remote: "https://github.com/gnolang/gno.git", want: "https://github.com/gnolang/gno"},
		{remote: "https://github.com/gnolang/gno", want: "https://github.com/gnolang/gno"},
		{remote: "https://github.com/gnolang/gno/", want: "https://github.com/gnolang/gno"},
		{remote: "git@github.com:gnolang/gno.git", want: "https://github.com/gnolang/gno"},
		{remote: "github.com:gnolang/gno", want: "https://github.com/gnolang/gno"},
		{remote: "ssh://git@github.com/gnolang/gno.git", want: "https://github.com/gnolang/gno"},
		{remote: "ssh://git@github.com:22/gnolang/gno.git", want: "https://github.com/gnolang/gno"},
		{remote: "https://x-access-token:ghs_secret@github.com/gnolang/gno.git", want: "https://github.com/gnolang/gno"},
		{remote: "git@gitlab.com:group/sub/repo.git", want: "https://gitlab.com/group/sub/repo"},
		{remote: "/srv/git/repo.git", wantErr: true},
		{remote: "file:///srv/git/repo.git", wantErr: true},
		{remote: "https://github.com", wantErr: true},
		{remote: "https://GitHub.com/gnolang/gno", want: "https://github.com/gnolang/gno"},
		{remote: "https://git.example.org:8443/team/repo.git", want: "https://git.example.org:8443/team/repo"},
		{remote: "https://git.example.org/team/my%20repo.git", want: "https://git.example.org/team/my%20repo"},
		{remote: `C:\repo`, wantErr: true},
		{remote: `C:/repo`, wantErr: true},
		{remote: "github-work:o/r.git", wantErr: true},
		{remote: "git@github-work:o/r.git", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.remote, func(t *testing.T) {
			got, err := NormalizeRemote(tc.remote)
			if tc.wantErr {
				assert.Error(t, err, "got %q", got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDetect(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	// Keep git from finding a repository above the temp dirs.
	t.Setenv("GIT_CEILING_DIRECTORIES", os.TempDir())

	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run("init", "-q")
	pkg := filepath.Join(root, "r", "demo")
	require.NoError(t, os.MkdirAll(pkg, 0o755))

	_, err := Detect(pkg)
	assert.ErrorContains(t, err, "origin", "no origin remote")

	run("remote", "add", "origin", "git@github.com:someone/contracts.git")
	src, err := Detect(pkg)
	require.NoError(t, err)
	assert.Equal(t, gnomod.Source{Repository: "https://github.com/someone/contracts", Path: "r/demo"}, src)

	top, err := Detect(root)
	require.NoError(t, err)
	assert.Equal(t, "", top.Path)

	// A symlink to the checkout resolves to the same path inside it.
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(root, link); err == nil {
		src, err := Detect(filepath.Join(link, "r", "demo"))
		require.NoError(t, err)
		assert.Equal(t, "r/demo", src.Path)
	}

	run("remote", "set-url", "origin", "git@github-work:someone/contracts.git")
	_, err = Detect(pkg)
	assert.ErrorContains(t, err, "not a web host")

	_, err = Detect(t.TempDir())
	assert.ErrorContains(t, err, "not a git repository", "outside a repository")
}
