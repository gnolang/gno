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
		{"s10", "title10", "event3", "author1", "2024-01-01", "## s10.0\n## s10.1"},
		{"s11", "title11", "event1", "author2", "2024-01-15", "## s11.0\n## s11.1"},
		{"s12", "title12", "event2", "author1", "2024-02-01", "## s12.0\n## s12.1"},
		{"s13", "title13", "event1", "author3", "2024-01-20", "## s13.0\n## s13.1"},
		{"s14", "title14", "event3", "author2", "2024-03-01", "## s14.0\n## s14.1"},
		{"s15", "title15", "event2", "author1", "2024-02-15", "## s15.0\n## s15.1\n## s15.2"},
		{"s16", "title16", "event1", "author4", "2024-03-15", "## s16.0\n## s16.1"},
		{"s17", "title17", "event3", "author2", "2024-01-10", "## s17.0\n## s17.1"},
		{"s18", "title18", "event2", "author3", "2024-02-20", "## s18.0\n## s18.1"},
		{"s19", "title19", "event1", "author1", "2024-03-10", "## s19.0\n## s19.1"},
		{"s20", "title20", "event3", "author4", "2024-01-05", "## s20.0\n## s20.1"},
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
	printRender("s15") // view by slug

	// Test slides
	printRender("s15/slides")        // slides by slug
	printRender("s15/slides?page=2") // slides by slug, second page
	printRender("s15/slides?page=3") // slides by slug, third page
	printRender("s15/slides?page=4") // slides by slug, fourth page
	printRender("s15/slides?page=5") // slides by slug, fifth page
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
// | [Date ↑](?order=asc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | 2024-03-15 | [title16](/r/moul/present:s16) | event1 | author4 | 4 |
// | 2024-03-10 | [title19](/r/moul/present:s19) | event1 | author1 | 4 |
// | 2024-03-01 | [title14](/r/moul/present:s14) | event3 | author2 | 4 |
// | 2024-02-20 | [title18](/r/moul/present:s18) | event2 | author3 | 4 |
// | 2024-02-15 | [title15](/r/moul/present:s15) | event2 | author1 | 5 |
// | 2024-02-01 | [title12](/r/moul/present:s12) | event2 | author1 | 4 |
// | 2024-01-20 | [title13](/r/moul/present:s13) | event1 | author3 | 4 |
// | 2024-01-15 | [title11](/r/moul/present:s11) | event1 | author2 | 4 |
// | 2024-01-10 | [title17](/r/moul/present:s17) | event3 | author2 | 4 |
// | 2024-01-05 | [title20](/r/moul/present:s20) | event3 | author4 | 4 |
// **1** | [2](?page=2&order=desc&sort=author)
//
// +-------------------------------
// | PATH: ?order=asc&sort=date
// | RESULT:
// # Presentations
// | [Date ↓](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | 2024-01-01 | [title10](/r/moul/present:s10) | event3 | author1 | 4 |
// | 2024-01-05 | [title20](/r/moul/present:s20) | event3 | author4 | 4 |
// | 2024-01-10 | [title17](/r/moul/present:s17) | event3 | author2 | 4 |
// | 2024-01-15 | [title11](/r/moul/present:s11) | event1 | author2 | 4 |
// | 2024-01-20 | [title13](/r/moul/present:s13) | event1 | author3 | 4 |
// | 2024-02-01 | [title12](/r/moul/present:s12) | event2 | author1 | 4 |
// | 2024-02-15 | [title15](/r/moul/present:s15) | event2 | author1 | 5 |
// | 2024-02-20 | [title18](/r/moul/present:s18) | event2 | author3 | 4 |
// | 2024-03-01 | [title14](/r/moul/present:s14) | event3 | author2 | 4 |
// | 2024-03-10 | [title19](/r/moul/present:s19) | event1 | author1 | 4 |
// **1** | [2](?page=2&order=desc&sort=author)
//
// +-------------------------------
// | PATH: ?order=asc&sort=title
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | 2024-01-01 | [title10](/r/moul/present:s10) | event3 | author1 | 4 |
// | 2024-01-15 | [title11](/r/moul/present:s11) | event1 | author2 | 4 |
// | 2024-02-01 | [title12](/r/moul/present:s12) | event2 | author1 | 4 |
// | 2024-01-20 | [title13](/r/moul/present:s13) | event1 | author3 | 4 |
// | 2024-03-01 | [title14](/r/moul/present:s14) | event3 | author2 | 4 |
// | 2024-02-15 | [title15](/r/moul/present:s15) | event2 | author1 | 5 |
// | 2024-03-15 | [title16](/r/moul/present:s16) | event1 | author4 | 4 |
// | 2024-01-10 | [title17](/r/moul/present:s17) | event3 | author2 | 4 |
// | 2024-02-20 | [title18](/r/moul/present:s18) | event2 | author3 | 4 |
// | 2024-03-10 | [title19](/r/moul/present:s19) | event1 | author1 | 4 |
// **1** | [2](?page=2&order=desc&sort=author)
//
// +-------------------------------
// | PATH: ?order=asc&sort=author
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | 2024-01-01 | [title10](/r/moul/present:s10) | event3 | author1 | 4 |
// | 2024-02-01 | [title12](/r/moul/present:s12) | event2 | author1 | 4 |
// | 2024-02-15 | [title15](/r/moul/present:s15) | event2 | author1 | 5 |
// | 2024-03-10 | [title19](/r/moul/present:s19) | event1 | author1 | 4 |
// | 2024-01-15 | [title11](/r/moul/present:s11) | event1 | author2 | 4 |
// | 2024-03-01 | [title14](/r/moul/present:s14) | event3 | author2 | 4 |
// | 2024-01-10 | [title17](/r/moul/present:s17) | event3 | author2 | 4 |
// | 2024-01-20 | [title13](/r/moul/present:s13) | event1 | author3 | 4 |
// | 2024-02-20 | [title18](/r/moul/present:s18) | event2 | author3 | 4 |
// | 2024-03-15 | [title16](/r/moul/present:s16) | event1 | author4 | 4 |
// | 2024-01-05 | [title20](/r/moul/present:s20) | event3 | author4 | 4 |
//
//
// +-------------------------------
// | PATH: ?order=asc&sort=title&page=2
// | RESULT:
// # Presentations
// | [Date](?order=desc&page=2&sort=date) | [Title](?order=desc&page=2&sort=title) | Event | [Author](?order=desc&page=2&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | 2024-01-05 | [title20](/r/moul/present:s20) | event3 | author4 | 4 |
// [1](?page=1&order=desc&sort=author) | **2**
//
// +-------------------------------
// | PATH: s15
// | RESULT:
// # title15
// **event2**, 15 Feb 2024
//
// by **author1**
//
// [View as slides](/r/moul/present:s15/slides)
//
// ---
// ## s15.0
// ## s15.1
// ## s15.2
//
// ---
// Last edited: 2009-02-13 23:31:30
//
// Uploader: `g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm`
//
// Number of slides: 5
//
// [Edit](/r/moul/present$help&func=Set&author=author1&date=2024-02-15&event=event2&slug=s15&title=title15) | [Delete](/r/moul/present$help&func=Delete&slug=s15)
//
//
//
// +-------------------------------
// | PATH: s15/slides
// | RESULT:
// # title15
// **event2**, 15 Feb 2024
//
// by **author1**
//
// | 1/5 | [Next →](/r/moul/present:s15/slides?page=2)
//
//
//
// +-------------------------------
// | PATH: s15/slides?page=2
// | RESULT:
// # title15
// ## s15.0
//
// [← Prev](/r/moul/present:s15/slides?page=1) | 2/5 | [Next →](/r/moul/present:s15/slides?page=3)
//
//
//
// +-------------------------------
// | PATH: s15/slides?page=3
// | RESULT:
// # title15
// ## s15.1
//
// [← Prev](/r/moul/present:s15/slides?page=2) | 3/5 | [Next →](/r/moul/present:s15/slides?page=4)
//
//
//
// +-------------------------------
// | PATH: s15/slides?page=4
// | RESULT:
// # title15
// ## s15.2
//
// [← Prev](/r/moul/present:s15/slides?page=3) | 4/5 | [Next →](/r/moul/present:s15/slides?page=5)
//
//
//
// +-------------------------------
// | PATH: s15/slides?page=5
// | RESULT:
// # title15
// ## Thank You!
// author1
//
// 🔗 [https://gno\.land/r/moul/present:s15](https://gno.land/r/moul/present:s15)
//
// [← Prev](/r/moul/present:s15/slides?page=4) | 5/5 |
//
//
//
