module linter

go 1.22.7

toolchain go1.22.10

require (
	github.com/gnolang/gno v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.10.0
	golang.org/x/sync v0.10.0
	mvdan.cc/xurls/v2 v2.5.0
)

replace github.com/gnolang/gno => ../..

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/peterbourgon/ff/v3 v3.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
