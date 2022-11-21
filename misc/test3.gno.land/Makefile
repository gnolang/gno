all: build up logs

up: ./overlay/balances.txt ./overlay/genesis_txs.txt
	mkdir -p data/gnoland/config
	cp overlay/config.toml data/gnoland/config/config.toml
	docker compose up --build -d

logs:
	docker compose logs --tail=1000 -f

down:
	docker compose down
	docker volume rm -f test3gnoland_gnonode
	docker compose run gnoland rm -rf /opt/gno/src/testdir/data /opt/gno/src/testdir/config

pull:
	git pull

build:
	docker compose build

./overlay/balances.txt:
	wget https://github.com/gnolang/independence-day/raw/main/mkgenesis/balances.txt.gz -O overlay/balances.txt.gz
	gunzip overlay/balances.txt.gz

./overlay/genesis_txs.txt:
	wget https://github.com/gnolang/gno/raw/master/gnoland/genesis/genesis_txs.txt -O $@

gnotxport:
	#docker build -t gno ../..
	docker run -v "$(PWD):$(PWD)" -w "$(PWD)" gno gnotxport export --remote test3.gno.land:36657
	tail -n 1 txexport.log
	wc -l txexport.log

gnotximport:
	docker run -v "$(PWD):$(PWD)" -w "$(PWD)" gno gnotxport import --remote test3.gno.land:36657
