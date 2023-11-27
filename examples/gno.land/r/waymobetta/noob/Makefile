.PHONY: help deploy query tx
all: help deploy query tx

deploy:
	@echo "deploying.."
	@zsh scripts/deploy.zsh
	@echo "done"

query:
	@echo "querying waymobetta.."
	@zsh scripts/query.zsh
	@echo "done"

tx:
	@echo "submitting transaction"
	@zsh scripts/tx.zsh
	@echo "done"
