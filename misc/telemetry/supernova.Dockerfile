FROM golang:1.22-alpine

ARG supernova_version=latest

RUN go install github.com/gnolang/supernova/cmd@$supernova_version && mv /go/bin/cmd /go/bin/supernova
RUN export SUPERNOVA_PATH=$(go list -m -f "{{.Dir}}" github.com/gnolang/supernova@${supernova_version}) && \
    mkdir -p /supernova && \
    cp -r $SUPERNOVA_PATH/* /supernova

WORKDIR /supernova

ENTRYPOINT ["supernova"]
