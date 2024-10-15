FROM golang:alpine AS builder

COPY . /go/src/github.com/gnolang/gno/misc/loop

WORKDIR /go/src/github.com/gnolang/gno/misc/loop

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go build -o /build/portalloopd ./cmd

# Final image for portalloopd
FROM docker AS portalloopd

RUN apk add bash curl jq

COPY --from=builder /build/portalloopd /usr/bin/portalloopd

ENTRYPOINT [ "/usr/bin/portalloopd" ]
CMD [ "serve" ]
