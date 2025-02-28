MISC_DOCS=../misc/docs

build: $(MISC_DOCS)/docusaurus/static/sidebar.json

$(MISC_DOCS)/docusaurus/static/sidebar.json: README.md
	go build -C $(MISC_DOCS)/tools/indexparser -o ./build
	$(MISC_DOCS)/tools/indexparser/build -path ../docs/README.md > $(MISC_DOCS)/docusaurus/static/sidebar.json

# Run the linter for the docs/ folder
lint:
	go build -C $(MISC_DOCS)/tools/linter -o ./build
	$(MISC_DOCS)/tools/linter/build -path ../docs

dev:
	cd $(MISC_DOCS)/docusaurus/ && yarn start

clean:
	rm -f $(MISC_DOCS)/docusaurus/static/sidebar.json
