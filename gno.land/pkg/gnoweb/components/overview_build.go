package components

import (
	"go/token"
	"path"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

// computeStats derives numeric counters from the file list and qdoc payload.
// bugsNotInDoc returns the BUG notes not already visible in the rendered package
// doc. go/doc leaves a note's body in Doc when the note lives inside the package
// comment, so without this the same text would render both inline and in the
// dedicated Bugs section. Floating BUG notes (absent from PackageDoc) are kept.
func bugsNotInDoc(bugs []string, packageDoc string) []string {
	if len(bugs) == 0 {
		return nil
	}
	var out []string
	for _, b := range bugs {
		if strings.Contains(packageDoc, strings.TrimSpace(b)) {
			continue
		}
		out = append(out, b)
	}
	return out
}

func computeStats(files []string, jdoc *doc.JSONDocumentation, imports []ImportLink) PackageStats {
	s := PackageStats{
		FileCount:   len(files),
		ImportCount: len(imports),
	}
	for _, f := range files {
		c := ClassifyFile(f)
		switch {
		case c.IsTest:
			s.TestCount++
			s.GnoFileCount++
		case c.IsGno:
			s.GnoFileCount++
		}
	}
	if jdoc == nil {
		return s
	}
	for _, fn := range jdoc.Funcs {
		s.FuncCount++
		if fn.Type == "" && token.IsExported(fn.Name) {
			s.ExportedFunc++
		}
		if fn.Crossing {
			s.CrossingCount++
		}
	}
	// Types/consts/vars count only exported declarations, matching the render
	// path (buildSymbols/buildValues drop unexported), so the sidebar totals
	// agree with the sections actually shown on the page.
	for _, t := range jdoc.Types {
		if token.IsExported(t.Name) {
			s.TypeCount++
		}
	}
	for _, v := range jdoc.Values {
		if !isExportedValueDecl(v) {
			continue
		}
		if v.Const {
			s.ConstCount++
		} else {
			s.VarCount++
		}
	}
	return s
}

// deriveQuality flags presence indicators used by the quality UI section.
func deriveQuality(files []string, jdoc *doc.JSONDocumentation) PackageQuality {
	var q PackageQuality
	for _, f := range files {
		c := ClassifyFile(f)
		if c.IsReadme {
			q.HasReadme = true
		}
		if c.IsLicense {
			q.HasLicense = true
		}
		if c.IsTest {
			q.HasTests = true
		}
	}
	if jdoc != nil && strings.TrimSpace(jdoc.PackageDoc) != "" {
		q.HasPkgDoc = true
	}
	return q
}

// extractSynopsis returns the first line of packageDoc, capped at 120 runes.
func extractSynopsis(packageDoc string) string {
	trimmed := strings.TrimSpace(packageDoc)
	if trimmed == "" {
		return ""
	}
	first := strings.SplitN(trimmed, "\n", 2)[0]
	runes := []rune(first)
	if len(runes) > 120 {
		return string(runes[:117]) + "..."
	}
	return first
}

// buildFileLinks turns file names into URL-backed entries for the Files section.
func buildFileLinks(pkgPath string, files []string) []FileLink {
	out := make([]FileLink, 0, len(files))
	for _, f := range files {
		link := pkgPath + "$source&file=" + f
		entry := FileLink{Name: f, Link: link}
		c := ClassifyFile(f)
		switch {
		case c.IsTest:
			entry.IsTest = true
		case c.IsReadme:
			entry.IsReadme = true
		}
		if c.IsLicense {
			entry.IsLicense = true
		}
		out = append(out, entry)
	}
	return out
}

// buildSubpackages keeps direct children only (one level deep) and drops self.
// Input paths are expected to be domain-relative (e.g. "/r/demo/foo/bar").
func buildSubpackages(self string, paths []string) []SubpackageLink {
	out := make([]SubpackageLink, 0)
	seen := map[string]bool{self: true}
	selfPrefix := self + "/"
	for _, p := range paths {
		p = strings.TrimSuffix(p, "/")
		if seen[p] {
			continue
		}
		seen[p] = true
		rel := strings.TrimPrefix(p, selfPrefix)
		if rel == p || rel == "" || strings.Contains(rel, "/") {
			continue
		}
		out = append(out, SubpackageLink{Name: rel, Path: p})
	}
	return out
}

// buildOverviewTOC builds the hierarchical table-of-contents items used by the sidebar.
func buildOverviewTOC(quality PackageQuality, funcs []FuncEntry, types []TypeEntry, values []ValueGroup, imports []ImportLink, subpacks []SubpackageLink) []*TocItem {
	var toc []*TocItem
	if quality.HasPkgDoc {
		toc = append(toc, &TocItem{Title: "Overview", ID: "overview"})
	}
	if quality.HasReadme {
		toc = append(toc, &TocItem{Title: "README", ID: "readme"})
	}
	hasConst, hasVar := false, false
	for _, v := range values {
		if v.Kind == "const" {
			hasConst = true
		}
		if v.Kind == "var" {
			hasVar = true
		}
	}
	if hasConst {
		toc = append(toc, &TocItem{Title: "Constants", ID: "constants"})
	}
	if hasVar {
		toc = append(toc, &TocItem{Title: "Variables", ID: "variables"})
	}
	if len(funcs) > 0 {
		item := &TocItem{Title: "Functions", ID: "functions"}
		for _, fn := range funcs {
			item.Items = append(item.Items, &TocItem{Title: fn.Name, ID: fn.AnchorID, Icon: "kind-func"})
		}
		toc = append(toc, item)
	}
	if len(types) > 0 {
		item := &TocItem{Title: "Types", ID: "types"}
		for _, t := range types {
			ti := &TocItem{Title: t.Name, ID: t.AnchorID, Icon: typeKindIcon(t.Kind)}
			for _, m := range t.Methods {
				ti.Items = append(ti.Items, &TocItem{Title: m.Name, ID: m.AnchorID, Icon: "kind-func"})
			}
			item.Items = append(item.Items, ti)
		}
		toc = append(toc, item)
	}
	if len(imports) > 0 {
		toc = append(toc, &TocItem{Title: "Imports", ID: "imports"})
	}
	toc = append(toc, &TocItem{Title: "Files", ID: "files"})
	if len(subpacks) > 0 {
		toc = append(toc, &TocItem{Title: "Directories", ID: "subpackages"})
	}
	return toc
}

// typeKindIcon maps a doc type Kind (struct, interface, map, func, ...) to a
// kind-glyph sprite id. Unknown or empty kinds fall back to the generic type box.
func typeKindIcon(kind string) string {
	switch kind {
	case "struct":
		return "kind-struct"
	case "interface":
		return "kind-interface"
	case "slice", "array":
		return "kind-slice"
	case "map":
		return "kind-map"
	case "pointer":
		return "kind-pointer"
	case "func":
		return "kind-func"
	default:
		return "kind-type"
	}
}

// deriveInfo builds the sidebar metadata for a package.
func deriveInfo(gnourl *weburl.GnoURL, gnomodData []byte) PackageInfo {
	info := PackageInfo{
		Namespace:   gnourl.Namespace(),
		PackagePath: gnourl.Path,
		PackageType: packageTypeOf(gnourl),
	}
	if len(gnomodData) > 0 {
		if mod, err := gnomod.ParseBytes("gnomod.toml", gnomodData); err == nil {
			info.GnoVersion = mod.Gno
			info.Draft = mod.Draft
			info.Private = mod.Private
			info.Creator = mod.AddPkg.Creator
			info.Height = mod.AddPkg.Height
		}
	}
	return info
}

func packageTypeOf(gnourl *weburl.GnoURL) string {
	switch {
	case gnourl.IsRealm():
		return "realm"
	case gnourl.IsPure():
		return "pure"
	}
	return ""
}

// maxOverviewSymbols caps how many top-level funcs, types, or value groups the
// overview renders. SymbolsTruncated is set when the cap is hit so the template
// can signal truncation to the reader.
const maxOverviewSymbols = 500

// capSymbols truncates s to at most limit items, reporting whether it was truncated.
func capSymbols[S ~[]E, E any](s S, limit int) (S, bool) {
	if len(s) > limit {
		return s[:limit], true
	}
	return s, false
}

// BuildOverview is pure: given fetched inputs, it returns the rendered OverviewData.
func BuildOverview(in OverviewInput) OverviewData {
	info := deriveInfo(in.URL, in.Sources["gnomod.toml"])
	info.License = deriveLicense(in.Files, func(name string) ([]byte, bool) {
		v, ok := in.Sources[name]
		return v, ok
	})
	quality := deriveQuality(in.Files, in.Doc)
	var importPaths []string
	if in.Doc != nil {
		importPaths = in.Doc.Imports
	}
	imports := buildImports(importPaths, in.Domain)
	funcs, types := buildSymbols(in.Doc, in.DocRenderer, in.URL.Path)
	values := buildValues(in.Doc, in.DocRenderer, in.URL.Path)
	funcs, fTrunc := capSymbols(funcs, maxOverviewSymbols)
	types, tTrunc := capSymbols(types, maxOverviewSymbols)
	values, vTrunc := capSymbols(values, maxOverviewSymbols)
	symbolsTruncated := fTrunc || tTrunc || vTrunc

	// Split value groups by kind so the template renders pre-shaped data
	// (Constants/Variables sections + filter-tab counts) without filtering.
	var consts, vars []ValueGroup
	for _, v := range values {
		if v.Kind == "const" {
			consts = append(consts, v)
		} else {
			vars = append(vars, v)
		}
	}

	stats := computeStats(in.Files, in.Doc, imports)
	files := buildFileLinks(in.URL.Path, in.Files)
	subpacks := buildSubpackages(in.URL.Path, in.Subpaths)
	toc := buildOverviewTOC(quality, funcs, types, values, imports, subpacks)

	pkgDocSynopsis := ""
	var pkgDocComp Component
	if in.Doc != nil {
		pkgDocSynopsis = extractSynopsis(in.Doc.PackageDoc)
		pkgDocComp = renderDocString(in.Doc.PackageDoc, in.DocRenderer)
	}

	var bugs []string
	if in.Doc != nil {
		bugs = bugsNotInDoc(in.Doc.Bugs, in.Doc.PackageDoc)
	}

	return OverviewData{
		PkgPath:          in.URL.Path,
		Title:            path.Base(in.URL.Path),
		Synopsis:         pkgDocSynopsis,
		PackageDoc:       pkgDocComp,
		Readme:           in.Readme,
		Info:             info,
		Stats:            stats,
		Quality:          quality,
		Funcs:            funcs,
		Types:            types,
		Consts:           consts,
		Vars:             vars,
		Imports:          imports,
		Files:            files,
		Subpackages:      subpacks,
		Bugs:             bugs,
		SymbolsTruncated: symbolsTruncated,
		ComponentTOC:     NewTemplateComponent("ui/toc_realm", &RealmTOCData{Items: toc}),
	}
}
