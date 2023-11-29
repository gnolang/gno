.PHONY: deploy owner query tx
all: deploy owner query tx

deploy:
	@echo "deploying.."
	@zsh scripts/deploy.zsh
	@echo "done"

owner:
	@echo "querying owner"
	@zsh scripts/owner.zsh
	@echo "done"

query:
	@echo "querying waymobetta.."
	@zsh scripts/query.zsh
	@echo "done"

tx:
	@echo "submitting transaction"
	@zsh scripts/tx.zsh
	@echo "done"
	@make query
