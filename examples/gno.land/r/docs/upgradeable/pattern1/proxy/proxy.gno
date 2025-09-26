package proxy

import "std"

// MyApp defines an interface that all versions of my app will follow
// The limitation with this approach is that the interface itself cannot be changed, only the implementation can
// This code does not have to live in the proxy realm
type MyApp interface {
	Render() string
	Do(arg string)
	Version() string
}

var (
	// App is a safe-object exposing the MyApp interface
	// It can be accessed using MsgRun
	// See r/docs/safeobjects
	App MyApp

	owner = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
)

func SetApp(cur realm, a MyApp) {
	if a == nil {
		panic("app is nil")
	}

	if std.OriginCaller() != owner {
		panic("unauthorized")
	}

	App = a

	std.Emit("Upgraded", "from", a.Version())
}

func Render(_ string) string {
	// This is an edge-case because of dependency loading order
	// Implementations should import the proxy, and the proxy should not import implementations
	if App == nil {
		return "No implementation is set currently."
	}

	return App.Render() // The proxy realm shows the render of the current version
}
