# Gno mod proxy

Go mod proxy implementation that is able to retrieve Gno dependencies from the tendermint v2 network and transpile realms and packages to be included as dependencies on Go programs.

## How to use

- Start a local gnoland server
- Start a gnomodproxy server: `go run cmd/gnomodproxy/`
- Go to faucet  gno realm, create a go.mod file and execute `GOSUMDB="off" GOPROXY="http://localhost:9999" go mod download` to see how it works (this command will be replaced by `gno mod ... in the future`).
- Go to any Go project and execute `GOSUMDB="off" GOPROXY="http://localhost:9999,https://proxy.golang.org,direct" go mod vendor`. this will transpile on the fly the gno dependencies into go.

## API examples

### Precompiled

- Obtain a Go dep: `http://localhost:9999/github.com/gnolang/gno/examples/gno.land/p/demo/blog/@v/v0.0.0.zip`
- Get version info: `http://localhost:9999/github.com/gnolang/gno/examples/gno.land/p/demo/blog/@v/v0.0.0.info`
- List of all versions: `http://localhost:9999/github.com/gnolang/gno/examples/gno.land/p/demo/blog/@v/list`
- Go mod file: `http://localhost:9999/github.com/gnolang/gno/examples/gno.land/p/demo/blog/@v/v0.0.0.mod`

### Gno

- Obtain a Go dep: `http://localhost:9999/gno.land/p/demo/blog/@v/v0.0.0.zip`
- Get version info: `http://localhost:9999/gno.land/p/demo/blog/@v/v0.0.0.info`
- List of all versions: `http://localhost:9999/gno.land/p/demo/blog/@v/list`
- Go mod file: `http://localhost:9999/github.com/gno.land/p/demo/blog/@v/v0.0.0.mod`

## Pending things

- Implement a way to obtain the date when a package or realm was published
- Implement package and realm versioning
- Implement package and realm go.mod
