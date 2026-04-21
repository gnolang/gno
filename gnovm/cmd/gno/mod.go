package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/mod/module"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload"
	"github.com/gnolang/gno/gnovm/pkg/packages/pkgdownload/rpcpkgfetcher"
	"github.com/gnolang/gno/tm2/pkg/commands"
	terrors "github.com/gnolang/gno/tm2/pkg/errors"
)

// testPackageFetcher allows to override the package fetcher during tests.
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
		// edit
		newModGraphCmd(io),
		newModTidy(io),
		// vendor
		// verify
		newModWhy(io),
	)

	return cmd
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
  gno init --template basic gno.land/r/demo/foo  # explicit template`,
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

	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	if !filepath.IsAbs(rootDir) {
		return fmt.Errorf("create gnomod.toml: dir %q is not absolute", rootDir)
	}

	// Early check: if gnomod.toml already exists, bail out immediately.
	modFilePath := filepath.Join(rootDir, "gnomod.toml")
	if _, err := os.Stat(modFilePath); err == nil {
		return fmt.Errorf("gnomod.toml already exists")
	}

	// Non-interactive: bare flag or no TTY
	if cfg.bare || !commands.IsInteractive() {
		if len(args) == 0 {
			return fmt.Errorf("module path is required (non-interactive mode)")
		}
		return writeGnomod(rootDir, normalizeModulePath(args[0]))
	}

	// Interactive with argument: create gnomod.toml + resolve template
	if len(args) == 1 {
		modPath := normalizeModulePath(args[0])
		if err := writeGnomod(rootDir, modPath); err != nil {
			return err
		}
		kind := kindFromPath(modPath)
		templates := templatesForKind(kind)
		tmpl, err := resolveTemplate(templates, cfg.template)
		if err != nil {
			return err
		}
		if kind == kindRun {
			return execInitRun(rootDir, *tmpl, io)
		}
		return writeModule(rootDir, modPath, tmpl, io)
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
	var tmpl *initTemplate
	if cfg.template != "" {
		tmpl, err = resolveTemplate(templates, cfg.template)
	} else {
		tmpl, err = selectTemplate(templates, io)
	}
	if err != nil {
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

func writeModule(rootDir, modPath string, tmpl *initTemplate, io commands.IO) error {
	pkgName := filepath.Base(modPath)
	data := templateData{PkgName: pkgName}

	kindLabel := "package"
	if gno.IsRealmPath(modPath) {
		kindLabel = "realm"
	}

	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	for name := range files {
		if _, err := os.Stat(filepath.Join(rootDir, name)); err == nil {
			return fmt.Errorf("file already exists: %s", name)
		}
	}

	var created []string
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(rootDir, name), content, 0o644); err != nil {
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

func execInitRun(rootDir string, tmpl initTemplate, io commands.IO) error {
	defaultName := sanitizeModuleName(filepath.Base(rootDir))
	if defaultName == "" {
		defaultName = "main"
	}
	scriptName, err := commands.PromptString(io, "Script name", defaultName, validateName)
	if err != nil {
		return err
	}

	runDir := filepath.Join(rootDir, "run")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}

	data := templateData{PkgName: "main", ScriptName: scriptName}
	files, err := renderTemplateDir(tmpl.FS, tmpl.Dir, data)
	if err != nil {
		return fmt.Errorf("render run template: %w", err)
	}

	for name := range files {
		if _, err := os.Stat(filepath.Join(runDir, name)); err == nil {
			return fmt.Errorf("file already exists: run/%s", name)
		}
	}

	var created []string
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(runDir, name), content, 0o644); err != nil {
			return err
		}
		created = append(created, name)
	}

	fmt.Fprintf(io.Err(), "Initialized run script\n")
	for _, f := range created {
		fmt.Fprintf(io.Err(), "  run/%s\n", f)
	}
	return nil
}

// --- Wizard helpers ---

func promptModuleKind(io commands.IO) (moduleKind, error) {
	choices := []commands.Choice{
		{Key: "r", Aliases: []string{"realm"}, Description: "realm"},
		{Key: "p", Aliases: []string{"package"}, Description: "package", IsDefault: true},
		{Key: "m", Aliases: []string{"main", "run"}, Description: "run script"},
	}
	kinds := []moduleKind{kindRealm, kindPackage, kindRun}

	idx, err := commands.PromptChoice(io, "Module kind — [r]ealm, [P]ackage, or [m]ain: ", choices)
	if err != nil {
		return kindPackage, err
	}
	return kinds[idx], nil
}

func promptModulePath(kind moduleKind, rootDir string, io commands.IO) (string, error) {
	// Step 1: namespace (retry on empty or invalid)
	namespace, err := commands.PromptString(io, "Address or namespace", "", validateName)
	if err != nil {
		return "", err
	}

	// Step 2: module name (retry on invalid)
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

	idx, err := commands.PromptSelect(io, "Template:", items)
	if err != nil {
		return nil, err
	}
	return &templates[idx], nil
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
	// path is like "gno.land/namespace/name"
	// insert /r/ or /p/ after the domain
	idx := strings.Index(path, "/")
	if idx == -1 || idx == len(path)-1 {
		return "", fmt.Errorf("invalid module path: %q", path)
	}
	domain := path[:idx]
	rest := path[idx+1:]

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

// reValidName matches the same pattern as gnolang's Re_name: optional leading _,
// must start with a lowercase letter, then lowercase letters, digits, underscores.
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

// normalizeModulePath expands short-form module paths to their fully-qualified
// gno.land equivalents. For example:
//   - "p/nt/hello"  -> "gno.land/p/nt/hello"
//   - "r/demo/foo"  -> "gno.land/r/demo/foo"
//   - "gno.land/p/nt/hello" -> unchanged
//
// Paths that don't start with "p/" or "r/" and don't contain "." in the first
// segment (i.e. no domain) are returned unchanged so the validator can report
// a clear error.
func normalizeModulePath(modPath string) string {
	if modPath == "" {
		return modPath
	}
	// Already has a domain (contains a dot in the first segment).
	first, _, _ := strings.Cut(modPath, "/")
	if strings.Contains(first, ".") {
		return modPath
	}
	// Short form: "p/..." or "r/..." — prepend gno.land.
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
	// /out std
	// /out remote
	// /out _test processing
	// ...
	fs.StringVar(&c.format, "format", "", "Output format, must be one of 'dot' or empty. Empty is a minimalist format.")
}

func execModGraph(cfg *modGraphCfg, args []string, io commands.IO) error {
	// default to current directory if no args provided
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
		// XXX: xtests and filetests should probably be treated as their own packages since they can/will have cycles
		// when considered as part of the source package
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
		return terrors.New("%d build error(s)", errCount)
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

	// XXX: recursively check parents if no $PWD/gno.mod
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
			// migrate from gno.mod to gnomod.toml
			newPath := filepath.Join(pkgdir, "gnomod.toml")
			gm.WriteFile(newPath) // gnomod.toml
		} else {
			gm.WriteFile(fpath) // gnomod.toml
		}
	}

	// there is no gno.mod nor gnomod.toml
	if !modExists {
		return gnomod.ErrNoModFile
	}

	// remove gno.mod if it exists.
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

	// Format and print `gno mod why` output stanzas
	out := formatModWhyStanzas(gm.Module, args, importToFilesMap)
	io.Printf(out)

	return nil
}

// formatModWhyStanzas returns a formatted output for the go mod why command.
// It takes three parameters:
//   - modulePath (the path of the module)
//   - args (input arguments)
//   - importToFilesMap (a map of import to files).
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
		if i < len(args)-1 { // Add a newline if it's not the last stanza
			out += "\n"
		}
	}
	return
}

// getImportToFilesMap returns a map where each key is an import path and its
// value is a list of files importing that package with the specified import path.
func getImportToFilesMap(pkgPath string) (map[string][]string, error) {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]string) // import -> []file
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
