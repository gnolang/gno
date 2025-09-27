package proxy

// MyApp defines an interface that all versions of my app will follow
// The limitation with this approach is that the interface itself cannot be changed, only the implementation can
// This code does not have to live in the proxy realm
type MyApp interface {
	Render() string
	Do(arg string)
	Version() string
}
