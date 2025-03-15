module github.com/gnolang/gno/contribs/tm2backup

go 1.23

toolchain go1.23.6

replace github.com/gnolang/gno => ../..

require (
	connectrpc.com/connect v1.18.1
	github.com/gnolang/gno v0.0.0-00010101000000-000000000000
	github.com/gofrs/flock v0.12.1
	github.com/klauspost/compress v1.18.0
)

require (
	github.com/peterbourgon/ff/v3 v3.4.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
)
