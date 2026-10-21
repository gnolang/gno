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

func TestDetectAndRevision(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run("init", "-q")
	run("remote", "add", "origin", "git@github.com:someone/contracts.git")
	pkg := filepath.Join(root, "r", "demo")
	require.NoError(t, os.MkdirAll(pkg, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "demo.gno"), []byte("package demo\n"), 0o644))

	src, err := Detect(pkg)
	require.NoError(t, err)
	assert.Equal(t, gnomod.Source{Repository: "https://github.com/someone/contracts", Path: "r/demo"}, src)

	root2, err := Detect(root)
	require.NoError(t, err)
	assert.Equal(t, "", root2.Path)
}
