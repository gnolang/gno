# Prepare the root gno dir in the image
FROM golang:1.23-alpine AS build-gnoland

WORKDIR /src

# Copy the source files (from the tar) into the image
COPY . .

# Set up Go environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Build gnoland binary
RUN go build -o /gnoland ./gno.land/cmd/gnoland

# Prepare the contribs dependencies
FROM build-gnoland AS build-contribs

WORKDIR /src/contribs/gnogenesis

RUN go mod download -x

RUN go build -o /gnogenesis .

# Final minimal image
FROM alpine:3.17

COPY --from=build-gnoland /gnoland /usr/bin/gnoland
COPY --from=build-contribs /gnogenesis /usr/bin/gnogenesis

EXPOSE 26656 26657

ENTRYPOINT ["/bin/sh", "-c", "sleep infinity"]