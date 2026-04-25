all:
	@echo ""
	@echo "============================================================"
	@echo "  WARNING: misc/genproto (genproto1 / pbbindings) is DEPRECATED."
	@echo ""
	@echo "  Use misc/genproto2 instead. genproto1's pbbindings are"
	@echo "  not guaranteed to match the reflect codec byte-for-byte;"
	@echo "  genproto2 is the source-of-truth generator with a tested"
	@echo "  parity contract against the reflect codec."
	@echo "============================================================"
	@echo ""
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
