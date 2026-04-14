# Markdown Form Package

The package provides a very simplistic [Gno-Flavored Markdown form](/r/docs/markdown#forms) generator.

Forms can be created by sequentially calling form methods to create each one of the form fields.

Example usage:

```go
import "gno.land/p/jeronimoalbi/mdform"

func Render(string) string {
    form := mdform.New()

    // Add a text input field
    form.Input(
        "name",
        "placeholder", "Name",
        "value", "John Doe",
    )

    // Add a select field with three possible values
    form.Select(
        "country",
        "United States",
        "description", "Select your country",
    )
    form.Select(
        "country",
        "Spain",
    )
    form.Select(
        "country",
        "Germany",
    )

    // Add a checkbox group with two possible values
    form.Checkbox(
        "interests",
        "music",
        "description", "What do you like to do?",
    )
    form.Checkbox(
        "interests",
        "tech",
        "checked", "true",
    )

    return form.String()
}
```

Form output:

```html
<gno-form exec="FunctionName">
    <gno-input name="name" placeholder="Name" value="John Doe" />
    <gno-select name="country" value="United States" description="Select your country" />
    <gno-select name="country" value="Spain" />
    <gno-select name="country" value="Germany" />
    <gno-input type="checkbox" name="interests" value="music" description="What do you like to do?" />
    <gno-input type="checkbox" name="interests" value="tech" checked="true" />
</gno-form>
```
