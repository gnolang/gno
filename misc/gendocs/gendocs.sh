#!/bin/bash
# Heavily modified version of the following script:
# https://gist.github.com/Kegsay/84ce060f237cb9ab4e0d2d321a91d920
set -u

DOC_DIR=godoc
PKG=github.com/gnolang/gno
# Used to load /static content
STATIC_PREFIX=/gno

# Run a pkgsite server which we will scrape. Use env to run it from our repo's root directory.
env -C ../.. pkgsite &
DOC_PID=$!

# Wait for the server to init
while :
do
    curl -s "http://localhost:8080" > /dev/null
    if [ $? -eq 0 ] # exit code is 0 if we connected
    then
        break
    fi
done

# Scrape the pkg directory for the API docs. Scrap lib for the CSS/JS. Ignore everything else.
wget \
	--verbose \
	--recursive \
	--mirror \
	--convert-links \
	--adjust-extension \
	--page-requisites \
	-erobots=off \
	--accept-regex='8080/((search|license-policy|about|)$|(static|images)/|github.com/gnolang/)' \
	http://localhost:8080/ \
	http://localhost:8080/static/frontend/frontend.js \
	http://localhost:8080/static/frontend/unit/unit.js \
	http://localhost:8080/static/frontend/unit/main/main.js \
	http://localhost:8080/third_party/dialog-polyfill/dialog-polyfill.js

# Stop the pkgsite server
kill -9 $DOC_PID

# Delete the old directory or else mv will put the localhost dir into
# the DOC_DIR if it already exists.
rm -rf $DOC_DIR
mv localhost\:8080 $DOC_DIR

# Perform various replacements to fix broken links/UI.
# /files/ will point to their github counterparts; we make links to importedby/version go nowhere;
# any other link will point to pkg.go.dev, and fix the /files/... text when viewing a pkg.
find godoc -type f -exec sed -ri 's#http://localhost:8080/files/[^"]*/github.com/gnolang/([^/"]+)/([^"]*)#https://github.com/gnolang/\1/blob/master/\2#g
s#http://localhost:8080/[^"?]*\?tab=(importedby|versions)#\##g
s#http://localhost:8080([^")]*)#https://pkg.go.dev\1#g
s#/files/[^" ]*/(github.com/[^" ]*)/#\1#g
s#s\.src = src;#s.src = "'"$STATIC_PREFIX"'" + src;#g' {} +

echo "Docs can be found in $DOC_DIR"
