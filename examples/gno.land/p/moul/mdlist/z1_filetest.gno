package main

import (
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdlist"
)

func main() {
	list := mdlist.List{
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
	}
	println("GridPager:")
	println(list.GridPager(1, 10))
	println("ULPager:")
	println(list.ULPager(1, 10))
	println("Done.")
}

// Output:
// GridPager:
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
// </gno-columns>
//
// ULPager:
// - **Entry 1**: *Body 1* [Example](https://example.com)
// - **Entry 2**: *Body 2*
// - **Entry 3**: *Body 3* [Example1](https://example.com/1) [Example2](https://example.com/2)
//
// Done.
