package components

const DirectoryViewType ViewType = "dir-view"

type DirData struct {
	PkgPath          string
	Files            []string
	FileCounter      int
	ComponentContent Component
}

type directoryViewParams struct {
	DirData
	Article ArticleData
}

func DirectoryView(data DirData) *View {
	viewData := directoryViewParams{
		DirData: data,
		Article: ArticleData{
			ComponentContent: data.ComponentContent,
			Classes:          "md-view",
		},
	}
	return NewTemplateView(DirectoryViewType, "renderDir", viewData)
}
