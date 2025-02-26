package threads

func Render(_ string) string {
	return `# Gno Threads
Gno Threads is a minimalistic social media platform made for example purposes. 

There are two versions of this app:
- [V1](/r/docs/gno_threads/v1) - handles simple post creation and stores posts in a slice
- [V2](/r/docs/gno_threads/v2) - handles post creation, updating, and deletion, 
and manages storage more efficiently with an AVL tree. V2 also utilizes different p/ packages to handle pagination, 
easier Markdown formatting, etc.`

}

// Original work & inspiration here: https://gno.land/r/leon/fosdem25/microposts
// Find the full tutorial on the official gno.land docs at docs.gno.land
