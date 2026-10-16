package minisocial

import "gno.land/p/nt/markdown/sanitize/v0"

func Render(_ string) string {
	output := "# MiniSocial\n\n" // \n is needed just like in standard Markdown

	// Handle the edge case
	if len(posts) == 0 {
		output += "No posts.\n"
		return output
	}

	// Let's append the text of each post to the output
	for _, post := range posts {
		output += sanitize.Block(post.text)
	}

	return output
}
