package main

import (
	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

func main() {
	data := []byte(`{
		"store": {
			"book": [
			{
				"category": "reference",
				"author": "Nigel Rees",
				"title": "Sayings of the Century",
				"price": 8.95
			},
			{
				"category": "fiction",
				"author": "Herman Melville",
				"title": "Moby Dick",
				"isbn": "0-553-21311-3",
				"price": 8.99
			},
			{
				"category": "fiction",
				"author": "J.R.R. Tolkien",
				"title": "The Lord of the Rings",
				"isbn": "0-395-19395-8",
				"price": 22.99
			}
			],
			"bicycle": {
			"color": "red",
			"price": 19.95
			}
		},
		"expensive": 10
	}`)

	paths := []string{
		"$.store.*",    			// All direct properties of `store` (not recursive)
		"$.store.bicycle.color", 	// The color of the bicycle in the store (result: red)
		"$.store.book[*]",			// All books in the store
		"$.store.book[0].title",	// The title of the first book

		"$.store..price",			// The prices of all items in the store
		"$..price",					// Result: [8.95, 8.99, 22.99, 19.95]
		"$..book[*].title",			// The titles of all books in the store
		"$..book[0]",				// The first book
	}

	for _, path := range paths {
		result, err := json.Path(data, path)
		if err != nil {
			ufmt.Println(err)
			return
		}

		ufmt.Println("Path:", path)

		for _, node := range result {
			ufmt.Println(node.String())
		}

		ufmt.Println()
	}
}

// output:
//
// Path: $.store.*
// {
//    "color": "red",
//    "price": 19.95
// }
// [
//   {
//    "category": "reference",
//    "author": "Nigel Rees",
//    "title": "Sayings of the Century",
//    "price": 8.95
//   },
//   {
//    "category": "fiction",
//    "author": "Herman Melville",
//    "title": "Moby Dick",
//    "isbn": "0-553-21311-3",
//    "price": 8.99
//   },
//   {
//    "category": "fiction",
//    "author": "J.R.R. Tolkien",
//    "title": "The Lord of the Rings",
//    "isbn": "0-395-19395-8",
//    "price": 22.99
//   }
//   ]
//
// Path: $.store.bicycle.color
// "red"
//
// Path: $.store.book[*]
// {
//    "category": "reference",
//    "author": "Nigel Rees",
//    "title": "Sayings of the Century",
//    "price": 8.95
// }
// {
//    "category": "fiction",
//    "author": "Herman Melville",
//    "title": "Moby Dick",
//    "isbn": "0-553-21311-3",
//    "price": 8.99
// }
// {
//    "category": "fiction",
//    "author": "J.R.R. Tolkien",
//    "title": "The Lord of the Rings",
//    "isbn": "0-395-19395-8",
//    "price": 22.99
// }
//
// Path: $.store.book[0].title
// "Sayings of the Century"
//
// Path: $.store..price
// 19.95
// 8.95
// 8.99
// 22.99
//
// Path: $..price
// 19.95
// 8.95
// 8.99
// 22.99
//
// Path: $..book[*].title
// "Sayings of the Century"
// "Moby Dick"
// "The Lord of the Rings"
//
// Path: $..book[0]
// {
//    "category": "reference",
//    "author": "Nigel Rees",
//    "title": "Sayings of the Century",
//    "price": 8.95
// }
