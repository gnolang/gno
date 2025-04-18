package components

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath     string
	Files       []string
	FileCounter int
}

func DirectoryView(data DirData) *View {
	return NewTemplateView(DirectoryViewType, "renderDir", data)
}
