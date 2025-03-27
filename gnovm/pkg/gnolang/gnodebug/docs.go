package gnodebug

type FlagDoc struct {
	Name        string
	Description string
}

var FlagDocs = [...]FlagDoc{
	{
		Name:        "log_machine",
		Description: "Enables logging machine-related logs, like pops and pushes to the Op, Value, Block and Frame stacks.",
	},
	{
		Name:        "log_preprocess",
		Description: "Enables logging preprocessing-related logs, including preprocessed files.",
	},
	{
		Name:        "log_types",
		Description: "Enable logging generated type ID's.",
	},
	{
		Name:        "pprof",
		Description: "Enables a pprof profiling server on http://localhost:8080.",
	},
}
