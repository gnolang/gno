build: ../misc/docs/docusaurus/static/sidebar.json

docusaurus/static/sidebar.json: README.md
	go run -C ../misc/docs/tools/tools/indexparser . -path ../../docs/README.md > docusaurus/static/sidebar.json

# Run the linter for the docs/ folder
lint:
	go run -C ../misc/docs/tools/linter . -path ../../docs/

dev:
	cd ../misc/docs/docusaurus/ && yarn start

clean:
	rm -f docusaurus/static/sidebar.json
