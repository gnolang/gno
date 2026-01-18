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
>> It can span multiple lines
±±±

> This is a blockquote
>
>> It can span multiple lines

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
     - ±exec±: Specify a realm function to execute (e.g. ±exec="CreatePost"±).
       When set, fields must be defined manually and their ±name± attributes must match the function's parameter names in the same order they appear in the function signature.
   - Form data is always sent as query parameters
   - Cannot be nested (forms inside forms are not allowed)
   - Automatically includes a header showing the realm name

2. **Input Fields**
   - Created with ±<gno-input>± tag
   - Attributes:
     - ±name±: Unique identifier for the input (required)
     - ±type±: Type of input (optional, defaults to "text")
     - ±placeholder±: Hint text shown in the input (optional, defaults to "Enter value")
     - ±value±: Default value for non-selectable inputs (optional)
     - ±description±: Description of the input as title (optional - can be used as title for group of one or multiple inputs)
     - ±required±: Set to "true" to mark the field as required (optional)
     - ±readonly±: Set to "true" to make the field read-only (optional)
   - Supported input types:
     - ±type="text"± (default): For text input
     - ±type="number"±: For numeric values only (with browser validation)
     - ±type="email"±: For email addresses (with basic browser validation)
     - ±type="tel"±: For phone numbers (no specific validation)
     - ±type="password"±: For password input (masked characters)
     - ±type="radio"±: For single selection from a group (requires ±value± attribute)
     - ±type="checkbox"±: For multiple selections (requires ±value± attribute)
   - Additional attributes for radio/checkbox:
     - ±value±: The value to submit when selected (required for radio/checkbox)
     - ±checked±: Set to "true" to pre-select the option (optional)
   - Note: Input validation is handled by the browser's HTML5 validation
   - Any other type will default to "text"
   - Each input must have a unique name
   - Inputs are rendered with labels and proper styling

3. **Textarea Fields**
   - Created with ±<gno-textarea>± tag
   - Required attributes:
     - ±name±: Unique identifier for the textarea (required)
     - ±placeholder±: Hint text shown in the textarea (optional, defaults to "Enter value")
   - Optional attributes:
     - ±value±: Default content for the textarea (optional)
     - ±rows±: Number of visible rows (2-10, defaults to 4)
     - ±required±: Set to "true" to mark the field as required (optional)
     - ±readonly±: Set to "true" to make the field read-only (optional)
   - Perfect for longer text input like messages, descriptions, or comments
   - Each textarea must have a unique name
   - Textareas are rendered with labels and proper styling

4. **Select Fields**
   - Created with ±<gno-select>± tag (similar to radio buttons)
   - Required attributes:
     - ±name±: Unique identifier for the select group (required). Use underscores to separate words.
     - ±value±: The value to submit and display text (required)
   - Optional attributes:
     - ±description±: Description of the select group (optional)
     - ±selected±: Set to "true" to pre-select the option (optional)
     - ±required±: Set to "true" to mark the field as required (optional)
   - Multiple ±<gno-select>± elements with the same ±name± form a select group
   - All options with the same ±name± form a group where only one can be selected
   - The ±value± attribute serves both as the submitted value and the displayed text

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
  <gno-input name="recipient" type="text" placeholder="Enter recipient address" description="Recipient Information" />
  <gno-input name="email" type="email" placeholder="Your email" />
  <gno-input name="pswd" type="password" placeholder="Your password" />
</gno-form>
±±±

<gno-form>
  <gno-input name="recipient" type="text" placeholder="Enter recipient address" description="Recipient Information" />
  <gno-input name="email" type="email" placeholder="Your email" />
  <gno-input name="pswd" type="password" placeholder="Your password" />
</gno-form>

This form will submit to ±?recipient=value&email=value&pswd=value± on the same realm.

3. Form with Radio Buttons
±±±markdown
<gno-form>
  <gno-input name="color" type="radio" value="blue" description="What is your favorite color?" />
  <gno-input name="color" type="radio" value="red" />
  <gno-input name="color" type="radio" value="green" />
</gno-form>
±±±

<gno-form>
  <gno-input name="color" type="radio" value="blue" description="What is your favorite color?" />
  <gno-input name="color" type="radio" value="red" />
  <gno-input name="color" type="radio" value="green" />
</gno-form>

Radio buttons allow users to select one option from a group. All radio buttons with the same ±name± form a group where only one can be selected.

4. Form with Checkboxes
±±±markdown
<gno-form>
  <gno-input name="interests" type="checkbox" value="coding" description="What do you like to do?"/>
  <gno-input name="interests" type="checkbox" value="gaming" />
  <gno-input name="interests" type="checkbox" value="reading" />
  <gno-input name="newsletter" type="checkbox" value="subscribe" description="Subscribe to newsletter" placeholder="To get the latest news" checked="true" />
</gno-form>
±±±

<gno-form>
  <gno-input name="interests" type="checkbox" value="coding" description="What do you like to do?" />
  <gno-input name="interests" type="checkbox" value="gaming" />
  <gno-input name="interests" type="checkbox" value="reading" />
  <gno-input name="newsletter" type="checkbox" value="subscribe" description="Subscribe to newsletter" placeholder="To get the latest news" checked="true" />
</gno-form>

Checkboxes allow users to select multiple options. Use ±checked="true"± to pre-select a checkbox.

5. Form with Textarea
±±±markdown
<gno-form>
  <gno-input name="message" type="text" placeholder="Subject" />
  <gno-textarea name="content" placeholder="Enter your message here" rows="6" />
</gno-form>
±±±

<gno-form>
  <gno-input name="message" type="text" placeholder="Subject" />
  <gno-textarea name="content" placeholder="Enter your message here" rows="6" />
</gno-form>

Textareas are perfect for longer text input. The ±rows± attribute controls the height (4-10 rows, default is 4).

6. Form with Select Options
±±±markdown
<gno-form>
  <gno-select name="country" value="United States" description="Select your country" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
</gno-form>
±±±

<gno-form>
  <gno-select name="country" value="United States" description="Select your country" selected="true" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
</gno-form>

Select options work like radio buttons but with a more semantic meaning. All ±<gno-select>± elements with the same ±name± form a group where only one can be selected. Use ±selected="true"± to pre-select an option. The ±value± attribute serves both as the submitted value and the displayed text.

7. Complex Form with Mixed Elements
±±±markdown
<gno-form path="contact">
  <gno-input name="name" type="text" placeholder="Your full name" description="Your Information" />
  <gno-input name="email" type="email" placeholder="Your email address" />
  <gno-input name="age" type="number" placeholder="Your age" />
  
  <gno-input name="color" type="radio" value="blue" description="What is your favorite color?" />
  <gno-input name="color" type="radio" value="red" />
  <gno-input name="color" type="radio" value="green" />
  
  <gno-input name="interests" type="checkbox" value="tech" description="What do you like to do?" />
  <gno-input name="interests" type="checkbox" value="sports" />
  <gno-input name="interests" type="checkbox" value="music"/>
  
  <gno-select name="country" value="us" text="United States" description="Select your country" />
  <gno-select name="country" value="fr" text="France" />
  <gno-select name="country" value="de" text="Germany" />
  
  <gno-textarea name="message" placeholder="Tell us about yourself" rows="5" />
</gno-form>
±±±

<gno-form path="contact">
  <gno-input name="name" type="text" placeholder="Your full name" description="Your Information" />
  <gno-input name="email" type="email" placeholder="Your email address" />
  <gno-input name="age" type="number" placeholder="Your age" />
  
  <gno-input name="color" type="radio" value="blue" description="What is your favorite color?" />
  <gno-input name="color" type="radio" value="red" />
  <gno-input name="color" type="radio" value="green" />
  
  <gno-input name="interests" type="checkbox" value="tech" description="What do you like to do?" />
  <gno-input name="interests" type="checkbox" value="sports" />
  <gno-input name="interests" type="checkbox" value="music"/>
  
  <gno-select name="country" value="us" text="United States" description="Select your country" />
  <gno-select name="country" value="fr" text="France" />
  <gno-select name="country" value="de" text="Germany" />
  
  <gno-textarea name="message" placeholder="Tell us about yourself" rows="5" />
</gno-form>

This example shows how to combine all form element types in a single form.

8. Form with Default Values
±±±markdown
<gno-form>
  <gno-input name="username" type="text" placeholder="Enter username" value="Alice" />
  <gno-input name="age" type="number" placeholder="Your age" value="42" />
  <gno-textarea name="bio" placeholder="Short bio" rows="4" value="Super builder" />
</gno-form>
±±±

<gno-form>
  <gno-input name="username" type="text" placeholder="Enter username" value="Alice" />
  <gno-input name="age" type="number" placeholder="Your age" value="42" />
  <gno-textarea name="bio" placeholder="Short bio" rows="4" value="Super builder" />
</gno-form>

9. Form with Readonly Fields
The ±readonly="true"± attribute makes fields non-editable while displaying their values. This is useful for showing fixed values that users should see but cannot modify.

±±±markdown
<gno-form>
  <gno-input name="username" value="alice" readonly="true" description="Your fixed username" />
  <gno-input name="role" type="text" value="admin" readonly="true" />
  <gno-input name="status" type="radio" value="active" checked="true" readonly="true" description="Account status" />
  <gno-input name="status" type="radio" value="inactive" readonly="true" />
  <gno-input name="verified" type="checkbox" value="yes" checked="true" readonly="true" />
  <gno-textarea name="terms" value="Terms and conditions text..." rows="3" readonly="true" />
  <gno-select name="plan" value="premium" selected="true" readonly="true" description="Your subscription" />
  <gno-select name="plan" value="basic" readonly="true" />
</gno-form>
±±±

<gno-form>
  <gno-input name="username" value="alice" readonly="true" description="Your fixed username" />
  <gno-input name="role" type="text" value="admin" readonly="true" />
  <gno-input name="status" type="radio" value="active" checked="true" readonly="true" description="Account status" />
  <gno-input name="status" type="radio" value="inactive" readonly="true" />
  <gno-input name="verified" type="checkbox" value="yes" checked="true" readonly="true" />
  <gno-textarea name="terms" value="Terms and conditions text..." rows="3" readonly="true" />
  <gno-select name="plan" value="premium" selected="true" readonly="true" description="Your subscription" />
  <gno-select name="plan" value="basic" readonly="true" />
</gno-form>

**Note**: For text-based inputs and textareas, the HTML ±readonly± attribute is used. For radio buttons, checkboxes, and select dropdowns, ±disabled± is used instead since ±readonly± doesn't work on these elements in HTML.

10. Form with Required Fields
The ±required="true"± attribute marks fields as mandatory. Required fields display a "(required)" badge next to their label and use HTML5 browser validation.

±±±markdown
<gno-form>
  <gno-input name="username" placeholder="Enter username" required="true" />
  <gno-input name="email" type="email" placeholder="Your email" required="true" />
  <gno-textarea name="bio" placeholder="Tell us about yourself" rows="4" required="true" />
  <gno-select name="country" value="USA" required="true" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
</gno-form>
±±±

<gno-form>
  <gno-input name="username" placeholder="Enter username" required="true"/>
  <gno-input name="email" type="email" placeholder="Your email" required="true" />
  <gno-textarea name="bio" placeholder="Tell us about yourself" rows="4" required="true" />
  <gno-select name="country" value="USA" required="true" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
</gno-form>

Required fields are validated by the browser before submission. The form cannot be submitted until all required fields are filled.

11. Exec Form with Manual Fields
When using the ±exec± attribute, a command preview appears for the specified function. You must manually define form fields whose ±name± attributes match the function's parameter names, placed in the same order as they appear in the function signature.

±±±markdown
<gno-form exec="CreatePost">
  <gno-input name="title" type="text" placeholder="Post title" />
  <gno-textarea name="content" rows="5" placeholder="Post content"></gno-textarea>
  <gno-input name="tags" type="text" placeholder="Comma-separated tags" />
</gno-form>
±±±

<gno-form exec="CreatePost">
  <gno-input name="title" type="text" placeholder="Post title" />
  <gno-textarea name="content" rows="5" placeholder="Post content"></gno-textarea>
  <gno-input name="tags" type="text" placeholder="Comma-separated tags" />
</gno-form>

The command preview shows the full gnokey command with placeholders for each field value.

12. Exec Form with All Field Types
Here's an example demonstrating all supported field types in a single exec form:

±±±markdown
<gno-form exec="CreateProfile">
  <gno-input name="username" placeholder="Your username" description="Profile Information" />
  <gno-input name="age" type="number" placeholder="Your age" />
  <gno-textarea name="bio" placeholder="Tell us about yourself" rows="4" />
  <gno-select name="country" value="USA" description="Select your country" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
  <gno-input name="visibility" type="radio" value="public" description="Profile visibility" required="true" />
  <gno-input name="visibility" type="radio" value="private" />
  <gno-input name="notifications" type="checkbox" value="email" description="Notifications" checked="true" />
  <gno-input name="notifications" type="checkbox" value="sms" />
</gno-form>
±±±

<gno-form exec="CreateProfile">
  <gno-input name="username" placeholder="Your username" description="Profile Information" value="alice" />
  <gno-input name="age" type="number" placeholder="Your age" />
  <gno-textarea name="bio" placeholder="Tell us about yourself" rows="4" />
  <gno-select name="country" value="USA" description="Select your country" />
  <gno-select name="country" value="France" />
  <gno-select name="country" value="Germany" />
  <gno-input name="visibility" type="radio" value="public" description="Profile visibility" required="true" />
  <gno-input name="visibility" type="radio" value="private" />
  <gno-input name="notifications" type="checkbox" value="email" description="Notifications" checked="true" />
  <gno-input name="notifications" type="checkbox" value="sms" />
</gno-form>

All field types update the command preview in real-time as you interact with them.

#### Important Rules

1. **Validation Rules**:
   - Every ±<gno-input>± must have a ±name± attribute
   - Every ±<gno-textarea>± must have a ±name± attribute
   - Every ±<gno-select>± must have a ±name± and ±value± attribute
   - Duplicate attribute names are not allowed (except for radio, checkbox, and select groups)
   - Forms must be properly closed
   - Nested forms will not render
   - Radio and checkbox inputs must have a ±value± attribute
   - The ±checked± attribute is only valid for radio and checkbox inputs
   - The ±selected± attribute is only valid for ±<gno-select>± elements
   - The ±readonly="true"± attribute makes fields non-editable (works on all field types)
   - The ±required="true"± attribute marks fields as required (works on input, textarea, and select)
   - Textarea ±rows± attribute must be between 2 and 10 (defaults to 4 if invalid or not specified)

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
