FROM golang:alpine AS builder

COPY . /go/src/github.com/gnolang/gno/misc/loop

WORKDIR /go/src/github.com/gnolang/gno/misc/loop

RUN go build -o /build/portalloopd ./cmd/portalloopd
RUN go build -o /build/autocounterd ./cmd/autocounterd

# Final image for autocounterd
FROM alpine AS autocounterd

COPY --from=builder /build/autocounterd /usr/bin/autocounterd

ENTRYPOINT [ "/usr/bin/autocounterd" ]
CMD [ "start" ]

# Final image for portalloopd
FROM docker AS portalloopd

RUN apk add bash curl jq

COPY --from=builder /build/portalloopd /usr/bin/portalloopd

ENTRYPOINT [ "/usr/bin/portalloopd" ]
CMD [ "serve" ]
