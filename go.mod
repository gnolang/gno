module github.com/gnolang/gno

go 1.17

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1
	github.com/fortytw2/leaktest v1.3.0
	github.com/gdamore/tcell/v2 v2.1.0
	github.com/gnolang/cors v1.8.1
	github.com/gnolang/overflow v0.0.0-20170615021017-4d914c927216
	github.com/golang/protobuf v1.5.0
	github.com/google/gofuzz v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/gotuna/gotuna v0.6.0
	github.com/jaekwon/testify v1.6.1
	github.com/jmhodges/levigo v1.0.0
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/linxGnu/grocksdb v1.7.0
	github.com/mattn/go-runewidth v0.0.10
	github.com/pelletier/go-toml v1.9.3
	github.com/stretchr/testify v1.7.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c
	go.etcd.io/bbolt v1.3.6
	go.uber.org/multierr v1.8.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/mod v0.4.2
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	golang.org/x/tools v0.1.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.1.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/gdamore/tcell/v2 => github.com/gnolang/tcell/v2 v2.1.0
