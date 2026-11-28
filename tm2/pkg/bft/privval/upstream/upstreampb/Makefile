# Regenerates upstream.pb.go with protoc directly, not through the amino
# generators: these types are byte-compatible with upstream Tendermint v0.34's
# privval socket protocol. Needs protoc, with google/protobuf/timestamp.proto
# in its bundled include directory, and protoc-gen-go
# (`make -C misc/devdeps install`). ci-codegen-verify runs this and fails on
# any diff.
#
# After protoc:
#  1. Strip the protoc / protoc-gen-go version lines from the header. They
#     vary across CI and developer machines and would otherwise fail the
#     ci-codegen-verify diff check on every protoc version skew.
#  2. Apply goimports. protoc-gen-go emits external imports above stdlib
#     imports in one block; goimports moves them after stdlib, as tm2's
#     `make fmt` expects.
all:
	protoc --go_out=. --go_opt=paths=source_relative \
		--proto_path=. --proto_path=$$(dirname $$(which protoc))/../include \
		upstream.proto
	sed -i.bak -e '/^\/\/ \tprotoc-gen-go v/d' -e '/^\/\/ \tprotoc        v/d' upstream.pb.go
	rm -f upstream.pb.go.bak
	go run -modfile ../../../../../../misc/devdeps/go.mod \
		golang.org/x/tools/cmd/goimports -w upstream.pb.go
