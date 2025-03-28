module github.com/gnolang/gno/contribs/tessera

go 1.23.0

toolchain go1.24.1

replace github.com/gnolang/gno => ../..

require (
	github.com/gnolang/gno v0.0.0-00010101000000-000000000000
	github.com/goccy/go-yaml v1.16.0
)

require (
	github.com/peterbourgon/ff/v3 v3.4.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
)
