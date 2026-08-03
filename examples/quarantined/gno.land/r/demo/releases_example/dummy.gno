package releases_example

func init() {
	// dummy example data
	changelog.NewRelease(
		"v1",
		"r/demo/examples_example_v1",
		"initial release",
	)
	changelog.NewRelease(
		"v2",
		"r/demo/examples_example_v2",
		"various improvements",
	)
}
