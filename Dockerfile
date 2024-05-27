# build
FROM        golang:1.21 AS build
RUN         mkdir -p /opt/gno/src /opt/build
WORKDIR     /opt/build
ADD         go.mod go.sum /opt/build/
RUN         go mod download
ADD         . /opt/gno/src/
ADD         . /opt/build/
RUN         go build -o ./build/gnoland ./cmd/gnoland
RUN         go build -o ./build/gnokey ./cmd/gnokey
RUN         go build -o ./build/gnodev ./cmd/gnodev
RUN         go build -o ./build/gnofaucet ./cmd/gnofaucet
RUN         go build -o ./build/gnotxport ./cmd/gnotxport
RUN         cd ./gnoland/website && go build -o ../../build/gnoweb .
RUN         rm -rf /opt/gno/src/.git
RUN         ls -la /opt/build/build/

# runtime-base + runtime-tls
FROM        debian:stable-slim AS runtime-base
ENV         PATH="${PATH}:/opt/gno/bin"
WORKDIR     /opt/gno/src
FROM        runtime-base AS runtime-tls
RUN         apt-get update && apt-get install -y expect ca-certificates && update-ca-certificates

# slim images
FROM        runtime-base AS gnoland-slim
COPY        --from=build /opt/build/build/gnoland /opt/gno/bin/
ENTRYPOINT  ["gnoland"]
EXPOSE      26657 36657

FROM        runtime-base AS gnokey-slim
COPY        --from=build /opt/build/build/gnokey /opt/gno/bin/
ENTRYPOINT  ["gnokey"]

FROM        runtime-base AS gnodev-slim
COPY        --from=build /opt/build/build/gnodev /opt/gno/bin/
ENTRYPOINT  ["gnodev"]

FROM        runtime-tls AS gnofaucet-slim
COPY        --from=build /opt/build/build/gnofaucet /opt/gno/bin/
ENTRYPOINT  ["gnofaucet"]
EXPOSE      5050

FROM        runtime-tls AS gnotxport-slim
COPY        --from=build /opt/build/build/gnotxport /opt/gno/bin/
ENTRYPOINT  ["gnotxport"]

FROM        runtime-tls AS gnoweb-slim
COPY        --from=build /opt/build/build/gnoweb /opt/gno/bin/
COPY        --from=build /opt/gno/src/gnoland/website /opt/gno/src/gnoland/website
ENTRYPOINT  ["gnoweb"]
EXPOSE      8888

# all, contains everything.
FROM        runtime-tls AS all
COPY        --from=build /opt/build/build/* /opt/gno/bin/
COPY        --from=build /opt/gno/src /opt/gno/src
