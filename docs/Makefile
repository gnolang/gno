MISC_DOCS=../misc/docs

build: $(MISC_DOCS)/docusaurus/static/sidebar.json gen

$(MISC_DOCS)/docusaurus/static/sidebar.json: README.md
	go build -C $(MISC_DOCS)/tools/indexparser -o ./build
	$(MISC_DOCS)/tools/indexparser/build -path ../docs/README.md > $(MISC_DOCS)/docusaurus/static/sidebar.json

# Run the link linter for the docs/ folder
lint:
	go build -C $(MISC_DOCS)/tools/linter -o ./build
	$(MISC_DOCS)/tools/linter/build -path ../docs

# Embeds up-to-date assets into the docs
gen:
	go run -modfile ../misc/devdeps/go.mod github.com/campoy/embedmd -w `find . -name "*.md"`

# Starts local deployment of docusaurus
dev:
	cd $(MISC_DOCS)/docusaurus/ && yarn start

clean:
	rm -f $(MISC_DOCS)/docusaurus/static/sidebar.json
