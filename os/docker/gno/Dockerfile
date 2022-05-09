# This dockerfile will be used to bootstrap the docker changes for the cosmos SDK, and will be built hourly.
# We'll also move faddat/archlinux to an hourly schedule, ensuring that when the images are put to use, they dont' need to update as many packages.

FROM ghcr.io/faddat/archlinux

ENV GOPATH=/go

# everything needed for cosmos development
RUN pacman -Syyu --noconfirm go base-devel git leveldb rocksdb make python snappy lz4 jq wget curl protobuf yarn neovim

# Useful tools
RUN go install mvdan.cc/gofumpt@latest && \
      go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
      cp /go/bin/gofumpt /usr/bin && \
      cp /go/bin/golangci-lint /usr/bin && \
      rm -rf /go

