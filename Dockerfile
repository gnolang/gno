# setup gnocore basics
FROM        golang:1.24-alpine AS setup-gnocore
ENV         GNOROOT="/gnoroot"
ENV         CGO_ENABLED=0 GOOS=linux
WORKDIR     /gnoroot
RUN         go env -w GOMODCACHE=/root/.cache/go-build
# Mod files
COPY        go.mod go.sum ./
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gomodcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
            go mod download -x
COPY        . ./

# build gnocore
FROM        setup-gnocore AS build-gnocore
# Gnoland
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gomodcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
            go build -ldflags "-w -s" -o ./build/gnoland ./gno.land/cmd/gnoland
# Gnokey
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gomodcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
            go build -ldflags "-w -s" -o ./build/gnokey ./gno.land/cmd/gnokey
# Gnoweb
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gomodcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
            go build -ldflags "-w -s" -o ./build/gnoweb ./gno.land/cmd/gnoweb
# Gno
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gomodcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gobuildcache \
            go build -ldflags "-w -s" -o ./build/gno ./gnovm/cmd/gno

# Gnofaucet build
FROM        setup-gnocore AS build-gnofaucet
WORKDIR     /gnoroot/contribs/gnofaucet
RUN         --mount=type=cache,target=/go/pkg/mod/,id=faucet-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=faucet-buildcache \
            go mod download -x
RUN         --mount=type=cache,target=/go/pkg/mod/,id=faucet \
            --mount=type=cache,target=/root/.cache/go-build,id=faucet-buildcache \
            go build -ldflags "-w -s" -o /gnoroot/build/gnofaucet .

# Gnodev build
FROM        setup-gnocore AS build-gnodev
WORKDIR     /gnoroot/contribs/gnodev
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gnodev-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gnodev-buildcache \
            go mod download -x
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gnodev-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gnodev-buildcache \
            go build \
            -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=/gnoroot" \
            -o /gnoroot/build/gnodev .

# Gnobro build
FROM        setup-gnocore AS build-gnobro
WORKDIR     /gnoroot/contribs/gnobro
RUN         --mount=type=cache,target=/go/pkg/mod/,id=gnobro-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=gnobro-buildcache \
            go build \
            -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=/gnoroot" \
            -o /gnoroot/build/gnobro .

# Gnocontribs
## Gnogenesis
FROM        setup-gnocore AS build-contribs
WORKDIR     /gnoroot/contribs/gnogenesis
RUN         --mount=type=cache,target=/go/pkg/mod/,id=contribs_modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=contribs_buildcache \
            go mod download -x
RUN         --mount=type=cache,target=/go/pkg/mod/,id=contribs_modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=contribs_buildcache \
            go build -ldflags "-w -s" -o /gnoroot/build/gnogenesis .
## GnoKMS
WORKDIR     /gnoroot/contribs/gnokms
RUN         --mount=type=cache,target=/go/pkg/mod/,id=kms_modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=kms_buildcache \
            go mod download -x
RUN         --mount=type=cache,target=/go/pkg/mod/,id=kms_modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=kms_buildcache \
            go build -o /gnoroot/build/gnokms .

# Misc build
FROM        setup-gnocore AS build-misc
## Staging
WORKDIR     /gnoroot/misc/loop
RUN         --mount=type=cache,target=/go/pkg/mod/,id=pl-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=pl-buildcache \
            go mod download -x
RUN         --mount=type=cache,target=/go/pkg/mod/,id=pl-modcache \
            --mount=type=cache,target=/root/.cache/go-build,id=pl-buildcache \
            go build -ldflags "-w -s" -o /gnoroot/build/portalloopd ./cmd

# Base image
FROM        alpine:3 AS base
WORKDIR     /gnoroot
ENV         GNOROOT="/gnoroot"
RUN         apk add --no-cache ca-certificates

# Gnoland image
## ghcr.io/gnolang/gno/gnoland
FROM        base AS gnoland
COPY        --from=build-gnocore /gnoroot/build/gnoland                         /usr/bin/gnoland
COPY        --from=build-gnocore /gnoroot/examples                              /gnoroot/examples
COPY        --from=build-gnocore /gnoroot/gnovm/stdlibs                         /gnoroot/gnovm/stdlibs
COPY        --from=build-gnocore /gnoroot/gnovm/tests/stdlibs                   /gnoroot/gnovm/tests/stdlibs
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
EXPOSE      26656 26657
ENTRYPOINT  ["/usr/bin/gnoland"]

# Gnokey image
## ghcr.io/gnolang/gno/gnokey
FROM        base AS gnokey
COPY        --from=build-gnocore /gnoroot/build/gnokey   /usr/bin/gnokey
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gnocore /usr/local/go/bin/gofmt /usr/bin/gofmt
ENTRYPOINT  ["/usr/bin/gnokey"]

# Gnoweb image
## ghcr.io/gnolang/gno/gnoweb
FROM        base AS gnoweb
COPY        --from=build-gnocore /gnoroot/build/gnoweb /usr/bin/gnoweb
EXPOSE      8888
ENTRYPOINT  ["/usr/bin/gnoweb"]

# Gnofaucet image
## ghcr.io/gnolang/gno/gnofaucet
FROM        base AS gnofaucet
COPY        --from=build-gnofaucet /gnoroot/build/gnofaucet /usr/bin/gnofaucet
EXPOSE      5050
ENTRYPOINT  ["/usr/bin/gnofaucet"]

# Gnodev image
## ghcr.io/gnolang/gno/gnodev
FROM        base AS gnodev
COPY        --from=build-gnodev /gnoroot/build/gnodev                       /usr/bin/gnodev
COPY        --from=build-gnocore /gnoroot/examples                              /gnoroot/examples
COPY        --from=build-gnocore /gnoroot/gnovm/stdlibs                         /gnoroot/gnovm/stdlibs
COPY        --from=build-gnocore /gnoroot/gnovm/tests/stdlibs                   /gnoroot/gnovm/tests/stdlibs
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
# gnoweb port exposed by default
EXPOSE     8888
ENTRYPOINT  ["/usr/bin/gnodev"]

# Gno
FROM        base AS gno
COPY        --from=build-gnocore /gnoroot/build/gno           /usr/bin/gno
COPY        --from=build-gnocore /gnoroot/examples            /gnoroot/examples
COPY        --from=build-gnocore /gnoroot/gnovm/stdlibs       /gnoroot/gnovm/stdlibs
COPY        --from=build-gnocore /gnoroot/gnovm/tests/stdlibs /gnoroot/gnovm/tests/stdlibs
ENTRYPOINT  ["/usr/bin/gno"]

# Gno Contribs [ Gnobro, Gnogenesis, GnoKMS ]
## ghcr.io/gnolang/gnocontribs
FROM        base AS gnocontribs
COPY        --from=build-gnobro /gnoroot/build/gnobro                           /usr/bin/gnobro
COPY        --from=build-contribs /gnoroot/build/gnogenesis                     /usr/bin/gnogenesis
COPY        --from=build-contribs /gnoroot/build/gnokms                          /usr/bin/gnokms
COPY        --from=build-gnocore /gnoroot/examples                              /gnoroot/examples
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
EXPOSE     22
ENTRYPOINT [ "/bin/sh", "-c" ]

# misc/loop
FROM        docker AS portalloopd
WORKDIR     /gnoroot
ENV         GNOROOT="/gnoroot"
RUN         apk add --no-cache ca-certificates bash curl jq
COPY        --from=build-misc /gnoroot/build/portalloopd /usr/bin/portalloopd
ENTRYPOINT  ["/usr/bin/portalloopd"]
CMD         ["serve"]

# all, contains everything.
FROM        base AS all
COPY        --from=build-gnocore /gnoroot/build/*                               /usr/bin/
COPY        --from=build-gnocore /gnoroot/examples                              /gnoroot/examples
COPY        --from=build-gnocore /gnoroot/gnovm/stdlibs                         /gnoroot/gnovm/stdlibs
COPY        --from=build-gnocore /gnoroot/gnovm/tests/stdlibs                   /gnoroot/gnovm/tests/stdlibs
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_txs.jsonl    /gnoroot/gno.land/genesis/genesis_txs.jsonl
COPY        --from=build-gnocore /gnoroot/gno.land/genesis/genesis_balances.txt /gnoroot/gno.land/genesis/genesis_balances.txt
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gnocore /usr/local/go/bin/gofmt                        /usr/bin
