package example

func Render(string) string {
	return `# Render Proxy

This example shows how proxying render calls can be used to allow updating realms to new
versions while keeping the same realm path. The idea is to have a simple "parent" realm
that only keeps track of the latest realm version and forwards all render calls to it.

By only focusing on the 'Render()' function the proxy realm keeps its public functions
stable allowing each version to update their public functions and exposed types without
needing to also update the proxy realm.

Any interaction or transaction must be sent to the latest, or target, realm version while
render calls are sent to the proxy realm.

Each realm version registers itself on deployment as the latest available version, and
its allowed to do so because each versioned realm path shares the proxy realm path.
`
}
