# Controllable signer image for misc/val-scenarios.
#
# Build context must be the repository root:
#   docker build -f misc/val-scenarios/Dockerfile --target valsignerd .

FROM golang:1.24-alpine AS build

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /gnoroot

RUN apk add --no-cache git
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/,id=valsignerd-gomodcache \
    --mount=type=cache,target=/root/.cache/go-build,id=valsignerd-gobuildcache \
    go mod download -x

COPY . ./

RUN { git describe --tags --exact-match 2>/dev/null \
    || echo "$(git rev-parse --abbrev-ref HEAD).$(git rev-list --count HEAD)+$(git rev-parse --short HEAD)"; \
    } > /gnoroot/build_version

RUN --mount=type=cache,target=/go/pkg/mod/,id=valsignerd-gomodcache \
    --mount=type=cache,target=/root/.cache/go-build,id=valsignerd-gobuildcache \
    go build -ldflags "-w -s -X github.com/gnolang/gno/tm2/pkg/version.Version=$(cat /gnoroot/build_version)" -o /out/valsignerd ./misc/val-scenarios/cmd/valsignerd

FROM alpine:3 AS valsignerd

RUN apk add --no-cache ca-certificates

COPY --from=build /out/valsignerd /usr/bin/valsignerd

EXPOSE 8080 26659
ENTRYPOINT ["/usr/bin/valsignerd"]
