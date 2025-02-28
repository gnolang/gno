build: ../misc/docs/docusaurus/static/sidebar.json

../misc/docs/docusaurus/static/sidebar.json: README.md
	go build -C ../misc/docs/tools/indexparser -o ./build
	../misc/docs/tools/indexparser/build -path ../docs/README.md > ../misc/docs/docusaurus/static/sidebar.json

# Run the linter for the docs/ folder
lint:
	go build -C ../misc/docs/tools/linter -o ./build
	../misc/docs/tools/linter/build -path ../docs

dev:
	cd ../misc/docs/docusaurus/ && yarn start

clean:
	rm -f docusaurus/static/sidebar.json
