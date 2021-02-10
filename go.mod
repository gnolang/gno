module github.com/gnolang/gno

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/gdamore/tcell v1.4.0
	github.com/gdamore/tcell/v2 v2.1.0
	github.com/jaekwon/testify v1.6.1
	github.com/mattn/go-runewidth v0.0.10
	github.com/stretchr/testify v1.6.1
)

replace github.com/gdamore/tcell => github.com/gnolang/tcell v1.4.0

replace github.com/gdamore/tcell/v2 => github.com/gnolang/tcell/v2 v2.1.0
