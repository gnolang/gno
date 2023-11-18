.PHONY: help deploy admin mappings wmb tx reg_user reg_addr
all: help deploy admin mappings wmb tx reg_user reg_addr

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

reg_user:
	@echo "registering username.."
	@zsh scripts/reg_user.zsh
	@echo "done"

reg_addr:
	@echo "registering address.."
	@zsh scripts/reg_addr.zsh
	@echo "done"
