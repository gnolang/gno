package present

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestPresentation(t *testing.T) {
	t.Skip()
	// Test Set
	result := Set(
		"test1",
		"Test Presentation",
		"Test Event",
		"Test Author",
		"2024-01-01",
		"Content\n## Slide 1\n## Slide 2",
	)
	uassert.Equal(t, "presentation saved successfully", result)

	// Test invalid date
	result = Set(
		"test2",
		"Test Presentation 2",
		"Test Event 2",
		"Test Author 2",
		"invalid-date",
		"Content",
	)
	uassert.Equal(t, "400: invalid date format (expected: YYYY-MM-DD)", result)

	// Test Render list
	output := Render("")
	uassert.Equal(t, `# Presentations
Sort by: [Date ↓](?sort=date&order=asc) | [Uploader](?sort=uploader&order=desc) | [Event](?sort=event&order=desc)

| Title | Author | Event | Date | Slides |
|-------|--------|-------|------|--------|
| [Test Presentation](?view=test1) | Test Author | Test Event | 2024-01-01 | 3 |

`, output)

	// Test Render view
	output = Render("view?view=test1")
	uassert.Equal(t, `# Test Presentation
Event: Test Event

Author: Test Author

Date: 2024-01-01

[View as slides](?slides=test1)

---
Content
## Slide 1
## Slide 2

---
Last edited: 2024-01-01 00:00:00

ID: `+"`test1`"+`

Uploader: `+"`g1jg0p4jqv9h`"+`

Number of slides: 3

`, output)

	// Test Render slides (intro)
	output = Render("slides?slides=test1&page=1")
	uassert.Equal(t, `# Test Presentation
Event: Test Event

Author: Test Author

Date: 2024-01-01

---

| 1/3 | [Next →](?slides=test1&page=2)`, output)

	// Test Render slides (middle)
	output = Render("slides?slides=test1&page=2")
	uassert.True(t, strings.Contains(output, "Slide 1"))
	uassert.True(t, strings.Contains(output, "2/3"))
	uassert.True(t, strings.Contains(output, "Prev"))
	uassert.True(t, strings.Contains(output, "Next"))

	// Test Render slides (last)
	output = Render("slides?slides=test1&page=3")
	uassert.True(t, strings.Contains(output, "Slide 2"))
	uassert.True(t, strings.Contains(output, "3/3"))
	uassert.True(t, strings.Contains(output, "Prev"))

	// Test Delete
	result = Delete("test1")
	uassert.Equal(t, "presentation deleted successfully", result)

	// Test Delete non-existent
	result = Delete("nonexistent")
	uassert.Equal(t, "404: presentation not found", result)

	// Test view non-existent
	output = Render("view?view=test1")
	uassert.Equal(t, "404: presentation not found", output)
}

func TestView(t *testing.T) {
	t.Skip()
	// Check if demo presentation was properly initialized
	output := Render("view?view=demo")
	uassert.Equal(t, `# Demo Slides
Event: Demo Event

Author: @demo

Date: 2025-02-02

[View as slides](?slides=demo)

---
# Demo Title

## Slide One
- Point A
- Point B
- Point C

## Slide Two
- Feature 1
- Feature 2
- Feature 3

## Slide Three
- Next step
- Another step
- Final step

---
Last edited: 2024-01-01 00:00:00

ID: `+"`demo`"+`

Uploader: `+"`g1jg0p4jqv9h`"+`

Number of slides: 4

`, output)

	// Check slide navigation
	output = Render("slides?slides=demo&page=1")
	uassert.Equal(t, `# Demo Slides
Event: Demo Event

Author: @demo

Date: 2025-02-02

---

| 1/4 | [Next →](?slides=demo&page=2)`, output)

	output = Render("slides?slides=demo&page=2")
	uassert.Equal(t, `## Slide One
- Point A
- Point B
- Point C

---

[← Prev](?slides=demo&page=1) | 2/4 | [Next →](?slides=demo&page=3)`, output)

	output = Render("slides?slides=demo&page=3")
	uassert.Equal(t, `## Slide Two
- Feature 1
- Feature 2
- Feature 3

---

[← Prev](?slides=demo&page=2) | 3/4 | [Next →](?slides=demo&page=4)`, output)

	output = Render("slides?slides=demo&page=4")
	uassert.Equal(t, `## Slide Three
- Next step
- Another step
- Final step

---

[← Prev](?slides=demo&page=3) | 4/4`, output)
}
