DOCS_FOLDER := .

REGEX := 'http[s]?://[^)]+'

.PHONY: links
links:
	@echo "Checking for broken links in Markdown files..."
	@find $(DOCS_FOLDER) -type f -name '*.md' | while read file; do \
	    egrep -o $(REGEX) "$$file" | sort | uniq | grep -v -E 'localhost|127\.0\.0\.1' | while read url; do \
	        http_status=$$(curl -L -o /dev/null -s -w "%{http_code}" -m 10 "$$url"); \
	        curl_exit_code=$$?; \
	        if [ "$$curl_exit_code" -eq 3 ]; then \
	             printf "Skipping (\033[31m$$url\033[0m) in file: $$file\n"; \
	        elif [ "$$curl_exit_code" -ne 0 ]; then \
	            printf "Error retrieving ($$url) in file: $$file (Curl exit code: $$curl_exit_code)\n"; \
	        elif [ "$$http_status" -eq 404 ]; then \
	            printf "Broken link in file $$file: $$url\n"; \
	        fi \
	    done \
	done
