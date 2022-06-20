FROM golang:1.18.1 as build

RUN mkdir /opt/src /opt/build
WORKDIR /opt/build
ADD go.mod go.sum /opt/build/
RUN go mod download
ADD . /opt/src/
ADD . /opt/build/
RUN go build -o ./build/gnoland ./cmd/gnoland
RUN go build -o ./build/gnokey ./cmd/gnokey
RUN go build -o ./build/gnodev ./cmd/gnodev
RUN go build -o ./build/gnofaucet ./cmd/gnofaucet
RUN cd ./gnoland/website && go build -o ../../build/gnoweb .
RUN rm -rf /opt/src/.git
RUN ls -la /opt/build/build/


FROM debian:stable-slim
RUN apt-get update && apt-get install -y --force-yes expect
COPY --from=build /opt/build/build/* /opt/gno/bin/
COPY --from=build /opt/src /opt/gno/src
ENV PATH="${PATH}:/opt/gno/bin"
WORKDIR /opt/gno/src
