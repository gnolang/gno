// Package main
// This is a filetest. It's somewhat similar to Go's 'Example' files/functions. It's there to showcase functionality and print the output of a specific scenario related to the package it's testing. In this case, it's related to the `Render` function.
package main

import "gno.land/r/docs/moul_md"

func main() {
	// Test the main Render function
	println(moul_md.Render(""))
}

// Output:
// # Package Demo: `p/moul/md`
// This document demonstrates the features of the [gno\.land/p/moul/md](https://gno.land/p/moul/md) package, showing both the source code and the rendered result for each feature.
//
//
// > **Note:** The md package provides helper functions to generate markdown programmatically, making it easier to create dynamic documentation and content.
//
// ## Text Formatting
//
// ### Bold Text
// ```go
// md.Bold("Important text")
// ```
// Result: **Important text**
//
//
// ### Italic Text
// ```go
// md.Italic("Emphasized text")
// ```
// Result: *Emphasized text*
//
//
// ### Strikethrough Text
// ```go
// md.Strikethrough("Deprecated feature")
// ```
// Result: ~~Deprecated feature~~
//
//
// ### Combined Formatting
// ```go
// md.Bold(md.Italic("Very important"))
// ```
// Result: ***Very important***
//
//
// ## Headers
// The package supports all six levels of markdown headers:
//
//
// ```go
// md.H1("Main Title")
// md.H2("Section")
// md.H3("Subsection")
// md.H4("Sub-subsection")
// md.H5("Minor heading")
// md.H6("Smallest heading")
// ```
// Result:
// # Main Title
// ## Section
// ### Subsection
// #### Sub-subsection
// ##### Minor heading
// ###### Smallest heading
//
// ## Lists
//
// ### Bullet Lists
// ```go
// items := []string{
//     "First item",
//     "Second item",
//     "Third item with\nmultiple lines",
// }
// md.BulletList(items)
// ```
// Result:
// - First item
// - Second item
// - Third item with
//   multiple lines
//
// ### Ordered Lists
// ```go
// steps := []string{
//     "First step",
//     "Second step",
//     "Third step with\nadditional details",
// }
// md.OrderedList(steps)
// ```
// Result:
// 1. First step
// 2. Second step
// 3. Third step with
//    additional details
//
// ### Todo Lists
// ```go
// tasks := []string{
//     "Completed task",
//     "Another completed task",
//     "Pending task",
//     "Another pending task",
// }
// completed := []bool{true, true, false, false}
// md.TodoList(tasks, completed)
// ```
// Result:
// - [x] Completed task
// - [x] Another completed task
// - [ ] Pending task
// - [ ] Another pending task
//
// ### Nested Lists
// ```go
// md.BulletItem("Parent item") +
// md.Nested(
//     md.BulletItem("Nested item 1") +
//     md.BulletItem("Nested item 2") +
//     md.Nested(
//         md.BulletItem("Deeply nested"),
//         "    ",
//     ),
//     "  ",
// )
// ```
// Result:
// - Parent item
//   - Nested item 1
//   - Nested item 2
//       - Deeply nested
//
// ## Links and Images
//
// ### Regular Links
// ```go
// md.Link("Gno Homepage", "https://gno.land")
// ```
// Result: [Gno Homepage](https://gno.land)
//
//
// ### User Links
// ```go
// md.UserLink("moul")
// ```
// Result: [@moul](/u/moul)
//
//
// ```go
// md.UserLink("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
// ```
// Result: [g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5](/u/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5)
//
//
// ### Images
// ```go
// md.Image("Gno Logo", "/public/imgs/gnoland.svg")
// ```
// Result:
// ![Gno Logo](/public/imgs/gnoland.svg)
// ### Clickable Images
// ```go
// md.InlineImageWithLink("Click me!", "/public/imgs/gnoland.svg", "https://gno.land")
// ```
// Result:
// [![Click me\!](/public/imgs/gnoland.svg)](https://gno.land)
// ## Code
//
// ### Inline Code
// ```go
// "Use " + md.InlineCode("gno test") + " to run tests"
// ```
// Result: Use `gno test` to run tests
//
//
// ### Simple Code Block
// ```go
// md.CodeBlock("func main() {\n    println(\"Hello, Gno!\")\n}")
// ```
// Result:
// ```
// func main() {
//     println("Hello, Gno!")
// }
// ```
// ### Language-specific Code Block
// ```go
// md.LanguageCodeBlock("go", "package main\n\nfunc main() {\n    println(\"Hello!\")\n}")
// ```
// Result:
// ```go
// package main
//
// func main() {
//     println("Hello!")
// }
// ```
// ## Blockquotes
// ```go
// md.Blockquote("This is an important note.")
// ```
// Result:
// > This is an important note.
//
// ```go
// md.Blockquote("Multi-line quotes\nare also supported\nwith proper formatting.")
// ```
// Result:
// > Multi-line quotes
// > are also supported
// > with proper formatting.
//
// ## Advanced Features
//
// ### Collapsible Sections
// ```go
// md.CollapsibleSection("Click to expand", "Hidden content here!")
// ```
// Result:
// <details><summary>Click to expand</summary>
//
// Hidden content here!
// </details>
//
// ### Two Columns
// ```go
// md.Columns([]string{
//     "Left column content",
//     "Right column content",
// }, false)
// ```
// Result:
// <gno-columns>
// Left column content
// |||
// Right column content
// </gno-columns>
//
// ### Multiple Columns (3 per row)
// ```go
// md.ColumnsN([]string{
//     "Item 1", "Item 2", "Item 3",
//     "Item 4", "Item 5", "Item 6",
// }, 3, true)
// ```
// Result:
// <gno-columns>
// Item 1
// |||
// Item 2
// |||
// Item 3
// </gno-columns>
// <gno-columns>
// Item 4
// |||
// Item 5
// |||
// Item 6
// </gno-columns>
//
// ## Utility Functions
//
// ### Horizontal Rule
// ```go
// md.HorizontalRule()
// ```
// Result:
// ---
//
// ### Footnotes
// ```go
// "This has a footnote[1].\n\n" + md.Footnote("1", "Footnote content here.")
// ```
// Result:
// This has a footnote[1].
//
// [1]: Footnote content here.
// ### Paragraphs
// ```go
// md.Paragraph("This ensures proper spacing.")
// ```
// Result:
// This ensures proper spacing.
//
//
// ## Practical Example
// Here's a complete example showing how to build a rich user profile page using columns, images, and various md features:
//
//
// ```go
// // Example function that generates a rich user profile with balanced columns
// func RenderProfile(username, name, bio string, avatar string, tasks []string, completed []bool) string {
//     // First row: Avatar/Basic Info | Bio
//     avatarSection := md.Image(name + " avatar", avatar) + "\n\n" +
//         md.H3("Basic Info") +
//         md.BulletList([]string{
//             md.Bold("Name:") + " " + name,
//             md.Bold("Username:") + " " + md.UserLink(username),
//             md.Bold("Status:") + " 🟢 Active",
//             md.Bold("Joined:") + " January 2024",
//         })
//
//     bioSection := md.H3("Bio") +
//         md.Blockquote(bio)
//
//     // Second row: Tasks | Links
//     tasksSection := md.H3("Current Tasks") +
//         md.TodoList(tasks, completed)
//
//     linksSection := md.H3("Links") +
//         md.BulletList([]string{
//             md.Link("GitHub", "https://github.com/" + username),
//             md.Link("Portfolio", "https://example.com/" + username),
//             md.Link("LinkedIn", "https://linkedin.com/in/" + username),
//         })
//
//     // Combine with main title and two sets of columns
//     return md.H1("User Profile: " + name) +
//         md.HorizontalRule() +
//         md.Columns([]string{avatarSection, bioSection}, true) +
//         "\n" +
//         md.Columns([]string{tasksSection, linksSection}, true)
// }
//
// // Usage:
// profile := RenderProfile(
//     "johndoe",
//     "John Doe",
//     "Passionate Gno developer building the future of smart contracts. Love working with blockchain technology and contributing to open source.",
//     "/public/imgs/gnoland.svg",
//     []string{"Complete Gno tutorial", "Build first realm", "Deploy to testnet", "Write documentation"},
//     []bool{true, true, false, false},
// )
// ```
// Result when called with the above parameters:
// # User Profile: John Doe
// ---
// <gno-columns>
// ![John Doe avatar](/public/imgs/gnoland.svg)
//
// ### Basic Info
// - **Name:** John Doe
// - **Username:** [@johndoe](/u/johndoe)
// - **Status:** 🟢 Active
// - **Joined:** January 2024
//
// |||
// ### Bio
// > Passionate Gno developer building the future of smart contracts. Love working with blockchain technology and contributing to open source.
//
// |||
//
// |||
//
// </gno-columns>
//
// <gno-columns>
// ### Current Tasks
// - [x] Complete Gno tutorial
// - [x] Build first realm
// - [ ] Deploy to testnet
// - [ ] Write documentation
//
// |||
// ### Links
// - [GitHub](https://github.com/johndoe)
// - [Portfolio](https://example.com/johndoe)
// - [LinkedIn](https://linkedin.com/in/johndoe)
//
// |||
//
// |||
//
// </gno-columns>
//
// ## Best Practices
// 1. Prioritize code readability over performance in Render() - use clear variable names, logical grouping, and readable concatenation for humans
// 2. Use the md package for all programmatic markdown generation in Render() methods
// 3. Combine functions for complex formatting: `md.Bold(md.Italic("text"))`
// 4. Use `md.Paragraph()` to ensure proper spacing between elements
// 5. Leverage `md.ColumnsN()` for responsive layouts
//
// ## See Also
// - [gno\.land/p/moul/md](https://gno.land/p/moul/md) - The md package source code
// - [Markdown on Gno](https://gno.land/r/docs/markdown) - Comprehensive guide on Gno Flavored Markdown syntax
