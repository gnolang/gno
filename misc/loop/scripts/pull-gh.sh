#!/usr/bin/env sh

TMP_DIR=temp-tx-exports

# The master backup file will contain the ultimate txs backup
# that the portal loop use when looping (generating the genesis)
MASTER_BACKUP_FILE="backup.jsonl"

# Clones the portal loop backups subdirectory, located in BACKUPS_REPO (tx-exports)
pullGHBackups () {
  BACKUPS_REPO=https://github.com/gnolang/tx-exports.git
  BACKUPS_REPO_PATH="portal-loop"

  # Clone just the root folder of the same name
  git clone --depth 1 --no-checkout $BACKUPS_REPO
  cd "$(basename "$BACKUPS_REPO" .git)" || exit 1

  # Clone just the backups path in the cloned repo
  git sparse-checkout set $BACKUPS_REPO_PATH
  git checkout

  # Go back to the parent directory
  cd ..
}

# Create the temporary working dir
rm -rf $TMP_DIR && mkdir $TMP_DIR
cd $TMP_DIR || exit 1

# Pull the backup repo data
pullGHBackups

# Combine the pulled backups into a single backup file
TXS_BACKUPS_PREFIX="backup_portal_loop_txs_"

find . -type f -name "${TXS_BACKUPS_PREFIX}*.jsonl" | sort | xargs cat > "temp_$MASTER_BACKUP_FILE"

BACKUPS_DIR="../backups"
TIMESTAMP=$(date +%s)

# Check if the master backup file already exists
if [ -e "$BACKUPS_DIR/$MASTER_BACKUP_FILE" ]; then
  # Back up the existing master txs file
  echo "Master backup file exists, backing up..."
  mv "$BACKUPS_DIR/$MASTER_BACKUP_FILE" "$BACKUPS_DIR/${MASTER_BACKUP_FILE}-legacy-$TIMESTAMP"

  echo "Renamed $MASTER_BACKUP_FILE to ${MASTER_BACKUP_FILE}-legacy-$TIMESTAMP"
fi

# Use the GitHub state as the canonical backup
mv "temp_$MASTER_BACKUP_FILE" "$BACKUPS_DIR/$MASTER_BACKUP_FILE"
echo "Moved temp_$MASTER_BACKUP_FILE to $BACKUPS_DIR/$MASTER_BACKUP_FILE"

# Clean up the temporary directory
cd ..
rm -rf $TMP_DIR
