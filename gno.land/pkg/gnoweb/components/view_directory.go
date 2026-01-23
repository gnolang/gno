package components

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	FileCounter int
	FilesLinks  FilesLinks
	Mode        ViewMode
	Readme      Component
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
	Link       string
	Name       string
	SourceLink string // Link to source code view
}

// FilesLinks is a slice of FullFileLink
type FilesLinks []FullFileLink

// buildFilesLinks creates FilesLinks from files
func buildFilesLinks(files []string, linkType DirLinkType, pkgPath string) FilesLinks {
	result := make(FilesLinks, len(files))
	for i, file := range files {
		result[i] = FullFileLink{
			Link:       linkType.LinkPrefix(pkgPath) + file,
			Name:       file,
			SourceLink: file + "$source",
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
