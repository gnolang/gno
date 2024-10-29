FROM golang:alpine AS builder

COPY . /go/src/github.com/gnolang/gno/misc/autocounterd

WORKDIR /go/src/github.com/gnolang/gno/misc/autocounterd

RUN go build -o /build/autocounterd ./cmd

# Final image for autocounterd
FROM alpine AS autocounterd

COPY --from=builder /build/autocounterd /usr/bin/autocounterd

ENTRYPOINT [ "/usr/bin/autocounterd" ]
CMD [ "start" ]

