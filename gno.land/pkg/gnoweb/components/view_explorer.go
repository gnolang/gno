package components

const ExplorerViewType ViewType = "explorer-view"

type ExplorerData struct {
	PkgPath        string
	Paths          []string
	PackageCount   int
	CardsListTitle string // Type of the card list (realm or pure)
	CardsList      *UserCardList
	PackageType    UserContributionType
}

// enrichExplorerCardList enriches the explorer card list with the data from paths
func enrichExplorerCardList(data *ExplorerData) {
	typeName := data.PackageType.String()

	if data.PackageCount > 1 {
		typeName += "s"
	}

	items := make([]UserContribution, len(data.Paths))

	// Convert paths to UserContribution items
	for i, path := range data.Paths {
		items[i] = UserContribution{
			Title: path,
			URL:   path,
			Type:  data.PackageType,
		}
	}

	data.CardsListTitle = data.PkgPath

	data.CardsList = &UserCardList{
		Title:             typeName,
		Items:             items,
		Categories:        []UserCardListCategory{},
		TotalCount:        data.PackageCount,
		SearchPlaceholder: "Search " + typeName,
	}
}

func ExplorerView(pkgPath string, paths []string, packageCounter int, packageType UserContributionType) *View {
	viewData := ExplorerData{
		PkgPath:      pkgPath,
		Paths:        paths,
		PackageCount: packageCounter,
		PackageType:  packageType,
	}

	enrichExplorerCardList(&viewData)

	return NewTemplateView(ExplorerViewType, "renderExplorer", viewData)
}
