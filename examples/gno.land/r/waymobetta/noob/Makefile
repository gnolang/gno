.PHONY: help deploy query tx owner
all: help deploy query tx owner

deploy:
	@echo "deploying.."
	@zsh scripts/deploy.zsh
	@echo "done"

query: owner
	@echo "querying waymobetta.."
	@zsh scripts/query.zsh
	@echo "done"

tx:
	@echo "submitting transaction"
	@zsh scripts/tx.zsh
	@echo "done"

owner:
	@echo "querying owner"
	@zsh scripts/owner.zsh
	@echo "done"
