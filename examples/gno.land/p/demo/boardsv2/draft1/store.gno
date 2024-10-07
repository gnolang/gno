package boardsv2

// NOTE: Maybe we could abstract the location where posts are stored
type PostStore interface {
	Set(*Post) (updated bool)
	Get(level int, path string) (_ *Post, found bool) // NOTE: Level could be a type alias for better semantics
}
