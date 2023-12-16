FROM ghcr.io/gnolang/tx-archive as archive-bin

FROM docker

RUN apk add bash curl jq

COPY --from=archive-bin /tx-archive /usr/local/bin/tx-archive

# Every day at 10am
RUN echo "* 10 * * * sh /scripts/switch.sh" > /var/spool/cron/crontabs/root
