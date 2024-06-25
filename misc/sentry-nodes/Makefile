
GNOLAND_GENESIS_DIR="../../../gno.land/genesis/"

all: genesis-txs genesis-balances

genesis.json:
	@cp -v genesis.validator.json genesis.json

genesis-balances: genesis.json
	gnoland genesis balances add --balance-sheet $(GNOLAND_GENESIS_DIR)/genesis_balances.txt

genesis-txs: genesis.json
	gnoland genesis txs add $(GNOLAND_GENESIS_DIR)/genesis_txs.jsonl

clean:
	@rm -rf genesis.json

re: clean all
