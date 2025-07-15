package blog

import (
	"gno.land/p/lou/blog"
)

func CreateSampleBlog() *blog.Blog {
	post1, err := blog.NewPost(
		"ai-future",
		"The Future of AI",
		"Artificial Intelligence is transforming every industry, from healthcare to entertainment.",
		"2024-03-01T10:00:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1abcde12345"},
		[]string{"ai", "technology", "future"},
	)
	if err != nil {
		panic(err)
	}
	comm, err := blog.NewComment(
		"g1abcde12345",
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
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"},
		[]string{"blockchain", "tech", "supplychain"},
	)
	if err != nil {
		panic(err)
	}
	reply, err := blog.NewComment(
		"g1ffzxha57dh0qgv9ma5v393ur0zexfvp6lsjpae",
		"heyyy",
	)
	if err != nil {
		panic(err)
	}
	comm2, err := blog.NewComment(
		"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n",
		"This tech still needs better UX.",
	)
	if err != nil {
		panic(err)
	}
	comm3, err := blog.NewComment(
		"g1fjnxhv3v5x02j49fyl49epmhjyzscm8e4au94t",
		"Absolutely, usability is key for adoption.",
	)
	if err != nil {
		panic(err)
	}
	comm4, err := blog.NewComment(
		"g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n",
		"... still needs better UX.",
	)
	if err != nil {
		panic(err)
	}
	post2.AddComment(comm2)
	post2.AddComment(comm3)
	post2.AddReply(comm2.ID(), reply)
	post2.AddReply(reply.ID(), comm4)

	post3, err := blog.NewPost(
		"getting-things-done",
		"Getting Things Done in 2025",
		"With distractions everywhere, productivity tools are more essential than ever.",
		"2025-01-10T09:15:00Z",
		"g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs",
		[]string{"g1produser1"},
		[]string{"productivity", "lifehacks", "focus", "ai"},
	)
	if err != nil {
		panic(err)
	}
	myBlog.AddPost(post1)
	myBlog.AddPost(post2)
	myBlog.AddPost(post3)
	return myBlog
}

func Test(_ realm) {
	CreateSampleBlog()
}
