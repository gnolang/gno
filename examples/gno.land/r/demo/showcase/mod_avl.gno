package showcase

func init() {
	registerModule(&module{
		name:        "p/avl",
		description: "Helper to store data in an avl.Tree",
		commands: []command{
			{name: "Store Example", callback: avlExample},
		},
	})
}

func avlExample(path string) string {
	return "XXX: show how data is stored in an avl.Tree"
}
