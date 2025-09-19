package main

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdlist"
)

func main() {
	list := mdlist.Entries{
		mdlist.Entry{
			Title: "Entry 1",
			Body:  "Body 1",
			Links: []string{
				md.Link("Example", "https://example.com"),
			},
		},
		mdlist.Entry{
			Title: "Entry 2",
			Body:  "Body 2",
		},
		mdlist.Entry{
			Title: "Entry 3",
			Body:  "Body 3",
			Links: []string{
				md.Link("Example1", "https://example.com/1"),
				md.Link("Example2", "https://example.com/2"),
			},
		},
		mdlist.Entry{
			Title: "Entry 4",
			Body:  "Body 4",
			Link:  "https://example.com/4",
		},
		mdlist.Entry{
			Title: "Entry 5",
			Body:  "Body 5",
			Link:  "https://example.com/5",
			Links: []string{
				md.Link("Example3", "https://example.com/3"),
			},
		},
	}
	println("Columns:")
	println(list.Columns())
	println("BulletList:")
	println(list.BulletList())
	println("NumberedList:")
	println(list.NumberedList())
	println("Paragraphs:")
	println(list.Paragraphs())
	println("Done.")
}

// Output:
// Columns:
// <gno-columns>
// ## Entry 1
// Body 1
//
// [Example](https://example.com)
// |||
// ## Entry 2
// Body 2
// |||
// ## Entry 3
// Body 3
//
// [Example1](https://example.com/1) [Example2](https://example.com/2)
// |||
// ## [Entry 4](https://example.com/4)
// Body 4
// |||
// ## [Entry 5](https://example.com/5)
// Body 5
//
// [Example3](https://example.com/3)
// </gno-columns>
//
// BulletList:
// - **Entry 1**: *Body 1* [Example](https://example.com)
// - **Entry 2**: *Body 2*
// - **Entry 3**: *Body 3* [Example1](https://example.com/1) [Example2](https://example.com/2)
// - **[Entry 4](https://example.com/4)**: *Body 4*
// - **[Entry 5](https://example.com/5)**: *Body 5* [Example3](https://example.com/3)
//
// NumberedList:
// 1. **Entry 1**: *Body 1* [Example](https://example.com)
// 2. **Entry 2**: *Body 2*
// 3. **Entry 3**: *Body 3* [Example1](https://example.com/1) [Example2](https://example.com/2)
// 4. **[Entry 4](https://example.com/4)**: *Body 4*
// 5. **[Entry 5](https://example.com/5)**: *Body 5* [Example3](https://example.com/3)
//
// Paragraphs:
// **Entry 1**. Body 1 [Example](https://example.com)
//
// **Entry 2**. Body 2
//
// **Entry 3**. Body 3 [Example1](https://example.com/1) [Example2](https://example.com/2)
//
// **[Entry 4](https://example.com/4)**. Body 4
//
// **[Entry 5](https://example.com/5)**. Body 5 [Example3](https://example.com/3)
// Done.
