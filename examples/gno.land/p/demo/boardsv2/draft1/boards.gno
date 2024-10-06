package boardsv2

// NOTE: Boards type should be a realm type
type Boards struct {
	// NOTE: Most of all types could be saved within the same AVL tree, prefixing the level
	// NOTE: We provably want different AVL trees (stores)
	// NOTE: Again, we could consider using Node instead of Post, and tree instead of posts (semantics)
	posts avl.Tree // string(Post.Level + slug) -> *Post (post, comment, poll)
}

// NOTE: Iterating only Post type is confusing semantically
// NOTE: Potential need cast/switch/type-check when iterating.
func (b Boards) Iterate(slug string, fn func(*Post) bool) bool         {}
func (b Boards) IteratePosts(slug string, fn func(*Post) bool) bool    {}
func (b Boards) IterateComments(slug string, fn func(*Post) bool) bool {}

// How to map render paths to actual post instances?
//
// AVL KEYS BY LEVEL PREFIX (start/end)
// Boards => 0_ ... 1_
// Posts => 1_BOARD/ ... 2_
// Comments => 2_BOARD/POST/ ... 3_
//
// HOW TO GUESS PREFIX FROM SLUG
// User enters a SLUG => (one part => 1_BOARD)(more than one part => 1_BOARD/POST)
// How to recognize comments? Should be URL accesible? We could use ":" as separator (not optimal)
