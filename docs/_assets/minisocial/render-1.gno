package minisocial

import "gno.land/p/nt/ufmt" // Gno counterpart to `fmt`, for formatting strings

func Render(_ string) string {
	output := "# MiniSocial\n\n" // \n is needed just like in standard Markdown

	// Handle the edge case
	if len(posts) == 0 {
		output += "No posts.\n"
		return output
	}

	// Let's append the text of each post to the output
	for i, post := range posts {
		// Let's append some post metadata
		output += ufmt.Sprintf("#### Post #%d\n\n", i)
		// Add the stringified post
		output += post.String()
		// Add a line break for cleaner UI
		output += "---\n\n"
	}

	return output
}
