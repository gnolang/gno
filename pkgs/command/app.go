package command

type AppItem struct {
	App      App
	Name     string      // arg name
	Desc     string      // short (single line) description of app
	Defaults interface{} // default options
	// Help string // long form help
}

type AppList []AppItem
