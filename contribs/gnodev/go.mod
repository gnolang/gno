module github.com/gnolang/gno/contribs/gnodev

go 1.20

replace github.com/gnolang/gno => ../..

require (
	github.com/fsnotify/fsnotify v1.7.0
	golang.org/x/term v0.15.0
)

require golang.org/x/sys v0.15.0 // indirect
