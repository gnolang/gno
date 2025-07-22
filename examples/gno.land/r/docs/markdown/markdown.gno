package markdown

import "strings"

// this package can be used to test markdown rendering engines.

func Render(path string) string {
	output := `# Markdown on Gno

## Introduction

Markdown on Gno is based on standard markdown, but also has some unique features, making it the Gno Flavored Markdown. This document describes the current markdown support in Gno, demonstrating both the syntax and its rendered output.

> [!NOTE]
> Markdown support in Gno is still evolving. New features and improvements will be added in future releases.

## Basic Syntax

### Headings

Headings are created using hash symbols (#). The number of hash symbols indicates the heading level.

±±±markdown
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
±±±

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

### Text Formatting

You can format text using the following syntax:

±±±markdown
**Bold text**
*Italic text*
~~Strikethrough text~~
**Bold and _nested italic_**
***All bold and italic***
±±±

**Bold text**
*Italic text*
~~Strikethrough text~~
**Bold and _nested italic_**
***All bold and italic***

### Links

Links can be created using the following syntax:

±±±markdown
[Link text](https://example.com)
[Link with title](https://example.com "Link title")
±±±

[Link text](https://example.com)
[Link with title](https://example.com "Link title")

XXX: custom CSS for internal/external links and if possible for "in the same realm/namespace".

### Lists

Unordered lists use asterisks, plus signs, or hyphens:

±±±markdown
* Item 1
* Item 2
  * Nested item 1
  * Nested item 2
±±±

* Item 1
* Item 2
  * Nested item 1
  * Nested item 2

Ordered lists use numbers:

±±±markdown
1. First item
2. Second item
   1. Nested item 1
   2. Nested item 2
±±±

1. First item
2. Second item
   1. Nested item 1
   2. Nested item 2

### Blockquotes

Blockquotes are created using the > character:

±±±markdown
> This is a blockquote
>
> It can span multiple lines
±±±

> This is a blockquote
>
> It can span multiple lines

### Code

Inline code uses single backticks:

±±±markdown
Use ±func main()± to define the entry point.
±±±

Use ±func main()± to define the entry point.

Code blocks use triple backticks with an optional language identifier:

<!-- XXX: make the example with 'gno' instead of 'go' -->

±±±markdown
±±±go
package main

func main() {
    println("Hello, Gno!")
}
±±±

±±±go
package main

func main() {
    println("Hello, Gno!")
}
±±±

### Horizontal Rules

Horizontal rules are created using three or more asterisks, hyphens, or underscores:

±±±markdown
---
±±±

---

### Task Lists

Gno supports task lists for tracking to-do items:

±±±markdown
- [x] Completed task
- [ ] Pending task
- [ ] Another pending task
- [x] Another completed task
±±±

- [x] Completed task
- [ ] Pending task
- [ ] Another pending task
- [x] Another completed task

### Footnotes

Gno supports footnotes for adding references and citations:

±±±markdown
Here is a sentence with a footnote[^1].

[^1]: This is the footnote content.
±±±

Here is a sentence with a footnote[^1].

[^1]: This is the footnote content.

You can also use multiple footnotes in the same document:

±±±markdown
This is the first sentence with a footnote[^1].
This is another sentence with a different footnote[^2].

[^1]: First footnote content.
[^2]: Second footnote content with more details.
±±±

This is the first sentence with a footnote[^1].
This is another sentence with a different footnote[^2].

[^1]: First footnote content.
[^2]: Second footnote content with more details.

## Tables

Tables are created using pipes and hyphens:

±±±markdown
| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
±±±

| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |

## Images

Images can be included using the following syntax:

±±±markdown
![Alt text](/public/imgs/gnoland.svg "Optional title")
±±±

![Alt text](/public/imgs/gnoland.svg "Optional title")

Currently, only a short list of content providers are supported:

±±±markdown
// Gno-related hosts
"https://gnolang.github.io"
"https://assets.gnoteam.com"
"https://sa.gno.services"

// Other providers should respect DMCA guidelines
// NOTE: Feel free to open a PR to add more providers here :)

// imgur
"https://imgur.com"
"https://*.imgur.com"

// GitHub
"https://*.github.io"
"https://github.com"
"https://*.githubusercontent.com"

// IPFS
"https://ipfs.io"
"https://cloudflare-ipfs.com"
±±±


## Gno-Specific Features

### HTML Support

By design, most typical HTML support is disabled in Gno's markdown implementation. This is an intentional decision for both security and ecosystem cohesion.

While traditional markdown often allows arbitrary HTML tags, Gno Flavored Markdown takes a more controlled approach:

- We may progressively whitelist certain HTML components or add custom ones over time
- Our priority is to enhance our flavored markdown to natively support all essential components
- We aim to eventually support all the initially HTML-supported features, but with syntax that is:
  - More readable when viewing the source directly
  - More integrable with custom browsers such as gnobro in CLI

This approach allows for a more consistent rendering experience across different Gno interfaces while maintaining security and readability as core principles.

### Columns

Gno-Flavored Markdown introduces a column layout system that uses special tags. You can create columns with
±<gno-columns>± blocks and separate the content with ±|||±.

#### Basic Syntax
±±±markdown
<gno-columns>
Left content

|||

Right content
</gno-columns>
±±±

- Renders as:

---

<gno-columns>
Left content

|||

Right content
</gno-columns>

---

#### Key Features

1. **Multiple Columns**:

Depending on the screen size. A maximum of four rows can be displayed in
one row.

***Note**: Any overflow will result in items being displayed in the next row. Also, keep in mind that on smaller screens
(e.g., phones), this overflow may occur with fewer elements.*

±±±markdown
<gno-columns>
First column

|||

Second column

|||

Third column

|||

Fourth column

|||

Fifth column
</gno-columns>
±±±

- Renders as:

---

<gno-columns>
First column

|||

Second column

|||

Third column

|||

Fourth column

|||

Fifth column
</gno-columns>

---

2. **Mixed Content**:

Columns can be mixed with any markdown content

***Note**: Nested columns are currently not possible*

±±±markdown
<gno-columns>
### Text Section
Paragraph with _formatting_

|||

My Image ![Alt](/public/imgs/gnoland.svg)
</gno-columns>
±±±

- Renders as:

---

<gno-columns>
### Text Section
Paragraph with _formatting_

|||

My Image ![Alt](/public/imgs/gnoland.svg)
</gno-columns>

---

### Forms

Gno-Flavored Markdown introduces a secure form system that integrates directly with realms. The system uses custom HTML-like tags that are rendered with proper styling and validation.

#### Basic Usage

Create a form using the ±<gno-form>± tag and add inputs with ±<gno-input>±:

±±±markdown
<gno-form>
  <gno-input name="name" placeholder="Enter your name" />
  <gno-input name="email" placeholder="Enter your email" />
</gno-form>
±±±

<gno-form>
  <gno-input name="name" placeholder="Enter your name" />
  <gno-input name="email" placeholder="Enter your email" />
</gno-form>

#### Form Structure

1. **Form Container**
   - Must start with ±<gno-form>± and end with ±</gno-form>±
   - Optional attributes:
     - ±path±: Render path after form submission (e.g. ±path="user"± will redirect to ±:user?[params]±)
   - Form data is always sent as query parameters
   - Cannot be nested (forms inside forms are not allowed)
   - Automatically includes a header showing the realm name

2. **Input Fields**
   - Created with ±<gno-input>± tag
   - Required attributes:
     - ±name±: Unique identifier for the input (required)
     - ±type±: Type of input (optional, defaults to "text")
     - ±placeholder±: Hint text shown in the input (optional, defaults to "Enter value")
   - Supported input types:
     - ±type="text"± (default): For text input
     - ±type="number"±: For numeric values only (with browser validation)
     - ±type="email"±: For email addresses (with basic browser validation)
     - ±type="tel"±: For phone numbers (no specific validation)
     - ±type="password"±: For password input (masked characters)
   - Note: Input validation is handled by the browser's HTML5 validation
   - Any other type will default to "text"
   - Each input must have a unique name
   - Inputs are rendered with labels and proper styling

#### Example Use Cases

1. Form with Path and Query Parameters
±±±markdown
<gno-form path="user">
  <gno-input name="username" type="text" placeholder="Enter username" />
  <gno-input name="age" type="number" placeholder="Your age" />
</gno-form>
±±±

<gno-form path="user">
  <gno-input name="username" type="text" placeholder="Enter username" />
  <gno-input name="age" type="number" placeholder="Your age" />
</gno-form>

This form will submit to ±:user?username=value&age=value± on the same realm.

2. Form with Query Parameters Only
±±±markdown
<gno-form>
  <gno-input name="recipient" type="text" placeholder="Enter recipient address" />
  <gno-input name="email" type="email" placeholder="Your email" />
  <gno-input name="pswd" type="password" placeholder="Your password" />
</gno-form>
±±±

<gno-form>
  <gno-input name="recipient" type="text" placeholder="Enter recipient address" />
  <gno-input name="email" type="email" placeholder="Your email" />
  <gno-input name="pswd" type="password" placeholder="Your password" />
</gno-form>

This form will submit to ±?recipient=value&email=value&pswd=value± on the same realm.

#### Important Rules

1. **Validation Rules**:
   - Every ±<gno-input>± must have a ±name± attribute
   - Duplicate attribute names are not allowed
   - Forms must be properly closed
   - Nested forms will not render

2. **Security Features**:
   - Forms are processed on the realm where they are defined
   - Each form submission is associated with its realm
   - The realm name is displayed in the form header
   - Input validation is handled by the realm's smart contract

--- 

### Alerts

Alerts are a way to highlight important information in your markdown.

There are five types of alerts:

- INFO
- NOTE
- TIP
- SUCCESS
- WARNING
- CAUTION

**NOTE**: The default alert type is INFO (if a wrong alert type is used).

The alert boxes are expandable/collapsible and can have a title. 
By default, the alert boxes are opened. The title is optional and if not provided, the alert type will be used as the title.

±±±markdown
> [!NOTE]
> This is a note
±±±

> [!NOTE]
> This is a note

±±±markdown
> [!NOTE]-
> This is a closed note
±±±

> [!NOTE]-
> This is a closed note

±±±markdown
> [!NOTE] This is a note with title
> This is a note with title
±±±

> [!NOTE] This is a note with title
> This is a note with title

±±±markdown
> [!INFO]
> This is an info
±±±

> [!INFO]
> This is an info

±±±markdown
> [!TIP]
> This is a tip
±±±

> [!TIP]
> This is a tip

±±±
> [!SUCCESS]
> This is a success
±±±

> [!SUCCESS]
> This is a success

±±±
> [!WARNING]
> This is a warning
±±±

> [!WARNING]
> This is a warning

±±±
> [!CAUTION]
> This is a caution
±±±

> [!CAUTION]
> This is a caution

±±±markdown
> [!FOO]
> The default alert if a wrong alert type is used
±±±

> [!FOO]
> This is the default alert

±±±markdown
> [!NOTE]
> This is a note
> > [!NOTE]
> > This is a scoped note
±±±

> [!NOTE]
> This is a note
> > [!NOTE]
> > This is a scoped note

### User Mentions and Bech32 Addresses

Gno markdown automatically recognizes and renders user mentions and Bech32 addresses as clickable links.

#### User Mentions

// You can mention users using the "@" symbol followed by a username. The username can contain letters, numbers, and underscores.

±±±markdown
Hello @alice, how are you doing?

This is a mention of @bob_user and @test-123.

You can also mention users like @user_name_123.
±±±


Hello @alice, how are you doing?

This is a mention of @bob_user.

You can also mention users like @user_name_123.

#### Bech32 Addresses

Gno markdown can automatically recognize and render valid Bech32 addresses as clickable links.

±±±markdown
This is a GNO address: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

You can also reference g1abc123def456ghi789jkl012mno345pqr678stu901vwx234yz5.
±±±

This is a GNO address: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

You can also reference g1abc123def456ghi789jkl012mno345pqr678stu901vwx234yz5.

### Smart Contract Integration

XXX: TODO

±±±markdown
gno.land/r/boards
gno.land/r/boards:foo/bar
gno.land/r/docs/markdown$source
±±±

gno.land/r/boards
gno.land/r/boards:foo/bar
gno.land/r/docs/markdown$source

### And more...

XXX: TODO

## Future Enhancements

The markdown support in Gno is being actively developed. Future enhancements may include:

- Extended support for custom components
- Interactive elements specific to blockchain functions
- Rich rendering of on-chain data
- Better integration with Gno's package system

[Read more](https://github.com/gnolang/gno/issues/3255)

## Conclusion

Markdown on Gno provides a familiar syntax for developers who have experience with GitHub Flavored Markdown, while adding blockchain-specific extensions that make it more powerful in the context of the Gno platform.

As the Gno ecosystem grows, expect the markdown capabilities to expand accordingly, providing even richer formatting and interactive elements for documentation and user interfaces.

## See Also

For programmatic markdown generation using Go code, check out the [±p/moul/md± demo](/r/docs/moul_md) which demonstrates how to use the ±p/moul/md± package to generate markdown content in your Gno realms.
`
	output = strings.ReplaceAll(output, "±", "`")
	return output
}
