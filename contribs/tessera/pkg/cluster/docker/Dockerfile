# Prepare the root gno dir in the image
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Copy the source files (from the tar) into the image
COPY . .

# Set up Go environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Build gnoland binary
RUN go build -o /gnoland ./gno.land/cmd/gnoland

# Final minimal image
FROM alpine:3.17

COPY --from=builder /gnoland /usr/bin/gnoland

EXPOSE 26656 26657

ENTRYPOINT ["/bin/sh", "-c", "sleep infinity"]