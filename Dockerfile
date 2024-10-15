# build gno
FROM        golang:1.22-alpine AS build-gno
RUN         go env -w GOMODCACHE=/root/.cache/go-build
WORKDIR     /gnoroot
ENV         GNOROOT="/gnoroot"
COPY        . ./
RUN         --mount=type=cache,target=/root/.cache/go-build       go mod download
RUN         --mount=type=cache,target=/root/.cache/go-build       go build -o ./build/gnoland   ./gno.land/cmd/gnoland
RUN         --mount=type=cache,target=/root/.cache/go-build       go build -o ./build/gnokey    ./gno.land/cmd/gnokey
RUN         --mount=type=cache,target=/root/.cache/go-build       go build -o ./build/gnoweb    ./gno.land/cmd/gnoweb
RUN         --mount=type=cache,target=/root/.cache/go-build       go build -o ./build/gno       ./gnovm/cmd/gno

# Base image
FROM        alpine:3.17 AS base
WORKDIR     /gnoroot
ENV         GNOROOT="/gnoroot"
RUN         apk add ca-certificates
CMD         [ "" ]

# alpine images
# gnoland
FROM        base AS gnoland
COPY        --from=build-gno /gnoroot/build/gnoland /usr/bin/gnoland
COPY        --from=build-gno /gnoroot/examples      /gnoroot/examples
COPY        --from=build-gno /gnoroot/gnovm/stdlibs /gnoroot/gnovm/stdlibs
COPY        --from=build-gno /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gno /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
EXPOSE      26656 26657
ENTRYPOINT  ["/usr/bin/gnoland"]

# gnokey
FROM        base AS gnokey
COPY        --from=build-gno /gnoroot/build/gnokey   /usr/bin/gnokey
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gno /usr/local/go/bin/gofmt /usr/bin/gofmt
ENTRYPOINT  ["/usr/bin/gnokey"]

# gno
FROM        base AS gno
COPY        --from=build-gno /gnoroot/build/gno /usr/bin/gno
ENTRYPOINT  ["/usr/bin/gno"]

# gnoweb
FROM        base AS gnoweb
COPY        --from=build-gno /gnoroot/build/gnoweb /usr/bin/gnoweb
COPY        --from=build-gno /opt/gno/src/gno.land/cmd/gnoweb /opt/gno/src/gnoweb
EXPOSE      8888
ENTRYPOINT  ["/usr/bin/gnoweb"]

# all, contains everything.
FROM        base AS all
COPY        --from=build-gno /gnoroot/build/* /usr/bin/
COPY        --from=build-gno /gnoroot/examples      /gnoroot/examples
COPY        --from=build-gno /gnoroot/gnovm/stdlibs /gnoroot/gnovm/stdlibs
COPY        --from=build-gno /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gno /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gno /usr/local/go/bin/gofmt /usr/bin
