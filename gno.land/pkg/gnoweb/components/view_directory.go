package components

import "strings"

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	FileCounter int
	FilesLinks  FilesLinks
	Mode        ViewMode
	Readme      Component
	// Header heads an explorer listing (ui/listing_header).
	Header ListingHeader
	// Rail is an explorer listing's side rail (ui/listing_rail).
	Rail Component
	// Map is the listing drawn as a map; nil renders it as a list.
	Map Component
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

// ListingHeader heads a directory listing in either rendering, list or map.
type ListingHeader struct {
	// Path is the listing's root, ending with a slash.
	Path  string
	Count int
	// Truncated is set when there were more paths than were listed.
	Truncated bool
}

// NewListingHeader heads the listing of paths below root.
func NewListingHeader(root string, count int, truncated bool) ListingHeader {
	return ListingHeader{Path: strings.TrimSuffix(root, "/") + "/", Count: count, Truncated: truncated}
}

type DirLinkType int

const (
	DirLinkTypeSource DirLinkType = iota
	DirLinkTypeFile
)

// LinkPrefix returns the prefixed link depending on link type
func (d DirLinkType) LinkPrefix(pkgPath string) string {
	switch d {
	case DirLinkTypeSource:
		return pkgPath + "$source&file="
	case DirLinkTypeFile:
		return ""
	}
	return ""
}

// FullFileLink represents a package entry in the directory listing.
type FullFileLink struct {
	Link string
	Name string
}

// FilesLinks is a slice of FullFileLink
type FilesLinks []FullFileLink

// buildFilesLinks creates FilesLinks from files
func buildFilesLinks(files []string, linkType DirLinkType, pkgPath string) FilesLinks {
	result := make(FilesLinks, len(files))
	for i, file := range files {
		result[i] = FullFileLink{
			Link: linkType.LinkPrefix(pkgPath) + file,
			Name: file,
		}
	}
	return result
}

// DirectoryView creates a directory view
func DirectoryView(pkgPath string, files []string, fileCounter int, linkType DirLinkType, mode ViewMode, readme ...Component) *View {
	viewData := DirData{
		PkgPath:     pkgPath,
		FilesLinks:  buildFilesLinks(files, linkType, pkgPath),
		FileCounter: fileCounter,
		Mode:        mode,
	}
	if len(readme) > 0 {
		viewData.Readme = readme[0]
	}
	return NewTemplateView(DirectoryViewType, "renderDir", viewData)
}

// ExplorerView renders the package paths under pkgPath, as a list, or as a
// map when m is set: one page, one header and one rail for both renderings.
// truncated says the node stopped at its cap, which the view then states
// rather than letting the listing pass for complete.
func ExplorerView(pkgPath string, paths []string, truncated bool, m *MapParts) *View {
	data := DirData{
		PkgPath:     pkgPath,
		FileCounter: len(paths),
		Mode:        ViewModeExplorer,
		Header:      NewListingHeader(pkgPath, len(paths), truncated),
	}
	rail := listingRail{Packages: len(paths), Folders: countFolders(data.Header.Path, paths)}
	if m != nil {
		data.Map, rail.Key = m.Figure, m.Key
	} else {
		data.FilesLinks = buildFilesLinks(paths, DirLinkTypeFile, pkgPath)
	}
	data.Rail = NewTemplateComponent("ui/listing_rail", rail)
	view := NewTemplateView(DirectoryViewType, "renderDir", data)
	view.SkipTargetInBody = true // on the content header
	return view
}
