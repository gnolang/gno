package main

import (
	"gno.land/r/moul/present"
)

func main() {
	// Cleanup initial state
	ret := present.Delete("demo")
	if ret != "presentation deleted successfully" {
		panic("internal error")
	}

	// Create presentations with IDs from 10-20
	presentations := []struct {
		id      string
		title   string
		event   string
		author  string
		date    string
		content string
	}{
		{"s10", "title10", "event3", "author1", "2024-01-01", "# c10\n## s10"},
		{"s11", "title11", "event1", "author2", "2024-01-15", "# c11\n## s11"},
		{"s12", "title12", "event2", "author1", "2024-02-01", "# c12\n## s12"},
		{"s13", "title13", "event1", "author3", "2024-01-20", "# c13\n## s13"},
		{"s14", "title14", "event3", "author2", "2024-03-01", "# c14\n## s14"},
		{"s15", "title15", "event2", "author1", "2024-02-15", "# c15\n## s15"},
		{"s16", "title16", "event1", "author4", "2024-03-15", "# c16\n## s16"},
		{"s17", "title17", "event3", "author2", "2024-01-10", "# c17\n## s17"},
		{"s18", "title18", "event2", "author3", "2024-02-20", "# c18\n## s18"},
		{"s19", "title19", "event1", "author1", "2024-03-10", "# c19\n## s19"},
		{"s20", "title20", "event3", "author4", "2024-01-05", "# c20\n## s20"},
	}

	for _, p := range presentations {
		result := present.Set(p.id, p.title, p.event, p.author, p.date, p.content)
		if result != "presentation saved successfully" {
			panic("failed to add presentation: " + result)
		}
	}

	// Test different sorting scenarios
	printRender("")                       // default
	printRender("?order=asc&sort=date")   // by date ascending
	printRender("?order=asc&sort=title")  // by title ascending
	printRender("?order=asc&sort=author") // by author ascending (multiple entries per author)

	// Test pagination
	printRender("?order=asc&sort=title&page=2") // second page

	// Test view
	printRender("view/test5") // view by slug

	// Test slides
	printRender("slides/s15")        // slides by slug
	printRender("slides/s15?page=2") // slides by slug, second page
}

// Helper function to print path and render result
func printRender(path string) {
	println("+-------------------------------")
	println("| PATH:", path)
	println("| RESULT:\n" + present.Render(path) + "\n")
}

// Output:
// +-------------------------------
// | PATH:
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title16](/r/moul/present:view/s16) | author4 | event1 | 2024-03-15 | 2 |
// | [title19](/r/moul/present:view/s19) | author1 | event1 | 2024-03-10 | 2 |
// | [title14](/r/moul/present:view/s14) | author2 | event3 | 2024-03-01 | 2 |
// | [title18](/r/moul/present:view/s18) | author3 | event2 | 2024-02-20 | 2 |
// | [title15](/r/moul/present:view/s15) | author1 | event2 | 2024-02-15 | 2 |
// | [title12](/r/moul/present:view/s12) | author1 | event2 | 2024-02-01 | 2 |
// | [title13](/r/moul/present:view/s13) | author3 | event1 | 2024-01-20 | 2 |
// | [title11](/r/moul/present:view/s11) | author2 | event1 | 2024-01-15 | 2 |
// | [title17](/r/moul/present:view/s17) | author2 | event3 | 2024-01-10 | 2 |
// | [title20](/r/moul/present:view/s20) | author4 | event3 | 2024-01-05 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=date
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title10](/r/moul/present:view/s10) | author1 | event3 | 2024-01-01 | 2 |
// | [title20](/r/moul/present:view/s20) | author4 | event3 | 2024-01-05 | 2 |
// | [title17](/r/moul/present:view/s17) | author2 | event3 | 2024-01-10 | 2 |
// | [title11](/r/moul/present:view/s11) | author2 | event1 | 2024-01-15 | 2 |
// | [title13](/r/moul/present:view/s13) | author3 | event1 | 2024-01-20 | 2 |
// | [title12](/r/moul/present:view/s12) | author1 | event2 | 2024-02-01 | 2 |
// | [title15](/r/moul/present:view/s15) | author1 | event2 | 2024-02-15 | 2 |
// | [title18](/r/moul/present:view/s18) | author3 | event2 | 2024-02-20 | 2 |
// | [title14](/r/moul/present:view/s14) | author2 | event3 | 2024-03-01 | 2 |
// | [title19](/r/moul/present:view/s19) | author1 | event1 | 2024-03-10 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=title
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title ↓](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title10](/r/moul/present:view/s10) | author1 | event3 | 2024-01-01 | 2 |
// | [title11](/r/moul/present:view/s11) | author2 | event1 | 2024-01-15 | 2 |
// | [title12](/r/moul/present:view/s12) | author1 | event2 | 2024-02-01 | 2 |
// | [title13](/r/moul/present:view/s13) | author3 | event1 | 2024-01-20 | 2 |
// | [title14](/r/moul/present:view/s14) | author2 | event3 | 2024-03-01 | 2 |
// | [title15](/r/moul/present:view/s15) | author1 | event2 | 2024-02-15 | 2 |
// | [title16](/r/moul/present:view/s16) | author4 | event1 | 2024-03-15 | 2 |
// | [title17](/r/moul/present:view/s17) | author2 | event3 | 2024-01-10 | 2 |
// | [title18](/r/moul/present:view/s18) | author3 | event2 | 2024-02-20 | 2 |
// | [title19](/r/moul/present:view/s19) | author1 | event1 | 2024-03-10 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=author
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title10](/r/moul/present:view/s10) | author1 | event3 | 2024-01-01 | 2 |
// | [title12](/r/moul/present:view/s12) | author1 | event2 | 2024-02-01 | 2 |
// | [title15](/r/moul/present:view/s15) | author1 | event2 | 2024-02-15 | 2 |
// | [title19](/r/moul/present:view/s19) | author1 | event1 | 2024-03-10 | 2 |
// | [title11](/r/moul/present:view/s11) | author2 | event1 | 2024-01-15 | 2 |
// | [title14](/r/moul/present:view/s14) | author2 | event3 | 2024-03-01 | 2 |
// | [title17](/r/moul/present:view/s17) | author2 | event3 | 2024-01-10 | 2 |
// | [title13](/r/moul/present:view/s13) | author3 | event1 | 2024-01-20 | 2 |
// | [title18](/r/moul/present:view/s18) | author3 | event2 | 2024-02-20 | 2 |
// | [title16](/r/moul/present:view/s16) | author4 | event1 | 2024-03-15 | 2 |
// | [title20](/r/moul/present:view/s20) | author4 | event3 | 2024-01-05 | 2 |
//
//
// +-------------------------------
// | PATH: ?order=asc&sort=title&page=2
// | RESULT:
// # Presentations
// | [Date](?order=desc&page=2&sort=date) | [Title ↓](?order=desc&page=2&sort=title) | Event | [Author](?order=desc&page=2&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title20](/r/moul/present:view/s20) | author4 | event3 | 2024-01-05 | 2 |
// [1](?page=1) | **2**
//
// +-------------------------------
// | PATH: view/test5
// | RESULT:
// 404: presentation not found
//
// +-------------------------------
// | PATH: slides/s15
// | RESULT:
// # title15
// Event: event2
//
// Author: author1
//
// Date: 2024-02-15
//
//
//
// ---
//
// | 1/2 | [Next →](/r/moul/present:slides/s15?page=2)
//
// +-------------------------------
// | PATH: slides/s15?page=2
// | RESULT:
// ## s15
//
// ---
//
// [← Prev](/r/moul/present:slides/s15?page=1) | 2/2 |
//
