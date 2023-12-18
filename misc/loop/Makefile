# The startup delay (waits until the node is "ready")
DELAY ?= 10 # seconds
# The temporary backup file for transactions
BACKUP_FILE ?= $(abspath ./txs_backup.log)
# The entire txs history across all iterations
HISTORY_OUTPUT := $(abspath ./txs_history.log)

# The gnoland binary
gnoland_bin := go run github.com/gnolang/gno/gno.land/cmd/gnoland
# The tx archive binary
tx_bin := go run github.com/gnolang/tx-archive/cmd

# The relative gno.land directory
gnoland_dir := $(abspath ../../gno.land)

all: loop

start.gnoland:
	cd $(gnoland_dir) && $(gnoland_bin) start -skip-failing-genesis-txs -genesis-txs-file $(HISTORY_OUTPUT)
clean.gnoland:
	make -C $(gnoland_dir) fclean
.PHONY: start.gnoland clean.gnoland

# Starts the backup service
# and backs up transactions into a file
# that is wiped on every loop
tx.backup:
	sleep $(DELAY)
	$(tx_bin) backup -legacy -watch -overwrite -output-path "$(BACKUP_FILE)"
.PHONY: tx.backup

# Saves the history from previous iterations into
# a temporary transactions log
save.history:
	@test -e $(BACKUP_FILE) || (echo "No existing backup file not found: '$(BACKUP_FILE)'"; exit 1)
	cat $(BACKUP_FILE) >> $(HISTORY_OUTPUT)
.PHONY: save.history

loop: clean.gnoland
	# backup history, if needed
	$(MAKE) save.history || true
	# run our dev loop
	./run_loop.sh
.PHONY: loop
