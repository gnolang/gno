.PHONY: all local

all: local

local:
	echo "Giving permissions to run script"
	chmod u+x ./dev.sh
	echo "Running local development setup"
	@./dev.sh $$RESTORE_DIR $$BACKUP_DIR
