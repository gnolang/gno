package main

import (
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/mod/module"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

var testPackageFetcher pkgdownload.PackageFetcher

func newModCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "mod",
			ShortUsage: "mod <command>",
			ShortHelp:  "module maintenance",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModDownloadCmd(io),
		newModGraphCmd(io),
		newModInitLegacyCmd(io),
		newModTidy(io),
		newModWhy(io),
	)

	return cmd
}

// newModInitLegacyCmd registers `gno mod init` as a thin legacy alias that
// preserves the original behavior: create a bare gnomod.toml in CWD. It
// never triggers the interactive wizard, and hints at `gno init` for users
// who want the richer scaffolding flow.
func newModInitLegacyCmd(io commands.IO) *commands.Command {
	cfg := &modInitCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [<module-path>]",
			ShortHelp:  "create a bare gnomod.toml (see 'gno init' for templates)",
			LongHelp: `Create a bare gnomod.toml in the current directory.

For interactive scaffolding with template files, use 'gno init' instead.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			io.ErrPrintln("hint: 'gno init' scaffolds a module with template files interactively")
			// Always bare — never trigger the wizard.
			cfg.bare = true
			cfg.template = ""
			return execModInit(cfg, args, io)
		},
	)
}

func newModDownloadCmd(io commands.IO) *commands.Command {
	cfg := &modDownloadCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "download",
			ShortUsage: "download [flags]",
			ShortHelp:  "download modules to local cache",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModDownload(cfg, args, io)
		},
	)
}

func newModGraphCmd(io commands.IO) *commands.Command {
	cfg := &modGraphCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "graph",
			ShortUsage: "graph [path]",
			ShortHelp:  "print module requirement graph",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModGraph(cfg, args, io)
		},
	)
}

// --- gno init (top-level command) ---

// moduleKind enumerates the flavors of module gno init can scaffold.
type moduleKind int

const (
	kindPackage moduleKind = iota // stateless library under /p/
	kindRealm                     // stateful contract under /r/
	kindRun                       // one-off main script executed via gnokey maketx run
)

// runScriptDir is the default subdirectory where `gno init` places run scripts
// created by the interactive wizard (e.g. "run/hello.gno").
const runScriptDir = "run"

// modInitCfg holds the flags shared by `gno init` and the legacy `gno mod init` alias.
type modInitCfg struct {
	bare     bool   // when true, only write gnomod.toml (no template files)
	template string // explicit template name; empty means "use default / prompt"
}

// RegisterFlags implements the commands.Config interface for modInitCfg.
func (c *modInitCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.bare, "bare", false, "only create gnomod.toml, skip template files")
	fs.StringVar(&c.template, "template", "", "template name to use (e.g. basic); skips interactive selection")
}

// newInitCmd returns the top-level `gno init` command: an interactive wizard
// (or argument-driven non-interactive flow) that scaffolds a Gno module in CWD.
func newInitCmd(io commands.IO) *commands.Command {
	cfg := &modInitCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags] [<module-path>]",
			ShortHelp:  "initialize a new Gno module",
			LongHelp: `Initialize a new Gno module in the current directory.

When run in an interactive terminal with no arguments, a wizard guides you through:
  1. Module kind selection (realm, package, or run script)
  2. Namespace and module name
  3. Template selection (when multiple templates are available)

If a module path is given as an argument, the kind is auto-detected from the
path (/r/ for realms, /p/ for packages) and the first available template is
used. Short-form paths like "r/demo/foo" are expanded to "gno.land/r/demo/foo".

If the argument ends in .gno, a run script is created at that path without
a gnomod.toml (e.g. "gno init run/hello.gno" creates run/hello.gno).
Directories in the path are created as needed. This is the only case where
'gno init' creates directories.

Flags:
  --bare       Create only gnomod.toml, skip all template files.
  --template   Select a template by name (e.g. --template basic), skipping
               the interactive template menu. Use "gno init --help" to see
               the available templates for each module kind.

Available templates:
  Realms:   basic
  Packages: basic
  Run:      basic

Examples:
  gno init                              # interactive wizard
  gno init gno.land/r/myname/myrealm   # realm in CWD
  gno init r/myname/mypkg              # short form, auto-expanded
  gno init --bare gno.land/p/demo/lib  # gnomod.toml only
  gno init --template basic gno.land/r/demo/foo  # explicit template
  gno init run/create_proposal.gno     # run script, no gnomod.toml`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModInit(cfg, args, io)
		},
	)
}

// execModInit is the shared handler for both `gno init` and `gno mod init`.
// It dispatches to the appropriate flow based on the config and arguments:
// bare, .gno run script, argument-driven scaffold, or full wizard.
func execModInit(cfg *modInitCfg, args []string, io commands.IO) error {
	if len(args) > 1 {
		return flag.ErrHelp
	}
	if cfg.bare && cfg.template != "" {
		return fmt.Errorf("--bare and --template are mutually exclusive")
	}
	if cfg.bare && len(args) == 1 && strings.HasSuffix(args[0], ".gno") {
		return fmt.Errorf("--bare and .gno script paths are mutually exclusive")
	}

	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if !filepath.IsAbs(rootDir) {
		return fmt.Errorf("create gnomod.toml: dir %q is not absolute", rootDir)
	}

	// .gno argument: create a run script at the given path, no gnomod.toml.
	if len(args) == 1 && strings.HasSuffix(args[0], ".gno") {
		scriptName, err := validateGnoPath(args[0])
		if err != nil {
			return err
		}
		tmpl, err := resolveTemplate(templatesForKind(kindRun), cfg.template)
		if err != nil {
			return err
		}
		return writeRunScript(rootDir, args[0], scriptName, *tmpl, io)
	}

	// Below: scaffold a module in CWD. The only case where 'gno init'
	// creates directories is the .gno run-script branch handled above.

	// Early check: if gnomod.toml already exists in CWD, bail out immediately.
	if _, err := os.Stat(filepath.Join(rootDir, "gnomod.toml")); err == nil {
		return fmt.Errorf("gnomod.toml already exists")
	}

	// --bare: only create gnomod.toml in CWD.
	if cfg.bare {
		if len(args) == 0 {
			return fmt.Errorf("module path is required with --bare")
		}
		modPath := normalizeModulePath(args[0])
		if err := validateModulePath(modPath); err != nil {
			return err
		}
		if err := writeGnomod(rootDir, modPath); err != nil {
			return err
		}
		fmt.Fprintf(io.Err(), "Initialized module %s\n", modPath)
		fmt.Fprintf(io.Err(), "  gnomod.toml\n")
		printNextSteps(io, false)
		return nil
	}

	// With a module path argument: scaffold in CWD (same flow for both
	// interactive and non-interactive invocations).
	if len(args) == 1 {
		modPath := normalizeModulePath(args[0])
		tmpl, err := resolveTemplate(templatesForKind(kindFromPath(modPath)), cfg.template)
		if err != nil {
			return err
		}
		return scaffoldModule(rootDir, modPath, tmpl, io)
	}
	if !commands.IsInteractive() {
		return fmt.Errorf("module path is required (non-interactive mode)")
	}

	// Full interactive wizard.
	kind, err := promptModuleKind(io)
	if err != nil {
		return err
	}
	if kind == kindRun {
		tmpl, err := resolveTemplate(templatesForKind(kindRun), cfg.template)
		if err != nil {
			return err
		}
		return execInitRun(rootDir, *tmpl, io)
	}

	modPath, err := promptModulePath(kind, rootDir, io)
	if err != nil {
		return err
	}
	tmpl, err := resolveOrPickTemplate(templatesForKind(kind), cfg.template, io)
	if err != nil {
		return err
	}
	return scaffoldModule(rootDir, modPath, tmpl, io)
}

// resolveOrPickTemplate returns the named template, or prompts the user to
// pick one interactively when no name is given. It is only called from the
// wizard (which has already established we're in an interactive terminal).
func resolveOrPickTemplate(templates []initTemplate, name string, io commands.IO) (*initTemplate, error) {
	if name != "" {
		return resolveTemplate(templates, name)
	}
	return selectTemplate(templates, io)
}

// scaffoldModule is the shared tail of every non-bare, non-run code path:
// validate → render (with conflict check) → write gnomod.toml → write files →
// print summary. Everything is created in rootDir (CWD). Rendering fails fast
// on filename conflicts so a failed init never leaves an orphan gnomod.toml.
func scaffoldModule(rootDir, modPath string, tmpl *initTemplate, io commands.IO) error {
	if err := validateModulePath(modPath); err != nil {
		return err
	}
	files, err := renderModuleFiles(rootDir, modPath, tmpl)
	if err != nil {
		return err
	}
	if err := writeGnomod(rootDir, modPath); err != nil {
		return err
	}

	names, err := writeFiles(rootDir, files)
	if err != nil {
		return err
	}

	hasTests := false
	for _, name := range names {
		if strings.HasSuffix(name, "_test.gno") {
			hasTests = true
			break
		}
	}

	fmt.Fprintf(io.Err(), "Initialized %s %s (%s template)\n", kindLabel(modPath), modPath, tmpl.Name)
	fmt.Fprintf(io.Err(), "  gnomod.toml\n")
	for _, name := range names {
		fmt.Fprintf(io.Err(), "  %s\n", name)
	}
	printNextSteps(io, hasTests)
	return nil
}

// kindLabel returns "realm" or "package" for user-facing messages.
func kindLabel(modPath string) string {
	if gno.IsRealmPath(modPath) {
		return "realm"
	}
	return "package"
}

// writeFiles writes each entry of files under baseDir, failing if any
// destination already exists. It returns the sorted list of filenames written.
func writeFiles(baseDir string, files map[string][]byte) ([]string, error) {
	names := slices.Sorted(maps.Keys(files))
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(baseDir, name)); err == nil {
			return nil, fmt.Errorf("file already exists: %s", name)
		}
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(baseDir, name), files[name], 0o644); err != nil {
			return nil, err
		}
	}
	return names, nil
}

// printNextSteps emits a short "what to do next" hint after a successful init.
// Uses `gno test .` (not `./...`) because recursive patterns require a
// gnowork.toml, which a freshly-scaffolded single module doesn't have.
func printNextSteps(io commands.IO, hasTests bool) {
	hint := "gno test ."
	if !hasTests {
		hint = "add your code and run `gno test .`"
	}
	fmt.Fprintf(io.Err(), "Next: %s\n", hint)
}

// validateModulePath checks that modPath is syntactically valid as a Gno
// user-library path before any filesystem side effects.
func validateModulePath(modPath string) error {
	if err := module.CheckImportPath(modPath); err != nil {
		return fmt.Errorf("invalid module path: %w", err)
	}
	if !gno.IsUserlib(modPath) {
		return fmt.Errorf("invalid module path: %q is not a valid package path URL", modPath)
	}
	return nil
}

// writeGnomod writes a minimal gnomod.toml in rootDir with the given module
// path and the current Gno language version.
func writeGnomod(rootDir, modPath string) error {
	modfile := new(gnomod.File)
	modfile.Module = modPath
	modfile.Gno = gno.GnoVerLatest
	return modfile.WriteFile(filepath.Join(rootDir, "gnomod.toml"))
}

// resolveTemplate looks up a template by name from the given list.
// If name is empty, the first template is returned (default).
func resolveTemplate(templates []initTemplate, name string) (*initTemplate, error) {
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates available")
	}
	if name == "" {
		return &templates[0], nil
	}
	for i := range templates {
		if strings.EqualFold(templates[i].Name, name) {
			return &templates[i], nil
		}
	}
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name
	}
	return nil, fmt.Errorf("unknown template %q; available: %s", name, strings.Join(names, ", "))
}

// renderModuleFiles renders template files for a module and fails if any
// output file already exists on disk. It performs no writes — callers use it
// to fail-fast before touching gnomod.toml, avoiding partial state on error.
func renderModuleFiles(rootDir, modPath string, tmpl *initTemplate) (map[string][]byte, error) {
	data := templateData{PkgName: filepath.Base(modPath)}

	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	for name := range files {
		if _, err := os.Stat(filepath.Join(rootDir, name)); err == nil {
			return nil, fmt.Errorf("file already exists: %s", name)
		}
	}
	return files, nil
}

// writeRunScript creates a run script at the given relative path (e.g.
// "run/hello.gno"). No gnomod.toml is created — run scripts don't need one.
// scriptName is the base name without ".gno", already validated by the caller.
func writeRunScript(rootDir, relPath, scriptName string, tmpl initTemplate, io commands.IO) error {
	relDir := filepath.Dir(relPath)
	if err := os.MkdirAll(filepath.Join(rootDir, relDir), 0o755); err != nil {
		return err
	}

	data := templateData{PkgName: "main", ScriptName: scriptName, ScriptPath: relPath}
	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return fmt.Errorf("render run template: %w", err)
	}

	names, err := writeFiles(filepath.Join(rootDir, relDir), files)
	if err != nil {
		return err
	}

	fmt.Fprintf(io.Err(), "Initialized run script (%s template)\n", tmpl.Name)
	for _, name := range names {
		fmt.Fprintf(io.Err(), "  %s\n", filepath.Join(relDir, name))
	}
	return nil
}

// execInitRun is the wizard tail for the "run script" module kind: it prompts
// for a script name and writes the rendered run template under run/<name>.gno.
func execInitRun(rootDir string, tmpl initTemplate, io commands.IO) error {
	defaultName := sanitizeModuleName(filepath.Base(rootDir))
	if defaultName == "" {
		defaultName = "main"
	}
	scriptName, err := commands.PromptString(io, "Script name", defaultName, validateName)
	if err != nil {
		return err
	}
	relPath := filepath.Join(runScriptDir, scriptName+".gno")
	return writeRunScript(rootDir, relPath, scriptName, tmpl, io)
}

// --- Wizard helpers ---

// promptModuleKind asks the user to pick a module flavor (realm, package, or
// run script) and returns the corresponding moduleKind. Package is the default.
func promptModuleKind(io commands.IO) (moduleKind, error) {
	entries := []struct {
		key    string
		kind   moduleKind
		choice commands.Choice
	}{
		{"r", kindRealm, commands.Choice{Aliases: []string{"realm"}, Description: "realm"}},
		{"p", kindPackage, commands.Choice{Aliases: []string{"package"}, Description: "package"}},
		{"m", kindRun, commands.Choice{Aliases: []string{"main", "run"}, Description: "run script"}},
	}
	choices := make(map[string]commands.Choice, len(entries))
	for _, e := range entries {
		choices[e.key] = e.choice
	}

	key, err := commands.PromptChoice(io, "Module kind — [r]ealm, [P]ackage, or [m]ain: ", choices, "p")
	if err != nil {
		return kindPackage, err
	}
	for _, e := range entries {
		if e.key == key {
			return e.kind, nil
		}
	}
	return kindPackage, fmt.Errorf("internal: unknown module kind %q", key)
}

// promptModulePath asks the user for a namespace and module name, then
// assembles the full "gno.land/{r,p}/<namespace>/<name>" path based on kind.
// The default module name is derived from rootDir's basename.
func promptModulePath(kind moduleKind, rootDir string, io commands.IO) (string, error) {
	namespace, err := commands.PromptString(io, "Namespace or address", "", validateName)
	if err != nil {
		return "", err
	}

	defaultName := sanitizeModuleName(filepath.Base(rootDir))
	name, err := commands.PromptString(io, "Module name", defaultName, validateName)
	if err != nil {
		return "", err
	}

	return insertPathLetter(fmt.Sprintf("gno.land/%s/%s", namespace, name), kind)
}

// selectTemplate presents the wizard's template menu and returns the picked
// template. When only one template is available it is auto-selected.
func selectTemplate(templates []initTemplate, io commands.IO) (*initTemplate, error) {
	items := make([]commands.SelectItem, len(templates))
	for i, t := range templates {
		items[i] = commands.SelectItem{Name: t.Name, Description: t.Description}
	}

	name, err := commands.PromptSelect(io, "Template:", items)
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].Name == name {
			return &templates[i], nil
		}
	}
	return nil, fmt.Errorf("internal: template %q not found", name)
}

// kindFromPath infers the module kind from a fully-qualified module path:
// realm for /r/ paths, package otherwise.
func kindFromPath(modPath string) moduleKind {
	if gno.IsRealmPath(modPath) {
		return kindRealm
	}
	return kindPackage
}

// templatesForKind returns the registered template set for the given kind.
func templatesForKind(kind moduleKind) []initTemplate {
	switch kind {
	case kindRealm:
		return realmTemplates
	case kindRun:
		return runTemplates
	default:
		return packageTemplates
	}
}

// insertPathLetter inserts the kind-specific path letter ("r" for realms,
// "p" for packages) between the domain and the rest of the module path.
// It is idempotent: paths already containing /r/ or /p/ after the domain
// are returned unchanged.
func insertPathLetter(path string, kind moduleKind) (string, error) {
	domain, rest, ok := strings.Cut(path, "/")
	if !ok || rest == "" {
		return "", fmt.Errorf("invalid module path: %q", path)
	}
	if strings.HasPrefix(rest, "r/") || strings.HasPrefix(rest, "p/") {
		return path, nil
	}
	letter := "p"
	if kind == kindRealm {
		letter = "r"
	}
	return fmt.Sprintf("%s/%s/%s", domain, letter, rest), nil
}

var (
	reModName   = regexp.MustCompile(`[^a-z0-9_]`)
	reValidName = regexp.MustCompile(`^_?[a-z][a-z0-9_]*$`)
)

// isValidName reports whether s is a valid lowercase Gno identifier: letters,
// digits, and underscores, starting with a letter or a single leading underscore.
func isValidName(s string) bool {
	return reValidName.MatchString(s)
}

// validateName is a PromptString-compatible validator that enforces isValidName
// and returns a user-facing error describing the expected format.
func validateName(s string) error {
	if s == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if !isValidName(s) {
		return fmt.Errorf("invalid value %q (must be lowercase letters, digits, and underscores)", s)
	}
	return nil
}

// validateGnoPath ensures the .gno argument is a safe relative path within CWD
// and returns the derived script name (base name without the .gno suffix).
// filepath.IsLocal rejects absolute paths, empty paths, and any traversal
// above the starting directory (including on Windows) in a single check.
func validateGnoPath(p string) (string, error) {
	if !filepath.IsLocal(p) {
		return "", fmt.Errorf("invalid path %q: must be relative and within the current directory", p)
	}
	scriptName := strings.TrimSuffix(filepath.Base(p), ".gno")
	if scriptName == "" {
		return "", fmt.Errorf("script name cannot be empty")
	}
	if !isValidName(scriptName) {
		return "", fmt.Errorf("invalid script name %q (must be lowercase letters, digits, and underscores)", scriptName)
	}
	return scriptName, nil
}

// normalizeModulePath expands short-form paths ("r/demo/foo", "p/demo/lib")
// into their canonical "gno.land/..." form. Paths whose first segment already
// contains a dot (a domain) are left untouched.
func normalizeModulePath(modPath string) string {
	if modPath == "" {
		return modPath
	}
	first, _, _ := strings.Cut(modPath, "/")
	if strings.Contains(first, ".") {
		return modPath
	}
	if first == "p" || first == "r" {
		return "gno.land/" + modPath
	}
	return modPath
}

// sanitizeModuleName turns a filesystem directory name into a plausible
// default module name by lowercasing, replacing dashes with underscores, and
// stripping characters that are not valid Gno identifier bytes.
func sanitizeModuleName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	return reModName.ReplaceAllString(name, "")
}

func newModTidy(io commands.IO) *commands.Command {
	cfg := &modTidyCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "tidy",
			ShortUsage: "tidy [flags]",
			ShortHelp:  "add missing and remove unused modules",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execModTidy(cfg, args, io)
		},
	)
}

func newModWhy(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "why",
			ShortUsage: "why <package> [<package>...]",
			ShortHelp:  "Explains why modules are needed",
			LongHelp: `Explains why modules are needed.

gno mod why shows a list of files where specified packages or modules are
being used, explaining why those specified packages or modules are being
kept by gno mod tidy.

The output is a sequence of stanzas, one for each module/package name
specified, separated by blank lines. Each stanza begins with a
comment line "# module" giving the target module/package. Subsequent lines
show files that import the specified module/package, one filename per line.
If the package or module is not being used/needed/imported, the stanza
will display a single parenthesized note indicating that fact.

For example:

	$ gno mod why gno.land/p/nt/avl/v0 gno.land/p/demo/users
	# gno.land/p/nt/avl/v0
	[FILENAME_1.gno]
	[FILENAME_2.gno]

	# gno.land/p/demo/users
	(module [MODULE_NAME] does not need package gno.land/p/demo/users)
	$
`,
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execModWhy(args, io)
		},
	)
}

type modDownloadCfg struct {
	remoteOverrides string
}

const remoteOverridesArgName = "remote-overrides"

func (c *modDownloadCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remoteOverrides,
		remoteOverridesArgName,
		"",
		"chain-domain=rpc-url comma-separated list",
	)
}

type modGraphCfg struct {
	format string
}

func (c *modGraphCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.format, "format", "", "Output format, must be one of 'dot' or empty. Empty is a minimalist format.")
}

func execModGraph(cfg *modGraphCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		args = []string{"."}
	}
	if len(args) > 1 {
		return flag.ErrHelp
	}

	loadConf := packages.LoadConfig{
		Fetcher: testPackageFetcher,
		Deps:    true,
		Test:    true,
		Out:     io.Err(),
	}
	pkgs, err := packages.Load(loadConf, args...)
	if err != nil {
		return err
	}

	sb := &strings.Builder{}

	if cfg.format == "dot" {
		fmt.Fprint(sb, "digraph gno {\n")
	}

	errCount := uint(0)

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			fmt.Fprintf(io.Err(), "%s: %v", pkg.ImportPath, err)
			errCount++
		}
		deps := pkg.ImportsSpecs.Merge()
		for _, dep := range deps {
			if cfg.format == "dot" {
				fmt.Fprintf(sb, "    %q -> %q;\n", pkg.ImportPath, dep.PkgPath)
			} else {
				fmt.Fprintf(sb, "%s %s\n", pkg.ImportPath, dep.PkgPath)
			}
		}
	}

	if cfg.format == "dot" {
		fmt.Fprint(sb, "}\n")
	}

	io.Out().Write([]byte(sb.String()))

	if errCount != 0 {
		return fmt.Errorf("%d build error(s)", errCount)
	}

	return nil
}

func execModDownload(cfg *modDownloadCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	fetcher := testPackageFetcher
	if fetcher == nil {
		remoteOverrides, err := parseRemoteOverrides(cfg.remoteOverrides)
		if err != nil {
			return fmt.Errorf("invalid %s flag: %w", remoteOverridesArgName, err)
		}
		fetcher = rpcpkgfetcher.New(remoteOverrides)
	} else if len(cfg.remoteOverrides) != 0 {
		return fmt.Errorf("can't use %s flag with a custom package fetcher", remoteOverridesArgName)
	}

	loadCfg := packages.LoadConfig{
		Fetcher:    fetcher,
		Deps:       true,
		Test:       true,
		AllowEmpty: true,
		Out:        io.Err(),
	}
	pkgs, err := packages.Load(loadCfg, "./...")
	if err != nil {
		return err
	}

	errCount := uint(0)
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			fmt.Fprintf(io.Err(), "%s: %v", pkg.ImportPath, err)
			errCount++
		}
	}
	if errCount != 0 {
		return fmt.Errorf("%d build error(s)", errCount)
	}

	return nil
}

func parseRemoteOverrides(arg string) (map[string]string, error) {
	if arg == "" {
		return map[string]string{}, nil
	}
	pairs := strings.Split(arg, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected 2 parts in chain-domain=rpc-url pair %q", arg)
		}
		domain := strings.TrimSpace(parts[0])
		rpcURL := strings.TrimSpace(parts[1])
		res[domain] = rpcURL
	}
	return res, nil
}

type modTidyCfg struct {
	verbose   bool
	recursive bool
}

func (c *modTidyCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)
	fs.BoolVar(
		&c.recursive,
		"recursive",
		false,
		"walk subdirs for gno.mod files",
	)
}

func execModTidy(cfg *modTidyCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if cfg.recursive {
		pkgs, err := packages.ReadPkgListFromDir(wd, gno.MPAnyAll)
		if err != nil {
			return err
		}
		var errs error
		for _, pkg := range pkgs {
			err := modTidyOnce(cfg, wd, pkg.Dir, io)
			errs = multierr.Append(errs, err)
		}
		return errs
	}

	return modTidyOnce(cfg, wd, wd, io)
}

func modTidyOnce(cfg *modTidyCfg, wd, pkgdir string, io commands.IO) error {
	modExists := false
	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		fpath := filepath.Join(pkgdir, fname)
		if !isFileExist(fpath) {
			continue
		}

		modExists = true

		relpath, err := filepath.Rel(wd, fpath)
		if err != nil {
			return err
		}
		if cfg.verbose {
			io.ErrPrintfln("%s", relpath)
		}

		gm, err := gnomod.ParseFilepath(fpath)
		if err != nil {
			return err
		}

		if fname == "gno.mod" {
			newPath := filepath.Join(pkgdir, "gnomod.toml")
			gm.WriteFile(newPath)
		} else {
			gm.WriteFile(fpath)
		}
	}

	if !modExists {
		return gnomod.ErrNoModFile
	}

	oldpath := filepath.Join(pkgdir, "gno.mod")
	os.Remove(oldpath)

	return nil
}

func execModWhy(args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	gm, err := gnomod.ParseDir(wd)
	if err != nil {
		return err
	}

	importToFilesMap, err := getImportToFilesMap(wd)
	if err != nil {
		return err
	}

	out := formatModWhyStanzas(gm.Module, args, importToFilesMap)
	io.Printf(out)

	return nil
}

func formatModWhyStanzas(modulePath string, args []string, importToFilesMap map[string][]string) (out string) {
	for i, arg := range args {
		out += fmt.Sprintf("# %s\n", arg)
		files, ok := importToFilesMap[arg]
		if !ok {
			out += fmt.Sprintf("(module %s does not need package %s)\n", modulePath, arg)
		} else {
			for _, file := range files {
				out += file + "\n"
			}
		}
		if i < len(args)-1 {
			out += "\n"
		}
	}
	return
}

func getImportToFilesMap(pkgPath string) (map[string][]string, error) {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string)
	for _, e := range entries {
		filename := e.Name()
		if ext := filepath.Ext(filename); ext != ".gno" {
			continue
		}
		if strings.HasSuffix(filename, "_filetest.gno") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(pkgPath, filename))
		if err != nil {
			return nil, err
		}
		imports, err := packages.FileImports(filename, string(data), nil)
		if err != nil {
			return nil, err
		}

		for _, imp := range imports {
			m[imp.PkgPath] = append(m[imp.PkgPath], filename)
		}
	}
	return m, nil
}
