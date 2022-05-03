FROM golang:1.18.1 as build

RUN mkdir /opt/src /opt/build
ADD . /opt/src/
ADD . /opt/build/
WORKDIR /opt/build
RUN make


FROM debian:stable-slim

COPY --from=build /opt/build/build/* /opt/gno/bin/
COPY --from=build /opt/src /opt/gno/src
ENV PATH="${PATH}:/opt/gno/bin"
WORKDIR /opt/gno/src
