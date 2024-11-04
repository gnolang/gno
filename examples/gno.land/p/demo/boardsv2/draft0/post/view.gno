package post

// Two cases to solve
// - Give me a list of boards (board list page)
// - Give me a list of comments, created by a user accross all boards (user activity page, of a user)
type View interface {
	Name() string
	Size() int
	Iterate(start, end string, fn func(key string, v interface{}) bool) bool
}
