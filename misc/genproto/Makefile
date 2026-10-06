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
	# exception (must run before `go run .`): the genproto1 binary transitively
	# imports tm2/pkg/bft/privval — which imports the upstreampb package below,
	# so its .pb.go has to exist before genproto1 can compile.
	# tm2/pkg/bft/privval/upstream/upstreampb uses protoc directly (production
	# code, not via the amino genproto generator) for byte-compat with upstream
	# Tendermint v0.34's privval socket protocol. Requires
	# google/protobuf/timestamp.proto from protoc's bundled include directory.
	# Post-protoc steps:
	#  1. Strip the protoc / protoc-gen-go version lines from the header —
	#     they vary across CI vs. developer machines and would otherwise
	#     trip the ci-codegen-verify "no diff after make -C misc/genproto"
	#     check on every protoc-version skew.
	#  2. Apply goimports — protoc-gen-go emits external imports above
	#     stdlib imports in one block; goimports moves them after stdlib
	#     and tm2's `make fmt` expects that ordering.
	cd ../../tm2/pkg/bft/privval/upstream/upstreampb; \
		protoc --go_out=. --go_opt=paths=source_relative \
		--proto_path=. --proto_path=$$(dirname $$(which protoc))/../include \
		upstream.proto; \
		sed -i.bak -e '/^\/\/ \tprotoc-gen-go v/d' -e '/^\/\/ \tprotoc        v/d' upstream.pb.go; \
		rm -f upstream.pb.go.bak; \
		go run -modfile ../../../../../../misc/devdeps/go.mod \
		golang.org/x/tools/cmd/goimports -w upstream.pb.go
	go run .
	# exception: gen .pb.go for folders with .proto files but no package.go.
	cd ../../tm2/pkg/amino/tests/proto3; \
		protoc --go_out=. --go_opt=paths=source_relative \
		--go_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_opt=Mproto/compat.proto=github.com/gnolang/gno/tm2/pkg/amino/tests/proto3 \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative proto/compat.proto
