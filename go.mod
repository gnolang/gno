module github.com/gnolang/gno

go 1.15

require (
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/gdamore/tcell v1.4.0
	github.com/gdamore/tcell/v2 v2.1.0
	github.com/golang/protobuf v1.4.1
	github.com/google/gofuzz v1.0.0
	github.com/jaekwon/testify v1.6.1
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.10
	github.com/stretchr/testify v1.6.1
	github.com/syndtr/goleveldb v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20200115085410-6d4e4cb37c7d
	google.golang.org/protobuf v1.25.0
)

replace github.com/gdamore/tcell => github.com/gnolang/tcell v1.4.0

replace github.com/gdamore/tcell/v2 => github.com/gnolang/tcell/v2 v2.1.0
