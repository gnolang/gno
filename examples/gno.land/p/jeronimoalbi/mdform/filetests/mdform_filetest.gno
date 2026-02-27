package main

import "gno.land/p/jeronimoalbi/mdform"

func main() {
	form := mdform.
		New(
			"exec", "FunctionName",
		).
		Input(
			"name",
			"placeholder", "Name",
			"value", "John Doe",
		).
		Select(
			"country",
			"United States",
			"description", "Select your country",
		).
		Select(
			"country",
			"Spain",
		).
		Select(
			"country",
			"Germany",
		).
		Checkbox(
			"interests",
			"music",
			"description", "What do you like to do?",
		).
		Checkbox(
			"interests",
			"tech",
			"checked", "true",
		)
	println(form.String())
}

// Output:
// <gno-form exec="FunctionName">
// <gno-input name="name" placeholder="Name" value="John Doe" />
// <gno-select name="country" value="United States" description="Select your country" />
// <gno-select name="country" value="Spain" />
// <gno-select name="country" value="Germany" />
// <gno-input type="checkbox" name="interests" value="music" description="What do you like to do?" />
// <gno-input type="checkbox" name="interests" value="tech" checked="true" />
// </gno-form>
