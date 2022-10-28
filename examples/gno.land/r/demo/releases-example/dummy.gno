package releases_example

func init() {
	// dummy example data
	changelog.NewRelease(
		"v1",
		"r/demo/examples-example-v1",
		"initial release",
	)
	changelog.NewRelease(
		"v2",
		"r/demo/examples-example-v2",
		"various improvements",
	)
}
