// TODO: Document how plugins work and best practices
package plugin

// NOTE: Consider adding lifecycle methods like `Post` creation, deletion, ...
type Plugin interface {
	Name() string
	Render() string
}
