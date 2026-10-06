package releases_example

func ExampleRender() {
	println("-----------")
	print(Render(""))

	println("-----------")
	print(Render("v1"))

	println("-----------")
	print(Render("v42"))

	// Output:
	// -----------
	// # example_app
	//
	// ## [example_app v2 (latest)](r/demo/examples_example_v2)
	//
	// various improvements
	//
	// ## [example_app v1](r/demo/examples_example_v1)
	//
	// initial release
	//
	// -----------
	// ## [example_app v1](r/demo/examples_example_v1)
	//
	// initial release
	//
	// -----------
	// no such release
}
