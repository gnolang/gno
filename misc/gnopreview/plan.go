package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	// examplesRel is the gno workspace root inside the monorepo.
	examplesRel = "examples"
	// pkgRoot is the only subtree of it that holds deployable packages.
	// examples/quarantined/ mirrors the same gno.land/… layout for packages
	// that are deliberately not loaded, and must never be previewed.
	pkgRoot = examplesRel + "/gno.land"
)

// gnowebPaths are the source trees whose changes alter what gnoweb renders for
// *every* page. A diff touching one of them makes the preview worth building
// even when no realm changed.
var gnowebPaths = []string{
	"gno.land/pkg/gnoweb/",
	"gno.land/cmd/gnoweb/",
	"contribs/gnodev/",
}

// gnowebSeedRealms is the fixed sample rendered when only gnoweb code changed:
// enough surface to eyeball the chrome (markdown, syntax highlighting, the help
// forms, the directory listings) without booting the whole examples tree.
var gnowebSeedRealms = []string{
	"gno.land/r/gnoland/home",           // markdown-heavy landing page
	"gno.land/r/docs/security_patterns", // long-form docs markdown
	"gno.land/r/gnoland/boards2/v0",     // many exported funcs -> the $help forms
	"gno.land/r/gnoland/blog",           // render arguments (:page paths)
}

// indirectExclude are realms never pulled in as *dependents* of a changed
// package: they are VM fixtures with no reader-facing render, and a change to a
// widely imported package would otherwise crowd the real realms out of the cap.
// A PR that edits one of them directly is still previewed.
var indirectExclude = []string{
	"gno.land/r/tests/",
}

// Pkg is one gno package found under examples/.
type Pkg struct {
	Path    string   // gno.land/r/gnoland/home
	Dir     string   // examples/gno.land/r/gnoland/home (repo-relative, slash-separated)
	Imports []string // gno.land/* imports of its non-test files
	Realm   bool     // lives under gno.land/r/
	Draft   bool     // gnomod.toml: draft = true
	Ignore  bool     // gnomod.toml: ignore = true
}

// Plan is what the preview job needs to do, derived from a changed-file list.
type Plan struct {
	// Gnoweb is set when the diff touches gnoweb/gnodev itself.
	Gnoweb bool `json:"gnoweb"`
	// ChangedRealms are realms whose own sources changed.
	ChangedRealms []string `json:"changed_realms"`
	// ChangedPkgs are non-realm packages whose sources changed.
	ChangedPkgs []string `json:"changed_pkgs"`
	// Realms is what actually gets rendered: changed realms plus every realm
	// that (transitively) imports a changed package, capped at MaxRealms.
	Realms []string `json:"realms"`
	// Dropped counts realms left out by the cap — never silently.
	Dropped int `json:"dropped"`
	// Dirs are the repo-relative package dirs handed to gnodev.
	Dirs []string `json:"dirs"`
	// ChangedFiles maps a changed realm to the base names of its files that the
	// pull request touched. Per-file $source pages are rendered only for these:
	// they are 60% of a wide preview's bytes, and a reviewer wants the files
	// that changed, not all of them.
	ChangedFiles map[string][]string `json:"changed_files,omitempty"`
	// Shots are screenshots taken of the finished snapshot, embedded in the
	// PR comment. Only populated for gnoweb changes.
	Shots []Shot `json:"shots,omitempty"`
	// Pairs are before/after screenshots of realms this PR changed, rendered
	// from the merge base and from the head with the same gnoweb.
	Pairs []ShotPair `json:"pairs,omitempty"`
}

// Empty reports whether there is nothing worth previewing.
func (p *Plan) Empty() bool { return !p.Gnoweb && len(p.Realms) == 0 }

// Mode is the one-word summary used by the PR comment.
func (p *Plan) Mode() string {
	switch {
	case p.Gnoweb && len(p.ChangedRealms) == 0 && len(p.ChangedPkgs) == 0:
		return "gnoweb"
	case p.Gnoweb:
		return "both"
	case len(p.Realms) > 0:
		return "realms"
	}
	return "none"
}

// LoadPkgs walks examples/ and returns every gno package, keyed by package path.
func LoadPkgs(root string) (map[string]*Pkg, error) {
	pkgs := map[string]*Pkg{}
	base := filepath.Join(root, filepath.FromSlash(pkgRoot))
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		ents, err := os.ReadDir(p)
		if err != nil {
			return err
		}
		var files []string
		hasMod := false
		for _, e := range ents {
			switch {
			case e.Name() == "gnomod.toml":
				hasMod = true
			case e.IsDir() || !strings.HasSuffix(e.Name(), ".gno"):
			case strings.HasSuffix(e.Name(), "_test.gno"), strings.HasSuffix(e.Name(), "_filetest.gno"):
			default:
				files = append(files, filepath.Join(p, e.Name()))
			}
		}
		if !hasMod || len(files) == 0 {
			return nil
		}
		rel := filepath.ToSlash(mustRel(root, p))
		pkg := &Pkg{
			Path:  strings.TrimPrefix(rel, examplesRel+"/"),
			Dir:   rel,
			Realm: strings.Contains(rel, "/r/"),
		}
		pkg.Draft, pkg.Ignore = modFlags(filepath.Join(p, "gnomod.toml"))
		pkg.Imports = gnoImports(files)
		pkgs[pkg.Path] = pkg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

// modFlags reads the two booleans we care about out of a gnomod.toml without
// pulling in a TOML parser: they are always plain top-level `key = true` lines.
func modFlags(p string) (draft, ignore bool) {
	b, err := os.ReadFile(p)
	if err != nil {
		return false, false
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		switch strings.ReplaceAll(strings.TrimSpace(line), " ", "") {
		case "draft=true":
			draft = true
		case "ignore=true":
			ignore = true
		}
	}
	return draft, ignore
}

// gnoImports returns the gno.land/* imports of the given files. .gno is Go
// syntax, so go/parser reads the import block as-is.
func gnoImports(files []string) []string {
	set := map[string]bool{}
	fset := token.NewFileSet()
	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			continue // a PR that does not compile still gets a best-effort plan
		}
		for _, imp := range af.Imports {
			v := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(v, "gno.land/") {
				set[v] = true
			}
		}
	}
	return sortedKeys(set)
}

// BuildPlan turns a changed-file list into the work the preview job must do.
// maxRealms caps how many realms are rendered; 0 means no cap.
func BuildPlan(root string, changed []string, maxRealms int) (*Plan, error) {
	pkgs, err := LoadPkgs(root)
	if err != nil {
		return nil, err
	}
	// dir -> package, so a changed file maps back to its package.
	byDir := map[string]*Pkg{}
	for _, p := range pkgs {
		byDir[p.Dir] = p
	}

	plan := &Plan{ChangedFiles: map[string][]string{}}
	changedPkgs := map[string]bool{}
	changedFiles := map[string]map[string]bool{}
	for _, f := range changed {
		f = filepath.ToSlash(strings.TrimSpace(f))
		if f == "" {
			continue
		}
		for _, gw := range gnowebPaths {
			if strings.HasPrefix(f, gw) {
				plan.Gnoweb = true
			}
		}
		if !strings.HasPrefix(f, pkgRoot+"/") || !renderRelevant(f) {
			continue
		}
		if p := byDir[path.Dir(f)]; p != nil && !p.Ignore {
			changedPkgs[p.Path] = true
			if changedFiles[p.Path] == nil {
				changedFiles[p.Path] = map[string]bool{}
			}
			changedFiles[p.Path][path.Base(f)] = true
		}
	}

	for _, p := range sortedKeys(changedPkgs) {
		if pkgs[p].Realm {
			plan.ChangedRealms = append(plan.ChangedRealms, p)
		} else {
			plan.ChangedPkgs = append(plan.ChangedPkgs, p)
		}
	}

	// Every realm that transitively imports something that changed has to be
	// re-rendered too: that is the "package being a dependency of realms" case.
	affected := map[string]bool{}
	for _, p := range dependents(pkgs, changedPkgs) {
		if !pkgs[p].Realm || pkgs[p].Ignore {
			continue
		}
		if !changedPkgs[p] && hasPrefixAny(p, indirectExclude) {
			continue
		}
		affected[p] = true
	}

	realms := sortedKeys(affected)
	// Directly changed realms first, so the cap never drops the realm the PR is
	// actually about.
	sort.SliceStable(realms, func(i, j int) bool {
		return changedPkgs[realms[i]] && !changedPkgs[realms[j]]
	})
	if maxRealms > 0 && len(realms) > maxRealms {
		plan.Dropped = len(realms) - maxRealms
		realms = realms[:maxRealms]
	}
	sort.Strings(realms)
	plan.Realms = realms

	if plan.Gnoweb {
		for _, r := range gnowebSeedRealms {
			if pkgs[r] != nil && !contains(plan.Realms, r) {
				plan.Realms = append(plan.Realms, r)
			}
		}
		sort.Strings(plan.Realms)
	}
	for _, r := range plan.Realms {
		plan.Dirs = append(plan.Dirs, pkgs[r].Dir)
		if f := changedFiles[r]; len(f) > 0 {
			plan.ChangedFiles[r] = sortedKeys(f)
		}
	}
	return plan, nil
}

// renderRelevant filters out files that cannot change what a realm renders.
func renderRelevant(f string) bool {
	base := path.Base(f)
	switch {
	case strings.HasSuffix(base, "_test.gno"), strings.HasSuffix(base, "_filetest.gno"):
		return false
	case strings.Contains(f, "/filetests/"):
		return false
	case strings.HasSuffix(base, ".gno"), base == "gnomod.toml":
		return true
	}
	return false
}

// dependents returns seeds plus every package that transitively imports one.
func dependents(pkgs map[string]*Pkg, seeds map[string]bool) []string {
	// reverse edges: imported -> importers
	rev := map[string][]string{}
	for _, p := range pkgs {
		for _, imp := range p.Imports {
			rev[imp] = append(rev[imp], p.Path)
		}
	}
	out := map[string]bool{}
	queue := sortedKeys(seeds)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if out[cur] {
			continue
		}
		out[cur] = true
		queue = append(queue, rev[cur]...)
	}
	return sortedKeys(out)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func hasPrefixAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func contains(s []string, v string) bool {
	return slices.Contains(s, v)
}

func mustRel(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		panic(fmt.Sprintf("rel %q %q: %v", base, p, err))
	}
	return r
}
