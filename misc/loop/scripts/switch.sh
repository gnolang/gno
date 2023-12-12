#!/usr/bin/env sh

set -e

# This script goal is to:
# 1. Snapshot txs for current running node
# 2. Start a new node with this tx in genesis file
# 3. switch traefik node


BACKUP_DIR="/backups"

NOW=$(date +'%Y-%m-%d_%s')
BACKUP_FILE="${BACKUP_DIR}/backup_${NOW}.jsonl"
BACKUP_LEGACY_FILE=$(echo ${BACKUP_FILE} | sed 's/.jsonl$/-legacy.jsonl/')

HOST_PWD="${HOST_PWD:=$(pwd)}"

TX_ARCHIVE_CMD=${TX_ARCHIVE_CMD:-"tx-archive"}

CONTAINER_NAME="gno-${NOW}"

# Get latest version of gno
docker pull ghcr.io/gnolang/gno || exit 0

# Set the current portal loop in READ-ONLY mode
sed -i -E 's/middlewares: \[.*\]/middlewares: ["ipwhitelist"]/' /etc/traefik/configs/gno.yml

# If there is no portal loop running, we start one
if docker ps --format json | jq '.Labels' | grep -q "the-portal-loop"; then
    ${TX_ARCHIVE_CMD} backup \
        --overwrite=true \
        --remote "rpc.gno.local:80" \
        --from-block 1 \
        --output-path="${BACKUP_FILE}"

    cat ${BACKUP_FILE} | jq -c -M '.tx' > ${BACKUP_LEGACY_FILE}
fi

docker volume create ${CONTAINER_NAME}

docker run -it \
    -d \
    --name "$CONTAINER_NAME" \
    -v ${HOST_PWD}/scripts:/scripts \
    -v ${HOST_PWD}/backups:/backups \
    -v ${CONTAINER_NAME}:/opt/gno/src/testdir \
    -p 26656 \
    -p 127.0.0.1::26657 \
    -e MONIKER="the-portal-loop" \
    -e GENESIS_BACKUP_FILE="${BACKUP_LEGACY_FILE}" \
    --label "the-portal-loop=${CONTAINER_NAME}" \
    --entrypoint /scripts/start.sh \
    ghcr.io/gnolang/gno

sleep 5

PORTS=$(docker inspect $CONTAINER_NAME | jq '.[0].NetworkSettings.Ports')

RPC_PORT=$(echo $PORTS | jq -r '."26657/tcp"[0].HostPort')

echo "New instance is running on: localhost:${RPC_PORT}"

# wait for RPC to be up
curl -s --retry 10 --retry-delay 5 --retry-all-errors -o /dev/null "localhost:${RPC_PORT}/status"


# Wait 5 blocks
while [ "$(curl -s localhost:${RPC_PORT}/status | jq -r '.result.sync_info.latest_block_height')" -le 5 ]
do
    sleep 1
done


# Update traefik url
sed -i -E "s#localhost:[0-9]+#localhost:${RPC_PORT}#"  /etc/traefik/configs/gno.yml

sed -i -E 's/middlewares: \[.*\]/middlewares: []/' /etc/traefik/configs/gno.yml

# Delete previous container
docker rm -f $(docker ps --format json --filter "label=the-portal-loop" | jq -r '.ID' | tail -n +2)

docker volume prune --all --force
