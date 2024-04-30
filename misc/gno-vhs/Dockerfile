FROM ghcr.io/charmbracelet/vhs:v0.7.2

VOLUME /gobin

ENV PATH=/gobin:$PATH

RUN apt-get update && apt-get install -y make && rm -rf /var/lib/apt/lists/*
