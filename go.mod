module github.com/gnolang/gno

go 1.23.0

toolchain go1.24.1

require (
	connectrpc.com/connect v1.18.1
	dario.cat/mergo v1.0.1
	github.com/alecthomas/chroma/v2 v2.15.0
	github.com/bendory/conway-hebrew-calendar v0.0.0-20210829020739-dcc34210ce9b
	github.com/btcsuite/btcd/btcec/v2 v2.3.4
	github.com/btcsuite/btcd/btcutil v1.1.6
	github.com/cockroachdb/apd/v3 v3.2.1
	github.com/cockroachdb/pebble v1.1.5
	github.com/cosmos/ics23/go v0.11.0
	github.com/cosmos/ledger-cosmos-go v0.14.0
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0
	github.com/emicklei/dot v1.6.2
	github.com/fortytw2/leaktest v1.3.0
	github.com/gofrs/flock v0.12.1
	github.com/golang/mock v1.6.0
	github.com/google/gofuzz v1.2.0
	github.com/gorilla/websocket v1.5.3
	github.com/klauspost/compress v1.18.0
	github.com/libp2p/go-buffer-pool v0.1.0
	github.com/pelletier/go-toml v1.9.5
	github.com/peterbourgon/ff/v3 v3.4.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/rogpeppe/go-internal v1.14.1
	github.com/rs/cors v1.11.1
	github.com/rs/xid v1.6.0
	github.com/sig-0/insertion-queue v0.0.0-20241004125609-6b3ca841346b
	github.com/stretchr/testify v1.11.1
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/valyala/bytebufferpool v1.0.0
	github.com/yuin/goldmark v1.7.8
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
	github.com/zondax/hid v0.9.2
	go.etcd.io/bbolt v1.3.11
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/sdk/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	go.uber.org/zap/exp v0.3.0
	golang.org/x/crypto v0.41.0
	golang.org/x/mod v0.26.0
	golang.org/x/net v0.43.0
	golang.org/x/sync v0.16.0
	golang.org/x/term v0.34.0
	golang.org/x/text v0.28.0
	golang.org/x/tools v0.35.0
	google.golang.org/protobuf v1.36.8
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/cosmos/gogoproto v1.7.0 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/onsi/gomega v1.26.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.15.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/zondax/ledger-go v0.14.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.1 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250825161204-c5933d9347a5 // indirect
	google.golang.org/grpc v1.75.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
