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

	// Create 15 presentations with some overlapping values and non-sequential dates
	presentations := []struct {
		id      string
		title   string
		event   string
		author  string
		date    string
		content string
	}{
		{"test1", "title1", "event3", "author1", "2024-01-01", "# c1\n## s1"},
		{"test2", "title2", "event1", "author2", "2024-01-15", "# c2\n## s2"},
		{"test3", "title3", "event2", "author1", "2024-02-01", "# c3\n## s3"},
		{"test4", "title4", "event1", "author3", "2024-01-20", "# c4\n## s4"},
		{"test5", "title5", "event3", "author2", "2024-03-01", "# c5\n## s5"},
		{"test6", "title6", "event2", "author1", "2024-02-15", "# c6\n## s6"},
		{"test7", "title7", "event1", "author4", "2024-03-15", "# c7\n## s7"},
		{"test8", "title8", "event3", "author2", "2024-01-10", "# c8\n## s8"},
		{"test9", "title9", "event2", "author3", "2024-02-20", "# c9\n## s9"},
		{"test10", "title10", "event1", "author1", "2024-03-10", "# c10\n## s10"},
		{"test11", "title11", "event3", "author4", "2024-01-05", "# c11\n## s11"},
		{"test12", "title12", "event2", "author2", "2024-02-10", "# c12\n## s12"},
		{"test13", "title13", "event1", "author3", "2024-03-05", "# c13\n## s13"},
		{"test14", "title14", "event3", "author1", "2024-01-25", "# c14\n## s14"},
		{"test15", "title15", "event2", "author4", "2024-02-25", "# c15\n## s15"},
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
	printRender("slides/test5")        // slides by slug
	printRender("slides/test5?page=2") // slides by slug, second page
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
// | [title7](/r/moul/present:view/test7) | author4 | event1 | 2024-03-15 | 2 |
// | [title10](/r/moul/present:view/test10) | author1 | event1 | 2024-03-10 | 2 |
// | [title13](/r/moul/present:view/test13) | author3 | event1 | 2024-03-05 | 2 |
// | [title5](/r/moul/present:view/test5) | author2 | event3 | 2024-03-01 | 2 |
// | [title15](/r/moul/present:view/test15) | author4 | event2 | 2024-02-25 | 2 |
// | [title9](/r/moul/present:view/test9) | author3 | event2 | 2024-02-20 | 2 |
// | [title6](/r/moul/present:view/test6) | author1 | event2 | 2024-02-15 | 2 |
// | [title12](/r/moul/present:view/test12) | author2 | event2 | 2024-02-10 | 2 |
// | [title3](/r/moul/present:view/test3) | author1 | event2 | 2024-02-01 | 2 |
// | [title14](/r/moul/present:view/test14) | author1 | event3 | 2024-01-25 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=date
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title1](/r/moul/present:view/test1) | author1 | event3 | 2024-01-01 | 2 |
// | [title11](/r/moul/present:view/test11) | author4 | event3 | 2024-01-05 | 2 |
// | [title8](/r/moul/present:view/test8) | author2 | event3 | 2024-01-10 | 2 |
// | [title2](/r/moul/present:view/test2) | author2 | event1 | 2024-01-15 | 2 |
// | [title4](/r/moul/present:view/test4) | author3 | event1 | 2024-01-20 | 2 |
// | [title14](/r/moul/present:view/test14) | author1 | event3 | 2024-01-25 | 2 |
// | [title3](/r/moul/present:view/test3) | author1 | event2 | 2024-02-01 | 2 |
// | [title12](/r/moul/present:view/test12) | author2 | event2 | 2024-02-10 | 2 |
// | [title6](/r/moul/present:view/test6) | author1 | event2 | 2024-02-15 | 2 |
// | [title9](/r/moul/present:view/test9) | author3 | event2 | 2024-02-20 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=title
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title ↓](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title1](/r/moul/present:view/test1) | author1 | event3 | 2024-01-01 | 2 |
// | [title10](/r/moul/present:view/test10) | author1 | event1 | 2024-03-10 | 2 |
// | [title11](/r/moul/present:view/test11) | author4 | event3 | 2024-01-05 | 2 |
// | [title12](/r/moul/present:view/test12) | author2 | event2 | 2024-02-10 | 2 |
// | [title13](/r/moul/present:view/test13) | author3 | event1 | 2024-03-05 | 2 |
// | [title14](/r/moul/present:view/test14) | author1 | event3 | 2024-01-25 | 2 |
// | [title15](/r/moul/present:view/test15) | author4 | event2 | 2024-02-25 | 2 |
// | [title2](/r/moul/present:view/test2) | author2 | event1 | 2024-01-15 | 2 |
// | [title3](/r/moul/present:view/test3) | author1 | event2 | 2024-02-01 | 2 |
// | [title4](/r/moul/present:view/test4) | author3 | event1 | 2024-01-20 | 2 |
// **1** | [2](?page=2)
//
// +-------------------------------
// | PATH: ?order=asc&sort=author
// | RESULT:
// # Presentations
// | [Date](?order=desc&sort=date) | [Title](?order=desc&sort=title) | Event | [Author](?order=desc&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title1](/r/moul/present:view/test1) | author1 | event3 | 2024-01-01 | 2 |
// | [title3](/r/moul/present:view/test3) | author1 | event2 | 2024-02-01 | 2 |
// | [title6](/r/moul/present:view/test6) | author1 | event2 | 2024-02-15 | 2 |
// | [title10](/r/moul/present:view/test10) | author1 | event1 | 2024-03-10 | 2 |
// | [title14](/r/moul/present:view/test14) | author1 | event3 | 2024-01-25 | 2 |
// | [title2](/r/moul/present:view/test2) | author2 | event1 | 2024-01-15 | 2 |
// | [title5](/r/moul/present:view/test5) | author2 | event3 | 2024-03-01 | 2 |
// | [title8](/r/moul/present:view/test8) | author2 | event3 | 2024-01-10 | 2 |
// | [title12](/r/moul/present:view/test12) | author2 | event2 | 2024-02-10 | 2 |
// | [title4](/r/moul/present:view/test4) | author3 | event1 | 2024-01-20 | 2 |
// | [title9](/r/moul/present:view/test9) | author3 | event2 | 2024-02-20 | 2 |
// | [title13](/r/moul/present:view/test13) | author3 | event1 | 2024-03-05 | 2 |
// | [title7](/r/moul/present:view/test7) | author4 | event1 | 2024-03-15 | 2 |
// | [title11](/r/moul/present:view/test11) | author4 | event3 | 2024-01-05 | 2 |
// | [title15](/r/moul/present:view/test15) | author4 | event2 | 2024-02-25 | 2 |
//
//
// +-------------------------------
// | PATH: ?order=asc&sort=title&page=2
// | RESULT:
// # Presentations
// | [Date](?order=desc&page=2&sort=date) | [Title ↓](?order=desc&page=2&sort=title) | Event | [Author](?order=desc&page=2&sort=author) | Slides |
// | --- | --- | --- | --- | --- |
// | [title5](/r/moul/present:view/test5) | author2 | event3 | 2024-03-01 | 2 |
// | [title6](/r/moul/present:view/test6) | author1 | event2 | 2024-02-15 | 2 |
// | [title7](/r/moul/present:view/test7) | author4 | event1 | 2024-03-15 | 2 |
// | [title8](/r/moul/present:view/test8) | author2 | event3 | 2024-01-10 | 2 |
// | [title9](/r/moul/present:view/test9) | author3 | event2 | 2024-02-20 | 2 |
// [1](?page=1) | **2**
//
// +-------------------------------
// | PATH: view/test5
// | RESULT:
// # title5
// Event: event3
//
// Author: author2
//
// Date: 2024-03-01
//
// [View as slides](/r/moul/present:slides/test5)
//
// ---
// # c5
// ## s5
//
// ---
// Last edited: 2009-02-13 23:31:30
//
// Slug: `test5`
//
// Uploader: `g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm`
//
// Number of slides: 2
//
//
//
// Admin actions:
//
// - [Edit presentation](/r/moul/present$help&func=Set&slug=test5&title=title5&event=event3&date=2024-03-01)
// - [Delete presentation](/r/moul/present$help&func=Delete&slug=test5)
//
//
// +-------------------------------
// | PATH: slides/test5
// | RESULT:
// # title5
// Event: event3
//
// Author: author2
//
// Date: 2024-03-01
//
//
//
// ---
//
// | 1/2 | [Next →](/r/moul/present:slides/test5?page=2)
//
// +-------------------------------
// | PATH: slides/test5?page=2
// | RESULT:
// ## s5
//
// ---
//
// [← Prev](/r/moul/present:slides/test5?page=1) | 2/2 |
//
