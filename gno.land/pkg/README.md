gno.land/pkg/ contains Go packages for gnoland. In a lot of cases a separate `cmd/` exists.

Here are the packages:

* **gnoclient** - rpc client package
* **gnoland** - package to start a gnoland chain (possibly in-memory)
* **gnoweb** - package to serve static Markdown files and render gno realms, as used by https://gno.land
* **integration** - package extending [go-internal/testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript) to run txtar tests against gnoland systems
* **keycli** - package extending tm2/keys/client with GNO specific features (`addpkg`, `run`, `maketx`)
* **log** - package for logging json, console, tests [using go.uber.org/zap](https://pkg.go.dev/go.uber.org/zap)
* **sdk/vm** - package for high-level usage of gnovm
