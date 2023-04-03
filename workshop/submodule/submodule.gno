package submodule

func init() {
	println("this never prints because submodule isn't in gno/examples/...")
	// if you improve `gnodev run` to support gno paths to search, could work.
}
