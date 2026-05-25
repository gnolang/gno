all:
	@echo ""
	@echo "============================================================"
	@echo "  WARNING: misc/genproto (genproto1 / pbbindings) is DEPRECATED"
	@echo "  for the gno.land project."
	@echo ""
	@echo "  Use misc/genproto2 instead. genproto1's pbbindings are"
	@echo "  not guaranteed to match the reflect codec byte-for-byte;"
	@echo "  genproto2 is the source-of-truth generator with a tested"
	@echo "  parity contract against the reflect codec."
	@echo ""
	@echo "  The underlying library tm2/pkg/amino/genproto is kept"
	@echo "  in-tree for future projects that need protobuf3-compatible"
	@echo "  generated code. To revive it, follow"
	@echo "  tm2/pkg/amino/genproto/HARDENING.md."
	@echo "============================================================"
	@echo ""
	find ../.. -name pbbindings.go -delete
	find ../.. -name "*.pb.go"     -delete
	mkdir -p proto
	go run .
	# exception: tm2/pkg/bft/privval/upstream/upstreampb uses protoc directly
	# (production code, not via the amino genproto generator) for byte-compat
	# with upstream Tendermint v0.34's privval socket protocol. Requires
	# google/protobuf/timestamp.proto from protoc's bundled include directory.
	cd ../../tm2/pkg/bft/privval/upstream/upstreampb; \
		protoc --go_out=. --go_opt=paths=source_relative \
		--proto_path=. --proto_path=$$(dirname $$(which protoc))/../include \
		upstream.proto
	# exception: gen .pb.go for folders with .proto files but no package.go.
	cd ../../tm2/pkg/amino/tests/proto3; \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative proto/compat.proto
