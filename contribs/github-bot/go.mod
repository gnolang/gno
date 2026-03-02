module github.com/gnolang/gno/contribs/github-bot

go 1.24.4

replace github.com/gnolang/gno => ../..

require (
	github.com/gnolang/gno v0.0.0-00010101000000-000000000000
	github.com/google/go-github/v64 v64.0.0
	github.com/migueleliasweb/go-github-mock v1.0.1
	github.com/sethvargo/go-githubactions v1.3.0
	github.com/stretchr/testify v1.11.1
	github.com/xlab/treeprint v1.2.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/peterbourgon/ff/v3 v3.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
