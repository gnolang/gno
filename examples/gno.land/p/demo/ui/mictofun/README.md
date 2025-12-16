# mictofun - Fluent Markdown Framework for Gno

A powerful, chainable markdown generation library for gno.land smart contracts.

## Features

- **Fluent Chainable API** - Build documents with method chaining
- **Component Composition** - Combine tables, alerts, and sections
- **Table Builder** - Create tables with alignment support (Left, Center, Right)
- **GitHub-style Alerts** - Note, Tip, Warning, Important, Caution callouts
- **gno.land Integration** - UserLink, RealmLink, Columns helpers
- **Layout Components** - Collapsible sections, centered content, badges

## Quick Start

```gno
import "gno.land/p/demo/ui/mictofun"

func Render(path string) string {
    return mictofun.New().
        H1("Welcome to My Realm").
        P("This is a " + mictofun.Bold("powerful") + " markdown library.").
        H2("Features").
        Bullet("Fluent API", "Chainable methods", "Component composition").
        HR().
        P("Created by " + mictofun.UserLink("mictofun")).
        String()
}
```

## API Reference

### Document Builder

```gno
doc := mictofun.New()          // Create new document
doc.H1("Heading")              // Add heading (H1-H6)
doc.P("Paragraph text")        // Add paragraph
doc.Quote("Quoted text")       // Add blockquote
doc.CodeBlock("code", "go")    // Add code block with language
doc.HR()                       // Add horizontal rule
doc.Bullet("a", "b", "c")      // Add bullet list
doc.Numbered("1", "2", "3")    // Add numbered list
doc.Todo(items, done)          // Add todo list
doc.Img("alt", "url")          // Add image
doc.String()                   // Get final markdown
```

### Inline Formatting (return string)

```gno
mictofun.Bold("text")          // **text**
mictofun.Italic("text")        // *text*
mictofun.Strike("text")        // ~~text~~
mictofun.Code("text")          // `text`
mictofun.Link("text", "url")   // [text](url)
mictofun.Image("alt", "url")   // ![alt](url)
mictofun.UserLink("name")      // [@name](/u/name)
mictofun.RealmLink("path")     // [path](/path)
```

### Tables

```gno
table := mictofun.NewTable().
    Headers("Name", "Amount", "Status").
    SetAlign(mictofun.AlignLeft, mictofun.AlignRight, mictofun.AlignCenter).
    Row("Alice", "$100", "✓").
    Row("Bob", "$250", "Pending")

doc.Add(table)
```

### Alerts (GitHub-style)

```gno
doc.Note("This is a note")
doc.Tip("This is a tip")
doc.Warning("This is a warning")
doc.Important("This is important")
doc.Caution("This is a caution")
```

Output:
```markdown
> [!NOTE]
> This is a note
```

### Layout Components

```gno
doc.Collapsible("FAQ Title", "Answer content")
doc.Cols("Column 1", "Column 2", "Column 3")
doc.Center("Centered content")
doc.Right("Right-aligned content")
```

## Testing

```bash
cd examples
go run ../gnovm/cmd/gno test -v ./gno.land/p/demo/ui/mictofun/...
```

## License

Same as gno.land project.
