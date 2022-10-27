package releases_example

func init() {
	// dummy example data
	changelog.NewRelease(
		"v1",
		"r/examples-example-v1",
		"initial release",
	)
	changelog.NewRelease(
		"v2",
		"r/examples-example-v2",
		"various improvements",
	)
}
