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

	"golang.org/x/mod/module"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// moduleKind enumerates the flavors of module gno init can scaffold.
type moduleKind int

const (
	kindPackage moduleKind = iota // stateless library under /p/
	kindRealm                     // stateful contract under /r/
	kindRun                       // one-off main script executed via gnokey maketx run
)

// runScriptDir is the default subdirectory for wizard-generated run scripts.
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

	// .gno argument: create a run script at that path, no gnomod.toml.
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

	// From here on: scaffold a module in CWD (no directory creation).

	if _, err := os.Stat(filepath.Join(rootDir, "gnomod.toml")); err == nil {
		return fmt.Errorf("gnomod.toml already exists")
	}

	// --bare: only create gnomod.toml.
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

	// Module path given: scaffold in CWD.
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

	// Render the template first — it has no side effects, so a failure here
	// must not leave behind an empty run/ directory.
	data := templateData{PkgName: "main", ScriptName: scriptName, ScriptPath: relPath}
	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return fmt.Errorf("render run template: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(rootDir, relDir), 0o755); err != nil {
		return err
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
	namespace, err := commands.PromptString(io, "Namespace or address", "", validateNamespace)
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

// isValidName reports whether s is a lowercase ASCII Gno identifier.
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

// isBech32Address reports whether s is a valid bech32-encoded address for the
// active chain prefix (e.g. "g1..." on gno.land). Used to allow users to
// scaffold modules under an address-based namespace.
func isBech32Address(s string) bool {
	_, err := crypto.AddressFromBech32(s)
	return err == nil
}

// validateNamespace is a PromptString-compatible validator for the namespace
// slot of a module path. It accepts either a plain name (see validateName) or
// a full bech32 address, mirroring what the chain's addpkg accepts.
func validateNamespace(s string) error {
	if s == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if isBech32Address(s) {
		return nil
	}
	if !isValidName(s) {
		return fmt.Errorf("invalid namespace %q (must be a name of lowercase letters, digits, and underscores, or a bech32 address)", s)
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
