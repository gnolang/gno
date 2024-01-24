FROM golang:alpine AS builder

COPY . /app

WORKDIR /app

RUN go build -o portalloopd ./cmd/portalloopd
RUN go build -o autocounterd ./cmd/autocounterd

# Final image for autocounterd
FROM alpine AS autocounterd

COPY --from=builder /app/autocounterd /usr/bin/autocounterd

ENTRYPOINT [ "/usr/bin/autocounterd" ]

# Final image for portalloopd
FROM docker AS portalloopd

RUN apk add bash curl jq

COPY --from=builder /app/portalloopd /usr/bin/portalloopd

ENTRYPOINT [ "/usr/bin/portalloopd" ]

CMD [ "serve" ]
