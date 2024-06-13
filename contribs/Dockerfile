# build gno
FROM        golang:1.22 AS build-gno
RUN         mkdir -p /opt/gno/src /opt/build
WORKDIR     /opt/build
ADD         go.mod go.sum ./
RUN         go mod download
ADD         contribs/gnodev/go.mod contribs/gnodev/go.sum ./contribs/gnodev/
RUN         cd ./contribs/gnodev && go mod download
ADD         . ./
RUN         cd ./contribs/gnodev && go build -o /opt/build/build/gnodev ./cmd/gnodev
RUN         ls -la ./build
ADD         . /opt/gno/src/
RUN         rm -rf /opt/gno/src/.git

# build faucet
FROM        golang:1.22 AS build-faucet
RUN         mkdir -p /opt/gno/src /opt/build
WORKDIR     /opt/build
ADD         contribs/gnofaucet/go.mod contribs/gnofaucet/go.sum ./
RUN         go mod download
ADD         contribs/gnofaucet ./
RUN         go build -o ./build/gnofaucet .

# runtime-base + runtime-tls
FROM        debian:stable-slim AS runtime-base
RUN         apt-get update && apt-get install -y curl
ENV         PATH="${PATH}:/opt/gno/bin" \
            GNOROOT="/opt/gno/src"
WORKDIR     /opt/gno/src
FROM        runtime-base AS runtime-tls
RUN         apt-get update && apt-get install -y expect ca-certificates && update-ca-certificates

FROM        runtime-tls AS gnofaucet-slim
COPY        --from=build-faucet /opt/build/build/gnofaucet /opt/gno/bin/
ENTRYPOINT  ["gnofaucet"]
EXPOSE      5050

FROM        runtime-tls AS gnodev-slim
COPY        --from=build-gno /opt/build/build/gnodev /opt/gno/bin/
ENTRYPOINT  ["gnodev"]
EXPOSE      26657

# all, contains everything.
FROM        runtime-tls AS all
COPY        --from=build-gno /opt/build/build/* /opt/gno/bin/
COPY        --from=build-faucet /opt/build/build/* /opt/gno/bin/
COPY        --from=build-gno /opt/gno/src /opt/gno/src
