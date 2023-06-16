package showcase

func init() {
	modules = append(modules, module{
		name: "p/avl",
		commands: []command{
			{name: "Example", callback: avlExample},
		},
	})
}

func avlExample(path string) string {
	return "XXX: show how data is stored in an avl.Tree"
}
