package proxy

import "std"

var (
	app MyApp

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

	app = a

	changelog = append(changelog, a.Version())
	std.Emit("UpgradeEvent", "to", a.Version())
}

func IsImplSet() bool {
	return app != nil
}

func Render(_ string) string {
	// This is an edge-case because of dependency loading order
	// The proxy should be deployed first, then the implementation should be set
	// Implementations should import the proxy, and the proxy should not import implementations
	if app == nil {
		return "No implementation currently available."
	}

	return app.Render() // The proxy realm shows the render of the current version
}

// Properly expose app functionality

func Do(cur realm, arg string) {
	if !IsImplSet() {
		panic("implementation currently not available")
	}

	app.Do(arg)
}

func Version() string {
	if !IsImplSet() {
		panic("implementation currently not available")
	}

	return app.Version()
}
