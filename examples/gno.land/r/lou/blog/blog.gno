package blog

import (
	"std"

	"gno.land/p/lou/blog"
)

func CreateSampleBlog() *blog.Blog {
	myBlog, err := blog.NewBlog("Tech Thoughts", "/r/lou/blog")
	if err != nil {
		panic(err)
	}
	post1, err := blog.NewPost(
		"ai-future",
		"The Future of AI",
		"Artificial Intelligence is transforming every industry, from healthcare to entertainment.",
		"2024-03-01T10:00:00Z",
		[]string{"g1abcde12345"},
		[]string{"ai", "technology", "future"},
		std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"),
	)
	if err != nil {
		panic(err)
	}
	comm, err := blog.NewComment(
		std.Address("g1abcde12345"),
		"Exciting times ahead!",
	)
	if err != nil {
		panic(err)
	}
	post1.AddComment(comm)

	post2, err := blog.NewPost(
		"blockchain-real-world",
		"Blockchain in the Real World",
		"Beyond cryptocurrencies, blockchain is finding uses in supply chains, voting systems, and more.",
		"2023-12-15T08:30:00Z",
		[]string{"g1xyz987654"},
		[]string{"blockchain", "tech", "supplychain"},
		std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"),
	)
	if err != nil {
		panic(err)
	}
	comm2, err := blog.NewComment(
		std.Address("g1xyz987654"),
		"This tech still needs better UX.",
	)
	if err != nil {
		panic(err)
	}
	post2.AddComment(comm2)

	post3, err := blog.NewPost(
		"getting-things-done",
		"Getting Things Done in 2025",
		"With distractions everywhere, productivity tools are more essential than ever.",
		"2025-01-10T09:15:00Z",
		[]string{"g1produser1"},
		[]string{"productivity", "lifehacks", "focus"},
		std.Address("g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"),
	)
	if err != nil {
		panic(err)
	}
	myBlog.AddPost(post1)
	myBlog.AddPost(post2)
	myBlog.AddPost(post3)
	return myBlog
}

func Render(path string) string {
	out := ""
	myBlog := CreateSampleBlog()
	out += myBlog.Render(path)
	return out
}
