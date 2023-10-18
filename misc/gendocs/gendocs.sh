#!/bin/bash
set -u

DOC_DIR=godoc
PKG=github.com/gnolang/gno

# Run a godoc server which we will scrape. Clobber the GOPATH to include
# only our dependencies.
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
wget -v -r -m -k -E -p -erobots=off --accept-regex='8080/((search|license-policy|about|)$|(static|images)/|github.com/gnolang/)' "http://localhost:8080/"

# Stop the godoc server
kill -9 $DOC_PID

# Delete the old directory or else mv will put the localhost dir into
# the DOC_DIR if it already exists.
rm -rf $DOC_DIR
mv localhost\:8080 $DOC_DIR

find godoc -type f -exec sed -ri 's#http://localhost:8080/files/[^"]*/github.com/gnolang/([^/"]+)/([^"]*)#https://github.com/gnolang/\1/blob/master/\2#g
s#http://localhost:8080/[^"?]*\?tab=(importedby|versions)#\##g
s#http://localhost:8080([^")]*)#https://pkg.go.dev\1#g
s#/files/[^" ]*/(github.com/[^" ]*)/#\1#g' {} +

echo "Docs can be found in $DOC_DIR"
