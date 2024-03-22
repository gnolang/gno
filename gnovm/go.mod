module github.com/gnolang/gno/gnovm

go 1.21

require (
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/davecgh/go-spew v1.1.1
	github.com/gnolang/gno/tm2 v0.0.0-00010101000000-000000000000
	github.com/jaekwon/testify v1.6.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/rogpeppe/go-internal v1.12.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/multierr v1.10.0
	golang.org/x/mod v0.15.0
	golang.org/x/tools v0.18.0
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/gnolang/overflow v0.0.0-20170615021017-4d914c927216 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/peterbourgon/ff/v3 v3.4.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.17.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gnolang/gno/tm2 => ../tm2
