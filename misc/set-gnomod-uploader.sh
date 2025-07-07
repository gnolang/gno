#!/bin/sh
# Set uploader address in gnomod.toml files
# Usage: ./set-gnomod-uploader.sh <path> <address>
# Example: ./set-gnomod-uploader.sh examples/gno.land/p/demo g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

[ $# -ne 2 ] && echo "Usage: $0 <path> <address>" && exit 1

find "$1" -name "gnomod.toml" 2>/dev/null | while read -r file; do
    if ! grep -q "upload_metadata" "$file"; then
        printf "\n[upload_metadata]\nuploader = \"%s\"\n" "$2" >> "$file"
        echo "Updated: $file"
    fi
done