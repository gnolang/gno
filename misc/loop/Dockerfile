FROM golang:alpine AS builder

COPY . /app

WORKDIR /app

RUN go build -o portalloopd ./cmd

FROM docker

RUN apk add bash curl jq

COPY --from=builder /app/portalloopd /usr/bin/portalloopd

ENTRYPOINT [ "/usr/bin/portalloopd" ]

CMD [ "serve" ]
