// Package mictofun_example demonstrates the mictofun markdown framework.
//
// This realm shows how to use all features of the mictofun library
// including fluent API, tables, alerts, and layout components.
package mictofun_example

import (
	"gno.land/p/demo/ui/mictofun"
)

func Render(path string) string {
	switch path {
	case "":
		return renderHome()
	case "tables":
		return renderTables()
	case "alerts":
		return renderAlerts()
	case "features":
		return renderFeatures()
	default:
		return renderHome()
	}
}

func renderHome() string {
	// Create a feature table
	features := mictofun.NewTable().
		Headers("Feature", "Description", "Status").
		SetAlign(mictofun.AlignLeft, mictofun.AlignLeft, mictofun.AlignCenter).
		Row("Fluent API", "Method chaining for readable code", "✓").
		Row("Tables", "With alignment support", "✓").
		Row("Alerts", "GitHub-style callouts", "✓").
		Row("Layouts", "Columns, collapsible sections", "✓")

	return mictofun.New().
		H1("mictofun - Fluent Markdown Framework").
		P("A powerful, chainable markdown generation library for gno.land.").
		Note("This realm demonstrates all features of the mictofun package.").
		H2("Quick Example").
		CodeBlock(`doc := mictofun.New().
    H1("Hello World").
    P("This is " + mictofun.Bold("amazing") + "!").
    Bullet("Fast", "Simple", "Powerful").
    String()`, "gno").
		H2("Features").
		Add(features).
		H2("Navigation").
		Bullet(
			mictofun.Link("Tables Demo", "/r/demo/mictofun_example:tables"),
			mictofun.Link("Alerts Demo", "/r/demo/mictofun_example:alerts"),
			mictofun.Link("All Features", "/r/demo/mictofun_example:features"),
		).
		HR().
		P("Created with " + mictofun.Bold("mictofun") + " | " + mictofun.Link("View Source", "https://github.com/gnolang/gno")).
		String()
}

func renderTables() string {
	// Basic table
	basic := mictofun.NewTable().
		Headers("ID", "Name", "Score").
		Row("1", "Alice", "100").
		Row("2", "Bob", "95").
		Row("3", "Charlie", "87")

	// Aligned table
	aligned := mictofun.NewTable().
		Headers("Left", "Center", "Right").
		SetAlign(mictofun.AlignLeft, mictofun.AlignCenter, mictofun.AlignRight).
		Row("Text", "Text", "Text").
		Row("More", "More", "More")

	return mictofun.New().
		H1("Table Examples").
		P(mictofun.Link("← Back to Home", "/r/demo/mictofun_example")).
		H2("Basic Table").
		Add(basic).
		H2("Aligned Table").
		P("Columns can be aligned left, center, or right:").
		Add(aligned).
		H2("Code").
		CodeBlock(`table := mictofun.NewTable().
    Headers("Name", "Value").
    SetAlign(mictofun.AlignLeft, mictofun.AlignRight).
    Row("Price", "$100").
    Row("Quantity", "5")`, "gno").
		String()
}

func renderAlerts() string {
	return mictofun.New().
		H1("Alert Examples").
		P(mictofun.Link("← Back to Home", "/r/demo/mictofun_example")).
		H2("Available Alert Types").
		Note("This is a NOTE alert for general information.").
		Tip("This is a TIP alert for helpful suggestions.").
		Important("This is an IMPORTANT alert for critical information.").
		Warning("This is a WARNING alert for potential issues.").
		Caution("This is a CAUTION alert for dangerous actions.").
		H2("Code").
		CodeBlock(`doc := mictofun.New().
    Note("General information").
    Warning("Be careful!").
    Tip("Pro tip here")`, "gno").
		String()
}

func renderFeatures() string {
	return mictofun.New().
		H1("All Features Demo").
		P(mictofun.Link("← Back to Home", "/r/demo/mictofun_example")).
		H2("Text Formatting").
		P("You can use " + mictofun.Bold("bold") + ", " + mictofun.Italic("italic") + ", and " + mictofun.Strike("strikethrough") + " text.").
		P("Inline " + mictofun.Code("code") + " is also supported.").
		H2("Lists").
		H3("Bullet List").
		Bullet("First item", "Second item", "Third item with\nmultiple lines").
		H3("Numbered List").
		Numbered("Step one", "Step two", "Step three").
		H3("Todo List").
		Todo([]string{"Completed task", "Pending task", "Another pending"}, []bool{true, false, false}).
		H2("Blockquote").
		Quote("This is a blockquote.\nIt can span multiple lines.").
		H2("Code Block").
		CodeBlock(`func main() {
    fmt.Println("Hello, Gno!")
}`, "go").
		H2("Links").
		Bullet(
			mictofun.UserLink("moul")+" - User link with @",
			mictofun.UserLink("g1abc123def456")+" - Address link",
			mictofun.RealmLink("r/demo/wugnot", "Wrapped UGNOT")+" - Realm link",
		).
		H2("Collapsible Section").
		Collapsible("Click to expand", "This content is hidden by default!\n\nYou can put any markdown content here.").
		H2("Horizontal Rule").
		P("Content above").
		HR().
		P("Content below").
		String()
}
