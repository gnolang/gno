# This dockerfile will be used to bootstrap the docker changes for gno, and will be built hourly.

FROM ghcr.io/faddat/archlinux

ENV GOPATH=/go

# everything needed for gno development
RUN pacman -Syyu --noconfirm go base-devel git leveldb rocksdb make snappy lz4 protobuf

# Useful tools
RUN go install mvdan.cc/gofumpt@latest && \
      go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
      cp /go/bin/gofumpt /usr/bin && \
      cp /go/bin/golangci-lint /usr/bin && \
      rm -rf /go

