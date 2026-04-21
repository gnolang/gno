package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

// newTestMockIO creates a commands.IO with the given string as stdin,
// stdout discarded, and stderr captured in the returned builder (if non-nil).
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

func TestModApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"mod"},
			errShouldBe: "flag: help requested",
		},

		// test `gno mod download`
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldContain:     "gnowork.toml file not found in current or any parent directory",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_workspace",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: warning: \"./...\" matched no packages\n",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "1 build error(s)",
			stderrShouldContain:  "invalid gnomod.toml: 'module' is required",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/invalid_module_name",
			simulateExternalRepo: true,
			errShouldBe:          "1 build error(s)",
			stderrShouldContain:  "invalid gnomod.toml: 'module' is required (type: *errors.errorString)",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_invalid_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
			errShouldBe:          "1 build error(s)",
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/require_std_lib",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace_with_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "download"},
			testDir:              "../../tests/integ/replace_with_module",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0",
		},
		// TODO: that functionality is not available on gnomod.toml anymore. should we remove this?
		// {
		// 	args:                 []string{"mod", "download"},
		// 	testDir:              "../../tests/integ/replace_with_invalid_module",
		// 	simulateExternalRepo: true,
		// 	stderrShouldContain:  "gno: downloading gno.land/p/demo/notexists",
		// 	errShouldContain:     "query files list for pkg \"gno.land/p/demo/notexists\": package \"gno.land/p/demo/notexists\" is not available",
		// },

		// test `gno init` with module name
		{
			args:                 []string{"init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno1",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno2",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gno3",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"init", "gno.land/p/demo/foo"},
			testDir:              "../../tests/integ/empty_gnomod",
			simulateExternalRepo: true,
			errShouldBe:          "gnomod.toml already exists",
		},

		// test `gno mod tidy`
		{
			args:                 []string{"mod", "tidy", "arg1"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			errShouldContain:     "flag: help requested",
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldContain:     "gnomod.toml doesn't exist",
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"mod", "tidy"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
		},

		// test `gno mod why`
		{
			args:                 []string{"mod", "why"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			errShouldContain:     "flag: help requested",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			errShouldContain:     "gnomod.toml doesn't exist",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/invalid_gno_file",
			simulateExternalRepo: true,
			errShouldContain:     "expected 'package', found packag",
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/t/minim does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/t/importavl does not need package std)
`,
		},
		{
			args:                 []string{"mod", "why", "std", "gno.land/p/nt/avl/v0"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
			stdoutShouldBe: `# std
(module gno.land/p/integ/valid does not need package std)

# gno.land/p/nt/avl/v0
valid.gno
`,
		},

		// test `gno mod graph`
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
			stdoutShouldBe:       ``,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid1",
			simulateExternalRepo: true,
			stdoutShouldBe: `gno.vm/r/tests/integ/valid1 testing
`,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid2",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/p/integ/valid gno.land/p/integ/valid
gno.land/p/integ/valid gno.land/p/nt/avl/v0
gno.land/p/integ/valid testing
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
		{
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/require_remote_module",
			simulateExternalRepo: true,
			stderrShouldBe:       "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/t/importavl gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
		{
			// gno.land/p/nt/avl/v0 is included from the test in the filetests subdir
			args:                 []string{"mod", "graph"},
			testDir:              "../../tests/integ/valid3",
			simulateExternalRepo: true,
			stderrShouldContain:  "gno: downloading gno.land/p/nt/avl/v0\n",
			stdoutShouldBe: `gno.land/p/integ/valid3 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/avl/v0
gno.land/p/nt/avl/v0 gno.land/p/nt/ufmt/v0
gno.land/p/nt/avl/v0 sort
gno.land/p/nt/avl/v0 strings
gno.land/p/nt/avl/v0 testing
`,
		},
	}

	testMainCaseRun(t, tc)
}

func TestModInitNonInteractive(t *testing.T) {
	tests := []struct {
		name    string
		modPath string
	}{
		{"realm", "gno.land/r/demo/myrealm"},
		{"package", "gno.land/p/demo/mypkg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			origDir, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			t.Cleanup(func() { os.Chdir(origDir) })

			mockIO := newTestMockIO("")
			cfg := &modInitCfg{bare: false}
			// In test, IsInteractive() is false, so only gnomod.toml is created.
			err = execModInit(cfg, []string{tt.modPath}, mockIO)
			require.NoError(t, err)

			content, err := os.ReadFile(filepath.Join(tmpDir, "gnomod.toml"))
			require.NoError(t, err)
			require.Contains(t, string(content), tt.modPath)

			// No template files in non-TTY mode
			pkgName := filepath.Base(tt.modPath)
			_, err = os.Stat(filepath.Join(tmpDir, pkgName+".gno"))
			require.True(t, os.IsNotExist(err), "template files should not be created in non-TTY mode")
		})
	}
}

func TestModInitBare(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{bare: true}
	err = execModInit(cfg, []string{"gno.land/r/demo/testrealm"}, newTestMockIO(""))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "gnomod.toml"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDir, "testrealm.gno"))
	require.True(t, os.IsNotExist(err), "--bare should not create template files")
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
		data := templateData{PkgName: "main", ScriptName: "create_proposal"}
		files, err := renderTemplateDir(runTemplatesFS, "templates/run/basic", data)
		require.NoError(t, err)
		src, ok := files["create_proposal.gno"]
		require.True(t, ok, "expected create_proposal.gno in output")
		require.Contains(t, string(src), "package main")
		require.Contains(t, string(src), "func main()")
		require.Contains(t, string(src), "./create_proposal.gno")
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
		require.Error(t, err) // EOF after retry
		require.Contains(t, stderrBuf.String(), "value cannot be empty")
	})
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

func TestModInitWithTemplateFlag(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{template: "basic"}
	err = execModInit(cfg, []string{"gno.land/r/demo/testrealm"}, newTestMockIO(""))
	require.NoError(t, err)

	// gnomod.toml should exist
	content, err := os.ReadFile(filepath.Join(tmpDir, "gnomod.toml"))
	require.NoError(t, err)
	require.Contains(t, string(content), "gno.land/r/demo/testrealm")

	// Template files should NOT be created in non-TTY mode (bare path)
	// but --template + arg should create them... however IsInteractive()
	// returns false in tests, so this goes through the bare path.
	// The --template flag only affects the interactive path.
}

func TestModInitBareAndTemplateExclusive(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{bare: true, template: "basic"}
	err = execModInit(cfg, []string{"gno.land/r/demo/testrealm"}, newTestMockIO(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "mutually exclusive")
}

func TestModInitGnoFile(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{}
	err = execModInit(cfg, []string{"run/create_proposal.gno"}, newTestMockIO(""))
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "run", "create_proposal.gno"))
	require.NoError(t, err)
	require.Contains(t, string(content), "package main")
	require.Contains(t, string(content), "./create_proposal.gno")

	// No gnomod.toml for run scripts
	_, err = os.Stat(filepath.Join(tmpDir, "gnomod.toml"))
	require.True(t, os.IsNotExist(err))
}

func TestModInitGnoFileConflict(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{}
	err = execModInit(cfg, []string{"run/hello.gno"}, newTestMockIO(""))
	require.NoError(t, err)

	err = execModInit(cfg, []string{"run/hello.gno"}, newTestMockIO(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "file already exists")
}

func TestModInitBareAndGnoExclusive(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg := &modInitCfg{bare: true}
	err = execModInit(cfg, []string{"run/hello.gno"}, newTestMockIO(""))
	require.Error(t, err)
	require.Contains(t, err.Error(), "mutually exclusive")
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
		{"nt/hello", "nt/hello"}, // no p/ or r/ prefix, left as-is
		{"", ""},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, normalizeModulePath(tt.in), "input: %q", tt.in)
	}
}
