.PHONY: help deploy admin mappings wmb tx
all: help deploy admin mappings wmb tx

deploy:
	@echo "deploying.."
	@zsh scripts/deploy.zsh
	@echo "done"

admin:
	@echo "querying admin.."
	@zsh scripts/admin.zsh
	@echo "done"

mappings:
	@echo "querying mappings.."
	@zsh scripts/mappings.zsh
	@echo "done"

wmb:
	@echo "querying waymobetta.."
	@zsh scripts/wmb.zsh
	@echo "done"

tx:
	@echo "executing transaction.."
	@zsh scripts/tx.zsh
	@echo "done"
