package components

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	Files       []string
	FileCounter int
	FilesLinks  FilesLinks
	Mode        ViewMode
}

type DirLinkType int

const (
	DirLinkTypeSource DirLinkType = iota
	DirLinkTypeFile
)

// Get the prefixed link depending on link type - Package Source Code or Package File
func (d DirLinkType) LinkPrefix(pkgPath string) string {
	switch d {
	case DirLinkTypeSource:
		return pkgPath + "$source&file="
	case DirLinkTypeFile:
		return "https://"
	}
	return ""
}

// Files has to be an array with Link (prefixed) and Name (filename)
type FullFileLink struct {
	Link string
	Name string
}

// FilesLinks has to be an array of FileLink
type FilesLinks []FullFileLink

func GetFullLinks(files []string, linkType DirLinkType, pkgPath string) FilesLinks {
	result := make(FilesLinks, len(files))
	for i, file := range files {
		result[i] = FullFileLink{Link: linkType.LinkPrefix(pkgPath) + file, Name: file}
	}
	return result
}

func DirectoryView(pkgPath string, files []string, fileCounter int, linkType DirLinkType, mode ViewMode) *View {
	viewData := DirData{
		PkgPath:     pkgPath,
		Files:       files,
		FilesLinks:  GetFullLinks(files, linkType, pkgPath),
		FileCounter: fileCounter,
		Mode:        mode,
	}
	return NewTemplateView(DirectoryViewType, "renderDir", viewData)
}
