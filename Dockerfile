# build gno
FROM        golang:1.22 AS build-gno
RUN         mkdir -p /opt/gno/src /opt/build
WORKDIR     /opt/build
ADD         go.mod go.sum .
RUN         go mod download
ADD         . ./
RUN         go build -o ./build/gnoland   ./gno.land/cmd/gnoland
RUN         go build -o ./build/gnokey    ./gno.land/cmd/gnokey
RUN         go build -o ./build/gnoweb    ./gno.land/cmd/gnoweb
RUN         go build -o ./build/gno       ./gnovm/cmd/gno
RUN         ls -la ./build
ADD         . /opt/gno/src/
RUN         rm -rf /opt/gno/src/.git

# build faucet
FROM        golang:1.22 AS build-faucet
RUN         mkdir -p /opt/gno/src /opt/build
WORKDIR     /opt/build
ADD         contribs/gnofaucet/go.mod contribs/gnofaucet/go.sum .
RUN         go mod download
ADD         contribs/gnofaucet ./
RUN         go build -o ./build/gnofaucet .


# runtime-base + runtime-tls
FROM        debian:stable-slim AS runtime-base
ENV         PATH="${PATH}:/opt/gno/bin" \
            GNOROOT="/opt/gno/src"
WORKDIR     /opt/gno/src
FROM        runtime-base AS runtime-tls
RUN         apt-get update && apt-get install -y expect ca-certificates && update-ca-certificates

# slim images
FROM        runtime-base AS gnoland-slim
WORKDIR     /opt/gno/src/gno.land/
COPY        --from=build-gno /opt/build/build/gnoland /opt/gno/bin/
ENTRYPOINT  ["gnoland"]
EXPOSE      26657 36657

FROM        runtime-base AS gnokey-slim
COPY        --from=build-gno /opt/build/build/gnokey /opt/gno/bin/
ENTRYPOINT  ["gnokey"]

FROM        runtime-base AS gno-slim
COPY        --from=build-gno /opt/build/build/gno /opt/gno/bin/
ENTRYPOINT  ["gno"]

FROM        runtime-tls AS gnofaucet-slim
COPY        --from=build-faucet /opt/build/build/gnofaucet /opt/gno/bin/
ENTRYPOINT  ["gnofaucet"]
EXPOSE      5050

FROM        runtime-tls AS gnoweb-slim
COPY        --from=build-gno /opt/build/build/gnoweb /opt/gno/bin/
COPY        --from=build-gno /opt/gno/src/gno.land/cmd/gnoweb /opt/gno/src/gnoweb
ENTRYPOINT  ["gnoweb"]
EXPOSE      8888

# all, contains everything.
FROM        runtime-tls AS all
COPY        --from=build-gno /opt/build/build/* /opt/gno/bin/
COPY        --from=build-faucet /opt/build/build/* /opt/gno/bin/
COPY        --from=build-gno /opt/gno/src /opt/gno/src
# gofmt is required by `gnokey maketx addpkg`
COPY        --from=build-gno /usr/local/go/bin/gofmt /usr/bin
