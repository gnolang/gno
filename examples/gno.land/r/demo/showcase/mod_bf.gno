package showcase

func init() {
	registerModule(&module{
		name:        "p/bf",
		description: "Helper to store data in an bf.Tree",
		commands: []command{
			{name: "Example", callback: bfExample},
		},
	})
}

func bfExample(path string) string {
	return "XXX: show data has bf"
}
