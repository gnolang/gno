package components

import (
	"strings"
	"time"
)

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	FileCounter int
	FilesLinks  FilesLinks
	Mode        ViewMode
	Readme      Component
	// Truncated is set on a listing that stopped at its cap.
	Truncated bool
	// Rail is an explorer listing's side rail (ui/listing_rail).
	Rail Component
	// Map is the listing drawn as a map; nil renders it as a list.
	Map Component
	// Cards are the listing's packages, one card each, when drawn as a list.
	Cards []PackageCard
}

// PackageCard is one package of a listing: what it says about itself, what a
// reader can do with it, and, with an indexer, how much it is used.
type PackageCard struct {
	// Path is the package's gnoweb path, such as "/r/gnoland/blog".
	Path string
	// Realm is set for a realm, which can be opened and called.
	Realm bool
	// Description is the opening of the package's doc comment, plain text.
	Description string
	// Actions counts the exported crossing functions a reader can call.
	Actions int

	// Known is set once the chain answered for the package: the fields
	// below are blank, not zero, without it.
	Known bool
	// Block is the height the package was deployed at; 0 is genesis.
	Block int
	// Date is when Block was produced, as an indexer reports it.
	Date time.Time

	// WeeklyUsers counts the distinct addresses that called the realm over
	// the indexer's last seven days; AtLeast marks it a lower bound.
	WeeklyUsers int
	AtLeast     bool
	// LastUsed is about when the latest of those calls ran.
	LastUsed time.Time
}

// MapParts are what a map adds to a listing: the figure in the body, and its
// key in the rail.
type MapParts struct {
	Figure, Key Component
}

// listingRail is the payload of ui/listing_rail.
type listingRail struct {
	Packages, Folders int
	Key               Component
}

// countFolders counts the first path segments below root among paths: the
// boxes a map of them draws, and what a reader scans a list for.
func countFolders(root string, paths []string) int {
	seen := make(map[string]struct{})
	for _, p := range paths {
		key, _, _ := strings.Cut(strings.TrimPrefix(p, root), "/")
		seen[key] = struct{}{}
	}
	return len(seen)
}

// FullFileLink represents a file entry in a package's directory.
type FullFileLink struct {
	Link string
	Name string
}

// FilesLinks is a slice of FullFileLink
type FilesLinks []FullFileLink

// buildFilesLinks links each file of pkgPath to its source view.
func buildFilesLinks(files []string, pkgPath string) FilesLinks {
	result := make(FilesLinks, len(files))
	for i, file := range files {
		result[i] = FullFileLink{
			Link: pkgPath + "$source&file=" + file,
			Name: file,
		}
	}
	return result
}

// DirectoryView renders the files of the package at pkgPath.
func DirectoryView(pkgPath string, files []string, fileCounter int, mode ViewMode, readme ...Component) *View {
	viewData := DirData{
		PkgPath:     pkgPath,
		FilesLinks:  buildFilesLinks(files, pkgPath),
		FileCounter: fileCounter,
		Mode:        mode,
	}
	if len(readme) > 0 {
		viewData.Readme = readme[0]
	}
	return NewTemplateView(DirectoryViewType, "renderDir", viewData)
}

// ExplorerView renders the package paths under pkgPath, as cards, or as a map
// when m is set: one page, one header and one rail for both renderings. cards
// hold one entry per path, in the same order. truncated says the node stopped
// at its cap, which the view then states rather than letting the listing pass
// for complete.
func ExplorerView(pkgPath string, paths []string, truncated bool, m *MapParts, cards []PackageCard) *View {
	// A listing is named by its root, with the slash that says so.
	root := strings.TrimSuffix(pkgPath, "/") + "/"
	data := DirData{
		PkgPath:     root,
		FileCounter: len(paths),
		Mode:        ViewModeExplorer,
		Truncated:   truncated,
	}
	rail := listingRail{Packages: len(paths), Folders: countFolders(root, paths)}
	if m != nil {
		data.Map, rail.Key = m.Figure, m.Key
	} else {
		data.Cards = cards
	}
	data.Rail = NewTemplateComponent("ui/listing_rail", rail)
	view := NewTemplateView(DirectoryViewType, "renderDir", data)
	view.SkipTargetInBody = true // on the content header
	return view
}
