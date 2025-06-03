package blog

import (
	"std"

	"gno.land/p/lou/blog"
	"gno.land/p/moul/md"
)

func CreateSampleBlog() *blog.Blog {
	myBlog := blog.NewBlog("Tech Thoughts", "/r/lou/blog")

	post1, _ := blog.NewPost(
		"ai-future",
		"The Future of AI",
		"Artificial Intelligence is transforming every industry, from healthcare to entertainment.",
		"2024-03-01T10:00:00Z",
		[]string{"g1abcde12345"},
		[]string{"ai", "technology", "future"},
	)
	post1.AddCommentToPost(blog.NewComment(
		std.Address("g1abcde12345"),
		"Exciting times ahead!",
	))

	post2, _ := blog.NewPost(
		"blockchain-real-world",
		"Blockchain in the Real World",
		"Beyond cryptocurrencies, blockchain is finding uses in supply chains, voting systems, and more.",
		"2023-12-15T08:30:00Z",
		[]string{"g1xyz987654"},
		[]string{"blockchain", "tech", "supplychain"},
	)
	post2.AddCommentToPost(blog.NewComment(
		std.Address("g1xyz987654"),
		"This tech still needs better UX.",
	))

	post3, _ := blog.NewPost(
		"getting-things-done",
		"Getting Things Done in 2025",
		"With distractions everywhere, productivity tools are more essential than ever.",
		"2025-01-10T09:15:00Z",
		[]string{"g1produser1"},
		[]string{"productivity", "lifehacks", "focus"},
	)

	myBlog.AddPostToBlog(post1)
	myBlog.AddPostToBlog(post2)
	myBlog.AddPostToBlog(post3)

	return myBlog
}

func Render(path string) string {
	out := md.H1("Blog")
	myBlog := CreateSampleBlog()
	out += myBlog.Render(path)
	return out
}
