package components

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	Files       []string
	FileCounter int
	FilesLinks  FilesLinks
	LinkType    DirLinkType
	HeaderData  HeaderData
}

type DirLinkType int

const (
	DirLinkTypeSource DirLinkType = iota
	DirLinkTypeFile
)

// not just a prefix, in DirLinkTypeSource case, it nedds to wraP THE LINK SUCH AS {{ $pkgpath }}$source&file={{ . }}
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

func (f FilesLinks) GetFullLinks(files []string, linkType DirLinkType, pkgPath string) FilesLinks {
	result := make(FilesLinks, len(files))
	for i, file := range files {
		result[i] = FullFileLink{Link: linkType.LinkPrefix(pkgPath) + file, Name: file}
	}
	return result
}

func DirectoryView(data DirData) *View {
	viewData := DirData{
		PkgPath:     data.PkgPath,
		Files:       data.Files,
		FilesLinks:  FilesLinks{}.GetFullLinks(data.Files, data.LinkType, data.PkgPath),
		FileCounter: data.FileCounter,
		LinkType:    data.LinkType,
		HeaderData:  data.HeaderData,
	}
	return NewTemplateView(DirectoryViewType, "renderDir", viewData)
}
