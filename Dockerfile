# build gno
FROM        golang:1.22-alpine AS build-gno
RUN         go env -w GOMODCACHE=/root/.cache/go-build

ENV         GNOROOT="/gnoroot"
RUN         mkdir -p $GNOROOT
WORKDIR     /gnoroot

ADD         go.mod go.sum .
RUN         --mount=type=cache,target=/root/.cache/go-build       go mod download
COPY        . ./
RUN         --mount=type=cache,target=/root/.cache/go-build       go install ./gno.land/cmd/gnoland
RUN         --mount=type=cache,target=/root/.cache/go-build       go install ./gno.land/cmd/gnokey
RUN         --mount=type=cache,target=/root/.cache/go-build       go install ./gno.land/cmd/gnoweb
RUN         --mount=type=cache,target=/root/.cache/go-build       go install ./gnovm/cmd/gno
COPY        . /gnoroot

# runtime-base
FROM        alpine:3.17 AS runtime-base
RUN         apk add ca-certificates
ENV         GNOROOT="/gnoroot"
RUN         mkdir -p $GNOROOT
WORKDIR     /gnoroot

# alpine images
FROM        runtime-base AS gno-alpine
COPY        --from=build-gno /usr/local/go/bin/gno /usr/bin/
ENTRYPOINT  ["/usr/bin/gno"]

FROM        runtime-base AS gnoland-alpine
COPY        --from=build-gno /usr/local/go/bin/gnoland /usr/bin/
EXPOSE      26657 26657
ENTRYPOINT  ["/usr/bin/gnoland"]

FROM        runtime-base AS gnokey-alpine
COPY        --from=build-gno /usr/local/go/bin/gnokey /usr/bin/
ENTRYPOINT  ["/usr/bin/gnokey"]

FROM        runtime-base AS gnofaucet-alpine
COPY        --from=build-faucet /usr/local/go/bin/gnofaucet /usr/bin/
EXPOSE      5050
ENTRYPOINT  ["/usr/bin/gnofaucet"]

FROM        runtime-base AS gnoweb-alpine
COPY        --from=build-gno /usr/local/go/bin/gnoweb /usr/bin/
COPY        --from=build-gno /gnoroot/gno.land/cmd/gnoweb /gnoroot/src/gnoweb
EXPOSE      8888
ENTRYPOINT  ["/usr/bin/gnoweb"]

# all, contains everything.
FROM        runtime-base AS all
COPY        --from=build-gno /usr/local/go/bin/gnoland /usr/bin/
COPY        --from=build-gno /usr/local/go/bin/gnokey /usr/bin/
COPY        --from=build-gno /usr/local/go/bin/gno /usr/bin/
COPY        --from=build-gno /usr/local/go/bin/gnoweb /usr/bin/
# COPY        --from=build-gno /gnoroot/gno.land/cmd/gnoweb /gnoroot/src/gnoweb

COPY        --from=build-gno /gnoroot /gnoroot
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gno /usr/local/go/bin/gofmt /usr/bin
