package showcase

func init() {
	registerModule(&module{
		name:        "p/svg",
		description: "Helper to store data in an svg.Tree",
		commands: []command{
			{name: "Example", callback: svgExample},
		},
	})
}

func svgExample(path string) string {
	return "XXX: show svg example"
}
