package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func newTestMockIO(stdin string) commands.IO {
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	io.SetOut(commands.WriteNopCloser(os.Stdout))
	io.SetErr(commands.WriteNopCloser(os.Stderr))
	return io
}

func newTestMockIOWithStderr(stdin string) (commands.IO, *strings.Builder) {
	var buf strings.Builder
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	io.SetOut(commands.WriteNopCloser(os.Stdout))
	io.SetErr(commands.WriteNopCloser(&buf))
	return io, &buf
}

func TestInsertPathLetter(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		kind    moduleKind
		want    string
		wantErr bool
	}{
		{"realm", "gno.land/myname/myrealm", kindRealm, "gno.land/r/myname/myrealm", false},
		{"package", "gno.land/myname/mypkg", kindPackage, "gno.land/p/myname/mypkg", false},
		{"deep path", "gno.land/myname/sub/deep", kindRealm, "gno.land/r/myname/sub/deep", false},
		{"no slash", "gno.land", kindRealm, "", true},
		{"idempotent realm", "gno.land/r/myname/myrealm", kindRealm, "gno.land/r/myname/myrealm", false},
		{"idempotent package", "gno.land/p/myname/mypkg", kindPackage, "gno.land/p/myname/mypkg", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := insertPathLetter(tt.path, tt.kind)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPromptModuleKind(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    moduleKind
		wantErr bool
	}{
		{"r", "r\n", kindRealm, false},
		{"realm", "realm\n", kindRealm, false},
		{"p", "p\n", kindPackage, false},
		{"package", "package\n", kindPackage, false},
		{"empty default", "\n", kindPackage, false},
		{"m", "m\n", kindRun, false},
		{"main", "main\n", kindRun, false},
		{"run", "run\n", kindRun, false},
		{"invalid", "xyz\n", kindPackage, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := promptModuleKind(newTestMockIO(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSelectTemplate(t *testing.T) {
	single := []initTemplate{
		{Name: "basic", Description: "test"},
	}
	multi := []initTemplate{
		{Name: "basic", Description: "basic desc"},
		{Name: "dao", Description: "dao desc"},
	}

	t.Run("single auto-selects", func(t *testing.T) {
		got, err := selectTemplate(single, newTestMockIO(""))
		require.NoError(t, err)
		require.Equal(t, "basic", got.Name)
	})

	t.Run("multi default", func(t *testing.T) {
		got, err := selectTemplate(multi, newTestMockIO("\n"))
		require.NoError(t, err)
		require.Equal(t, "basic", got.Name)
	})

	t.Run("multi by number", func(t *testing.T) {
		got, err := selectTemplate(multi, newTestMockIO("2\n"))
		require.NoError(t, err)
		require.Equal(t, "dao", got.Name)
	})

	t.Run("multi by name", func(t *testing.T) {
		got, err := selectTemplate(multi, newTestMockIO("dao\n"))
		require.NoError(t, err)
		require.Equal(t, "dao", got.Name)
	})

	t.Run("invalid choice", func(t *testing.T) {
		_, err := selectTemplate(multi, newTestMockIO("99\n"))
		require.Error(t, err)
	})
}

func TestKindFromPath(t *testing.T) {
	require.Equal(t, kindRealm, kindFromPath("gno.land/r/demo/myrealm"))
	require.Equal(t, kindPackage, kindFromPath("gno.land/p/demo/mypkg"))
	require.Equal(t, kindPackage, kindFromPath("gno.land/x/something"))
}

func TestRenderTemplateDir(t *testing.T) {
	data := templateData{PkgName: "myrealm"}

	t.Run("realm", func(t *testing.T) {
		files, err := renderTemplateDir(realmTemplatesFS, "templates/realm/basic", data)
		require.NoError(t, err)
		src, ok := files["myrealm.gno"]
		require.True(t, ok, "expected myrealm.gno in output")
		require.Contains(t, string(src), "package myrealm")
		require.Contains(t, string(src), "func Render")

		test, ok := files["myrealm_test.gno"]
		require.True(t, ok, "expected myrealm_test.gno in output")
		require.Contains(t, string(test), "package myrealm")
	})

	t.Run("run", func(t *testing.T) {
		data := templateData{PkgName: "main", ScriptName: "create_proposal", ScriptPath: "run/create_proposal.gno"}
		files, err := renderTemplateDir(runTemplatesFS, "templates/run/basic", data)
		require.NoError(t, err)
		src, ok := files["create_proposal.gno"]
		require.True(t, ok, "expected create_proposal.gno in output")
		require.Contains(t, string(src), "package main")
		require.Contains(t, string(src), "func main()")
		require.Contains(t, string(src), "./run/create_proposal.gno")
	})
}

func TestPromptModulePath(t *testing.T) {
	t.Run("accept default name", func(t *testing.T) {
		got, err := promptModulePath(kindRealm, "/tmp/myrealm", newTestMockIO("myuser\n\n"))
		require.NoError(t, err)
		require.Equal(t, "gno.land/r/myuser/myrealm", got)
	})

	t.Run("custom name", func(t *testing.T) {
		got, err := promptModulePath(kindRealm, "/tmp/myrealm", newTestMockIO("myuser\ncustom\n"))
		require.NoError(t, err)
		require.Equal(t, "gno.land/r/myuser/custom", got)
	})

	t.Run("package kind", func(t *testing.T) {
		got, err := promptModulePath(kindPackage, "/home/alice/mylib", newTestMockIO("alice\n\n"))
		require.NoError(t, err)
		require.Equal(t, "gno.land/p/alice/mylib", got)
	})

	t.Run("empty namespace retries then EOF", func(t *testing.T) {
		mockIO, stderrBuf := newTestMockIOWithStderr("\n")
		_, err := promptModulePath(kindRealm, "/tmp/myrealm", mockIO)
		require.Error(t, err)
		require.Contains(t, stderrBuf.String(), "value cannot be empty")
	})

	t.Run("bech32 address namespace", func(t *testing.T) {
		got, err := promptModulePath(kindRealm, "/tmp/myrealm", newTestMockIO("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5\n\n"))
		require.NoError(t, err)
		require.Equal(t, "gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/myrealm", got)
	})
}

func TestValidateNamespace(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"valid name", "alice", false},
		{"valid underscore name", "_foo", false},
		{"valid bech32 address", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", false},
		{"uppercase invalid", "Alice", true},
		{"dash invalid", "al-ice", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNamespace(tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResolveTemplate(t *testing.T) {
	templates := []initTemplate{
		{Name: "basic", Description: "basic desc"},
		{Name: "dao", Description: "dao desc"},
	}

	t.Run("empty name returns first", func(t *testing.T) {
		got, err := resolveTemplate(templates, "")
		require.NoError(t, err)
		require.Equal(t, "basic", got.Name)
	})

	t.Run("exact match", func(t *testing.T) {
		got, err := resolveTemplate(templates, "dao")
		require.NoError(t, err)
		require.Equal(t, "dao", got.Name)
	})

	t.Run("case insensitive", func(t *testing.T) {
		got, err := resolveTemplate(templates, "DAO")
		require.NoError(t, err)
		require.Equal(t, "dao", got.Name)
	})

	t.Run("unknown name", func(t *testing.T) {
		_, err := resolveTemplate(templates, "nope")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown template")
		require.Contains(t, err.Error(), "basic, dao")
	})

	t.Run("empty list", func(t *testing.T) {
		_, err := resolveTemplate(nil, "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "no templates")
	})
}

func TestValidateGnoPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid", "run/hello.gno", false},
		{"valid simple", "hello.gno", false},
		{"valid nested", "sub/dir/hello.gno", false},
		{"dotdot prefix not traversal", "..bar/hello.gno", false},
		{"absolute", "/tmp/hack.gno", true},
		{"traversal", "../escape.gno", true},
		{"traversal prefix", "../../escape.gno", true},
		{"empty name", ".gno", true},
		{"invalid chars", "run/Hello.gno", true},
		{"dots in name", "run/hello.world.gno", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateGnoPath(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSanitizeModuleName(t *testing.T) {
	require.Equal(t, "gno_fix_mod_init_template", sanitizeModuleName("gno-fix-mod-init-template"))
	require.Equal(t, "my_realm", sanitizeModuleName("My-Realm"))
	require.Equal(t, "simple", sanitizeModuleName("simple"))
	require.Equal(t, "has_123", sanitizeModuleName("has-123"))
	require.Equal(t, "nodots", sanitizeModuleName("no.dots"))
}

func TestNormalizeModulePath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"p/nt/hello", "gno.land/p/nt/hello"},
		{"r/demo/foo", "gno.land/r/demo/foo"},
		{"gno.land/p/nt/hello", "gno.land/p/nt/hello"},
		{"gno.land/r/demo/foo", "gno.land/r/demo/foo"},
		{"example.com/x/y", "example.com/x/y"},
		{"nt/hello", "nt/hello"},
		{"", ""},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, normalizeModulePath(tt.in), "input: %q", tt.in)
	}
}

// TestWriteRunScriptNoOrphanDir verifies that if template rendering fails,
// no parent directory (e.g. "run/") is left behind on disk. Regression test
// for the "render before mkdir" ordering in writeRunScript.
//
// This is a Go-level test (rather than a txtar scenario) because it needs
// to inject a deliberately broken initTemplate to force renderTemplateDir
// to fail — something that can't be triggered through the CLI surface.
func TestWriteRunScriptNoOrphanDir(t *testing.T) {
	tmpDir := t.TempDir()

	// initTemplate pointing at a non-existent directory inside the embed FS —
	// renderTemplateDir will fail before any filesystem side effect.
	bogus := initTemplate{
		Name: "bogus",
		Dir:  "templates/run/doesnotexist",
		FS:   runTemplatesFS,
	}
	err := writeRunScript(tmpDir, "run/hello.gno", "hello", bogus, newTestMockIO(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "render run template")

	_, err = os.Stat(filepath.Join(tmpDir, "run"))
	require.True(t, os.IsNotExist(err), "run/ must not be created when template rendering fails")
}
