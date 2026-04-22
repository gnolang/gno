package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
		newModInitDeprecatedCmd(io),
		newModTidy(io),
		newModWhy(io),
	)

	return cmd
}

// newModInitDeprecatedCmd registers `gno mod init` as a thin deprecation alias
// that forwards to the top-level `gno init` command. Prints a notice on stderr
// so existing scripts, Makefiles, and CI pipelines keep working while users
// migrate to the new entry point.
func newModInitDeprecatedCmd(io commands.IO) *commands.Command {
	cfg := &modInitCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags] [<module-path>]",
			ShortHelp:  "deprecated: use 'gno init' instead",
			LongHelp: `Deprecated alias for 'gno init'. Forwards all arguments and flags
to the top-level 'gno init' command. Please migrate your scripts.`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			io.ErrPrintln("warning: 'gno mod init' is deprecated; use 'gno init' instead")
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

type moduleKind int

const (
	kindPackage moduleKind = iota
	kindRealm
	kindRun
)

type modInitCfg struct {
	bare     bool
	template string
}

func (c *modInitCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.bare, "bare", false, "only create gnomod.toml, skip template files")
	fs.StringVar(&c.template, "template", "", "template name to use (e.g. basic); skips interactive selection")
}

func newInitCmd(io commands.IO) *commands.Command {
	cfg := &modInitCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "init",
			ShortUsage: "init [flags] [<module-path>]",
			ShortHelp:  "initialize a new Gno module",
			LongHelp: `Initialize a new Gno module in the current directory.

When run in an interactive terminal, a wizard guides you through:
  1. Module kind selection (realm, package, or run script)
  2. Namespace and module name
  3. Template selection (when multiple templates are available)

If a module path is given as an argument, the kind is auto-detected from the
path (/r/ for realms, /p/ for packages) and the first available template is
used. Short-form paths like "r/demo/foo" are expanded to "gno.land/r/demo/foo".

If the argument ends in .gno, a run script is created at that path without
a gnomod.toml (e.g. "gno init run/hello.gno" creates run/hello.gno).

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
  gno init gno.land/r/myname/myrealm   # realm with basic template
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

	// .gno argument: create a run script at the given path, no gnomod.toml
	if len(args) == 1 && strings.HasSuffix(args[0], ".gno") {
		if err := validateGnoPath(args[0]); err != nil {
			return err
		}
		templates := templatesForKind(kindRun)
		tmpl, err := resolveTemplate(templates, cfg.template)
		if err != nil {
			return err
		}
		return writeRunScript(rootDir, args[0], *tmpl, io)
	}

	// Early check: if gnomod.toml already exists, bail out immediately.
	modFilePath := filepath.Join(rootDir, "gnomod.toml")
	if _, err := os.Stat(modFilePath); err == nil {
		return fmt.Errorf("gnomod.toml already exists")
	}

	// --bare: only create gnomod.toml
	if cfg.bare {
		if len(args) == 0 {
			return fmt.Errorf("module path is required with --bare")
		}
		return writeGnomod(rootDir, normalizeModulePath(args[0]))
	}

	// Non-interactive (no TTY) with argument: create gnomod.toml + template files
	if !commands.IsInteractive() {
		if len(args) == 0 {
			return fmt.Errorf("module path is required (non-interactive mode)")
		}
		modPath := normalizeModulePath(args[0])
		return scaffoldModule(rootDir, modPath, templatesForKind(kindFromPath(modPath)), cfg.template, io)
	}

	// Interactive with argument: create gnomod.toml + resolve template
	if len(args) == 1 {
		modPath := normalizeModulePath(args[0])
		return scaffoldModule(rootDir, modPath, templatesForKind(kindFromPath(modPath)), cfg.template, io)
	}

	// Full interactive wizard
	kind, err := promptModuleKind(io)
	if err != nil {
		return err
	}
	if kind == kindRun {
		templates := templatesForKind(kindRun)
		tmpl, tmplErr := resolveTemplate(templates, cfg.template)
		if tmplErr != nil {
			return tmplErr
		}
		return execInitRun(rootDir, *tmpl, io)
	}

	modPath, err := promptModulePath(kind, rootDir, io)
	if err != nil {
		return err
	}

	templates := templatesForKind(kind)
	// In wizard mode with no explicit --template, prompt the user to pick one.
	if cfg.template == "" {
		tmpl, err := selectTemplate(templates, io)
		if err != nil {
			return err
		}
		return scaffoldModuleWith(rootDir, modPath, tmpl, io)
	}
	return scaffoldModule(rootDir, modPath, templates, cfg.template, io)
}

// scaffoldModule resolves a template by name (or the default when empty),
// pre-checks for file conflicts, writes gnomod.toml, then renders template
// files. Any resolution or conflict error is surfaced before gnomod.toml is
// written, so a failed init never leaves an orphan gnomod.toml on disk.
func scaffoldModule(rootDir, modPath string, templates []initTemplate, templateName string, io commands.IO) error {
	tmpl, err := resolveTemplate(templates, templateName)
	if err != nil {
		return err
	}
	return scaffoldModuleWith(rootDir, modPath, tmpl, io)
}

// scaffoldModuleWith is the shared tail of every non-bare, non-run code path:
// render (pre-check) → write gnomod.toml → write module files.
func scaffoldModuleWith(rootDir, modPath string, tmpl *initTemplate, io commands.IO) error {
	if _, err := renderModuleFiles(rootDir, modPath, tmpl); err != nil {
		return err
	}
	if err := writeGnomod(rootDir, modPath); err != nil {
		return err
	}
	return writeModule(rootDir, modPath, tmpl, io)
}

func writeGnomod(rootDir, modPath string) error {
	if err := module.CheckImportPath(modPath); err != nil {
		return fmt.Errorf("create gnomod.toml: %w", err)
	}
	if !gno.IsUserlib(modPath) {
		return fmt.Errorf("create gnomod.toml: %q is not a valid package path URL", modPath)
	}

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
	pkgName := filepath.Base(modPath)
	data := templateData{PkgName: pkgName}

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

func writeModule(rootDir, modPath string, tmpl *initTemplate, io commands.IO) error {
	files, err := renderModuleFiles(rootDir, modPath, tmpl)
	if err != nil {
		return err
	}

	kindLabel := "package"
	if gno.IsRealmPath(modPath) {
		kindLabel = "realm"
	}

	created := make([]string, 0, len(files))
	for _, name := range sortedKeys(files) {
		if err := os.WriteFile(filepath.Join(rootDir, name), files[name], 0o644); err != nil {
			return err
		}
		created = append(created, name)
	}

	fmt.Fprintf(io.Err(), "Initialized %s %s (%s template)\n", kindLabel, modPath, tmpl.Name)
	fmt.Fprintf(io.Err(), "  gnomod.toml\n")
	for _, f := range created {
		fmt.Fprintf(io.Err(), "  %s\n", f)
	}

	return nil
}

// writeRunScript creates a run script at the given relative path (e.g. "run/hello.gno").
// No gnomod.toml is created — run scripts don't need one.
func writeRunScript(rootDir, relPath string, tmpl initTemplate, io commands.IO) error {
	outDir := filepath.Join(rootDir, filepath.Dir(relPath))
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	scriptName := strings.TrimSuffix(filepath.Base(relPath), ".gno")
	data := templateData{PkgName: "main", ScriptName: scriptName, ScriptPath: relPath}

	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return fmt.Errorf("render run template: %w", err)
	}

	for name := range files {
		p := filepath.Join(filepath.Dir(relPath), name)
		if _, err := os.Stat(filepath.Join(rootDir, p)); err == nil {
			return fmt.Errorf("file already exists: %s", p)
		}
	}

	created := make([]string, 0, len(files))
	for _, name := range sortedKeys(files) {
		p := filepath.Join(filepath.Dir(relPath), name)
		if err := os.WriteFile(filepath.Join(rootDir, p), files[name], 0o644); err != nil {
			return err
		}
		created = append(created, p)
	}

	fmt.Fprintf(io.Err(), "Initialized run script (%s template)\n", tmpl.Name)
	for _, f := range created {
		fmt.Fprintf(io.Err(), "  %s\n", f)
	}
	return nil
}

func execInitRun(rootDir string, tmpl initTemplate, io commands.IO) error {
	defaultName := sanitizeModuleName(filepath.Base(rootDir))
	if defaultName == "" {
		defaultName = "main"
	}
	scriptName, err := commands.PromptString(io, "Script name", defaultName, validateName)
	if err != nil {
		return err
	}
	relPath := filepath.Join("run", scriptName+".gno")
	return writeRunScript(rootDir, relPath, tmpl, io)
}

// --- Wizard helpers ---

func promptModuleKind(io commands.IO) (moduleKind, error) {
	choices := map[string]commands.Choice{
		"r": {Aliases: []string{"realm"}, Description: "realm"},
		"p": {Aliases: []string{"package"}, Description: "package"},
		"m": {Aliases: []string{"main", "run"}, Description: "run script"},
	}

	key, err := commands.PromptChoice(io, "Module kind — [r]ealm, [p]ackage, or [m]ain: ", choices, "p")
	if err != nil {
		return kindPackage, err
	}

	switch key {
	case "r":
		return kindRealm, nil
	case "m":
		return kindRun, nil
	default:
		return kindPackage, nil
	}
}

func promptModulePath(kind moduleKind, rootDir string, io commands.IO) (string, error) {
	namespace, err := commands.PromptString(io, "Namespace", "", validateName)
	if err != nil {
		return "", err
	}

	defaultName := sanitizeModuleName(filepath.Base(rootDir))
	name, err := commands.PromptString(io, "Module name", defaultName, validateName)
	if err != nil {
		return "", err
	}

	basePath := fmt.Sprintf("gno.land/%s/%s", namespace, name)
	return insertPathLetter(basePath, kind)
}

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

func kindFromPath(modPath string) moduleKind {
	if gno.IsRealmPath(modPath) {
		return kindRealm
	}
	return kindPackage
}

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

func insertPathLetter(path string, kind moduleKind) (string, error) {
	idx := strings.Index(path, "/")
	if idx == -1 || idx == len(path)-1 {
		return "", fmt.Errorf("invalid module path: %q", path)
	}
	domain := path[:idx]
	rest := path[idx+1:]

	// Idempotency guard: if the path already has an /r/ or /p/ segment
	// immediately after the domain, return it unchanged.
	if strings.HasPrefix(rest, "r/") || strings.HasPrefix(rest, "p/") {
		return path, nil
	}

	var letter string
	switch kind {
	case kindRealm:
		letter = "r"
	default:
		letter = "p"
	}
	return fmt.Sprintf("%s/%s/%s", domain, letter, rest), nil
}

var reModName = regexp.MustCompile(`[^a-z0-9_]`)

var reValidName = regexp.MustCompile(`^_?[a-z][a-z0-9_]*$`)

func isValidName(s string) bool {
	return reValidName.MatchString(s)
}

func validateName(s string) error {
	if s == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if !isValidName(s) {
		return fmt.Errorf("invalid value %q (must be lowercase letters, digits, and underscores)", s)
	}
	return nil
}

// validateGnoPath ensures the .gno argument is a safe relative path within CWD.
// It rejects absolute paths, path traversal (..), and empty script names.
func validateGnoPath(p string) error {
	if filepath.IsAbs(p) {
		return fmt.Errorf("path must be relative, got %q", p)
	}
	cleaned := filepath.Clean(p)
	// After Clean, a real traversal is either exactly ".." or starts with
	// ".." followed by the platform separator. A plain HasPrefix(cleaned, "..")
	// would false-positive on legitimate names like "..bar.gno".
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path must not traverse outside current directory, got %q", p)
	}
	base := filepath.Base(p)
	scriptName := strings.TrimSuffix(base, ".gno")
	if scriptName == "" {
		return fmt.Errorf("script name cannot be empty")
	}
	if !isValidName(scriptName) {
		return fmt.Errorf("invalid script name %q (must be lowercase letters, digits, and underscores)", scriptName)
	}
	return nil
}

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

func sanitizeModuleName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = reModName.ReplaceAllString(name, "")
	return name
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
