package proxy

import "std"

var (
	// App is a safe-object exposing the MyApp interface
	// It can be accessed using MsgRun
	// See r/docs/safeobjects
	App MyApp

	owner     = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	changelog []string
)

func SetImplementation(cur realm, a MyApp) {
	if a == nil {
		panic("app is nil")
	}

	// Check the deployer of the implementation
	if std.OriginCaller() != owner {
		panic("unauthorized")
	}

	App = a

	changelog = append(changelog, a.Version())
	std.Emit("UpgradeEvent", "to", a.Version())
}

func CheckSet() bool {
	return App != nil
}

func Render(_ string) string {
	// This is an edge-case because of dependency loading order
	// The proxy should be deployed first, then the implementation should be set
	// Implementations should import the proxy, and the proxy should not import implementations
	if App == nil {
		return "No implementation currently available."
	}

	return App.Render() // The proxy realm shows the render of the current version
}

/// For usability purposes, below are MsgCall wrappers for the App object
// The app is fully functional without them via MsgRun transactions

func Do(cur realm, arg string) {
	App.Do(arg)
}

func Version() string {
	return App.Version()
}
