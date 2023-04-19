all:
	find ../.. -name pbbindings.go -delete
	find ../.. -name "*.pb.go"     -delete
	mkdir -p proto
	go run .
	# exception: gen .pb.go for folders with .proto files but no package.go.
	cd ../../tm2/pkg/amino/tests/proto3; \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative proto/compat.proto
