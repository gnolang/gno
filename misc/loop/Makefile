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

all: local

gnoland.start:
	cd $(gnoland_dir) && $(gnoland_bin) start -skip-failing-genesis-txs -genesis-txs-file $(HISTORY_OUTPUT)
gnoland.clean:
	make -C $(gnoland_dir) fclean
.PHONY: gnoland.start gnoland.clean

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
	@if [ -e $(BACKUP_FILE) ]; then \
  		touch $(HISTORY_OUTPUT); \
		cat $(HISTORY_OUTPUT) $(BACKUP_FILE) > $(HISTORY_OUTPUT).tmp; \
		mv $(HISTORY_OUTPUT).tmp $(HISTORY_OUTPUT); \
	else \
		echo "No existing backup file not found: $(BACKUP_FILE)"; \
	fi
.PHONY: save.history

local: gnoland.clean save.history
	./run_loop.sh
.PHONY: local
